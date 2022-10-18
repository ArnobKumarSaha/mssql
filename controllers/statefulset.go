/*
Copyright 2022 Appscode Inc..

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
	"github.com/fatih/structs"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	cu "kmodules.xyz/client-go/client"
	coreutil "kmodules.xyz/client-go/core/v1"
	metautil "kmodules.xyz/client-go/meta"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	msapi "kubedb.dev/mssql/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type workloadOptions struct {
	// App level options
	stsName   string
	labels    map[string]string
	selectors map[string]string

	// db container options
	// cmd, args, envList & volumeMount of the main(`mssql`) container
	cmd         []string
	args        []string
	envList     []core.EnvVar
	volumeMount []core.VolumeMount

	// pod Template level options
	replicas       *int32                          // sts.Spec.Replicas
	gvrSvcName     string                          // sts.Spec.ServiceName
	podTemplate    *ofst.PodTemplateSpec           //
	pvcSpec        *core.PersistentVolumeClaimSpec // sts.Spec.VolumeClaimTemplates if storageType == Ephemeral
	emptyDirSpec   *core.EmptyDirVolumeSource      // sts.Spec.Template.Spec.Volumes if storageType != Ephemeral
	initContainers []core.Container                // sts.Spec.Template.Spec.InitContainers
	volumes        []core.Volume                   // sts.Spec.Template.Spec.Volumes
}

func (r *MSSQLReconciler) ensureStatefulSet(opts workloadOptions) (*apps.StatefulSet, kutil.VerbType, error) {
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := r.checkStatefulSet(opts.stsName); err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	stsMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: r.db.Namespace,
	}
	statefulSet, vt, err := cu.CreateOrPatch(r.ctx, r.Client, &apps.StatefulSet{
		ObjectMeta: stsMeta,
	}, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*apps.StatefulSet)
		in.Labels = r.db.PodControllerLabels(pt.Controller.Labels, opts.labels)
		coreutil.EnsureOwnerReference(&in.ObjectMeta, r.getOwnerRef())
		in.Spec.Replicas = opts.replicas
		in.Spec.ServiceName = opts.gvrSvcName
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: opts.selectors,
		}
		in.Spec.Template.Labels = r.db.PodLabels(pt.Labels, opts.labels)

		// init containers
		in.Spec.Template.Spec.InitContainers = coreutil.UpsertContainers(in.Spec.Template.Spec.InitContainers, pt.Spec.InitContainers)
		in.Spec.Template.Spec.InitContainers = coreutil.UpsertContainers(in.Spec.Template.Spec.InitContainers, opts.initContainers)

		// containers
		in.Spec.Template.Spec.Containers = coreutil.UpsertContainer(in.Spec.Template.Spec.Containers, getMainContainer(opts, r.db.Spec.Version))

		// volumes
		in.Spec.Template.Spec.Volumes = coreutil.UpsertVolume(in.Spec.Template.Spec.Volumes, opts.volumes...)
		copyFromPodTemplate(in, pt)
		return in
	})
	if err != nil {
		klog.Info("+++++++++++++++++  ensure statefull set ", err)
		return nil, kutil.VerbUnchanged, err
	}
	return statefulSet.(*apps.StatefulSet), vt, err
}

func copyFromPodTemplate(in *apps.StatefulSet, pt ofst.PodTemplateSpec) {
	in.Annotations = pt.Controller.Annotations
	in.Spec.Template.Annotations = pt.Annotations
	in.Spec.Template.Spec.NodeSelector = pt.Spec.NodeSelector
	in.Spec.Template.Spec.Affinity = pt.Spec.Affinity
	if pt.Spec.SchedulerName != "" {
		in.Spec.Template.Spec.SchedulerName = pt.Spec.SchedulerName
	}
	in.Spec.Template.Spec.Tolerations = pt.Spec.Tolerations
	in.Spec.Template.Spec.ImagePullSecrets = pt.Spec.ImagePullSecrets
	in.Spec.Template.Spec.PriorityClassName = pt.Spec.PriorityClassName
	in.Spec.Template.Spec.Priority = pt.Spec.Priority
	in.Spec.Template.Spec.HostNetwork = pt.Spec.HostNetwork
	in.Spec.Template.Spec.HostPID = pt.Spec.HostPID
	in.Spec.Template.Spec.HostIPC = pt.Spec.HostIPC
	if pt.Spec.SecurityContext == nil {
		pt.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	in.Spec.Template.Spec.SecurityContext = pt.Spec.SecurityContext

	in.Spec.Template.Spec.ServiceAccountName = pt.Spec.ServiceAccountName
	in.Spec.UpdateStrategy = apps.StatefulSetUpdateStrategy{
		Type: apps.OnDeleteStatefulSetStrategyType,
	}
}

// if a statefulSet is already there that doesn't contain the required labels -> return err. otherwise nil
func (r *MSSQLReconciler) checkStatefulSet(stsName string) error {
	var sts apps.StatefulSet
	err := r.Client.Get(r.ctx, types.NamespacedName{
		Name:      stsName,
		Namespace: r.db.Namespace,
	}, &sts)
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if sts.Labels[metautil.NameLabelKey] != r.db.ResourceFQN() ||
		sts.Labels[metautil.InstanceLabelKey] != r.db.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, r.db.Namespace, stsName)
	}

	return nil
}

// make a containerSpec using opts & podTemplate
func getMainContainer(opts workloadOptions, dbImage string) core.Container {
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	readinessProbe := pt.Spec.ReadinessProbe
	if readinessProbe != nil && structs.IsZero(*readinessProbe) {
		readinessProbe = nil
	}
	livenessProbe := pt.Spec.LivenessProbe
	if livenessProbe != nil && structs.IsZero(*livenessProbe) {
		livenessProbe = nil
	}

	return core.Container{
		Name:            msapi.MSSQLContainerName,
		Image:           dbImage,
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         opts.cmd,
		Args:            metautil.UpsertArgumentList(opts.args, pt.Spec.Args),
		Ports: []core.ContainerPort{
			{
				Name:          msapi.MSSQLDatabasePortName,
				ContainerPort: msapi.MSSQLDatabasePort,
				Protocol:      core.ProtocolTCP,
			},
		},
		Env:             coreutil.UpsertEnvVars(opts.envList, pt.Spec.Env...),
		Resources:       pt.Spec.Resources,
		SecurityContext: pt.Spec.ContainerSecurityContext,
		Lifecycle:       pt.Spec.Lifecycle,
		LivenessProbe:   livenessProbe,
		ReadinessProbe:  readinessProbe,
		VolumeMounts:    opts.volumeMount,
	}
}
