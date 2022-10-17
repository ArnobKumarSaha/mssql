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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	cu "kmodules.xyz/client-go/client"
	coreutil "kmodules.xyz/client-go/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	msapi "kubedb.dev/mssql/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MSSQLReconciler) ensurePrimaryService() error {
	svcTemplate := dbapi.GetServiceTemplate(r.db.Spec.ServiceTemplates, dbapi.PrimaryServiceAlias)
	svcMeta := metav1.ObjectMeta{
		Name:      r.db.PrimaryServiceName(),
		Namespace: r.db.Namespace,
	}

	_, _, err := cu.CreateOrPatch(r.ctx, r.Client, &core.Service{
		ObjectMeta: svcMeta,
	}, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*core.Service)
		coreutil.EnsureOwnerReference(&in.ObjectMeta, r.getOwnerRef())
		in.Labels = r.db.ServiceLabels(dbapi.PrimaryServiceAlias, svcTemplate.Labels)
		in.Annotations = svcTemplate.Annotations

		in.Spec.Selector = r.db.OffshootSelectors()
		if r.db.Spec.Replicas != nil && *r.db.Spec.Replicas > 1 {
			in.Spec.Selector[dbapi.LabelRole] = dbapi.DatabasePodPrimary
		}
		in.Spec.Ports = coreutil.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       msapi.MSSQLDatabasePortName,
				Port:       msapi.MSSQLDatabasePort,
				TargetPort: intstr.FromString(msapi.MSSQLDatabasePortName),
			},
		})
		copyFromServiceTemplateSpec(in, svcTemplate.Spec)
		return in
	})
	return err
}

func copyFromServiceTemplateSpec(in *core.Service, svcSpec ofst.ServiceSpec) {
	in.Spec.Ports = ofst.PatchServicePorts(in.Spec.Ports, svcSpec.Ports)
	if svcSpec.ClusterIP != "" {
		in.Spec.ClusterIP = svcSpec.ClusterIP
	}
	if svcSpec.Type != "" {
		in.Spec.Type = svcSpec.Type
	}
	in.Spec.ExternalIPs = svcSpec.ExternalIPs
	in.Spec.LoadBalancerIP = svcSpec.LoadBalancerIP
	in.Spec.LoadBalancerSourceRanges = svcSpec.LoadBalancerSourceRanges
	in.Spec.ExternalTrafficPolicy = svcSpec.ExternalTrafficPolicy
	if svcSpec.HealthCheckNodePort > 0 {
		in.Spec.HealthCheckNodePort = svcSpec.HealthCheckNodePort
	}
}

func (r *MSSQLReconciler) ensureGoverningServices() error {
	svcFunc := func(svcName string, labels, selectors map[string]string) error {
		svcMeta := metav1.ObjectMeta{
			Name:      svcName,
			Namespace: r.db.Namespace,
		}
		_, vt, err := cu.CreateOrPatch(r.ctx, r.Client, &core.Service{
			ObjectMeta: svcMeta,
		}, func(obj client.Object, createOp bool) client.Object {
			in := obj.(*core.Service)
			coreutil.EnsureOwnerReference(&in.ObjectMeta, r.getOwnerRef())
			in.Labels = labels
			in.Spec.Selector = selectors

			in.Spec.Type = core.ServiceTypeClusterIP
			in.Spec.ClusterIP = core.ClusterIPNone // headless service
			in.Spec.PublishNotReadyAddresses = true
			in.Spec.Ports = coreutil.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
				{
					Name:       msapi.MSSQLDatabasePortName,
					Port:       msapi.MSSQLDatabasePort,
					TargetPort: intstr.FromString(msapi.MSSQLDatabasePortName),
				},
			})

			return in
		})

		if err == nil && vt != kutil.VerbUnchanged {
		}
		return err
	}

	err := svcFunc(r.db.GoverningServiceName(),
		r.db.OffshootLabels(),
		r.db.OffshootSelectors(),
	)
	if err != nil {
		return err
	}
	return nil
}
