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
	coreutil "kmodules.xyz/client-go/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	msapi "kubedb.dev/mssql/api/v1alpha1"
)

func (r *MSSQLReconciler) ensureNodes() error {
	args, err := r.getArgs()
	if err != nil {
		return err
	}
	envList := r.getEnvList()

	podTemplate := r.db.Spec.PodTemplate
	initContnr, initvolumes, err := r.installInitContainer(podTemplate, r.db.Spec.ConfigSecret)
	if err != nil {
		return err
	}

	opts := workloadOptions{
		stsName:   r.db.OffshootName(),
		labels:    r.db.OffshootLabels(),
		selectors: r.db.OffshootSelectors(),
		args:      args,
		cmd: func() []string {
			return []string{}
		}(),
		envList:        envList,
		initContainers: []core.Container{*initContnr},
		gvrSvcName:     r.db.GoverningServiceName(),
		podTemplate:    podTemplate,
		pvcSpec:        r.db.Spec.Storage,
		emptyDirSpec:   r.db.Spec.EphemeralStorage,
		replicas:       r.db.Spec.Replicas,
		volumes:        r.getVolumes(initvolumes, podTemplate),
		volumeMount:    r.getVolumeMounts(podTemplate),
	}

	_, _, err = r.ensureStatefulSet(opts)
	if err != nil {
	}
	return err
}

func (r *MSSQLReconciler) getArgs() ([]string, error) {
	var args []string
	return args, nil
}

func (r *MSSQLReconciler) getEnvList() []core.EnvVar {
	return []core.EnvVar{
		{
			Name: "POD_NAME",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &core.EnvVarSource{
				FieldRef: &core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name:  "AUTH",
			Value: "true",
		},
		{
			Name: "MSSQL_SA_USERNAME",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: r.db.Spec.AuthSecret.Name,
					},
					Key: core.BasicAuthUsernameKey,
				},
			},
		},
		{
			Name: "MSSQL_SA_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: r.db.Spec.AuthSecret.Name,
					},
					Key: core.BasicAuthPasswordKey,
				},
			},
		},
		{
			Name:  "MSSQL_PID",
			Value: string(r.db.Spec.Edition),
		},
		{
			Name:  "ACCEPT_EULA",
			Value: "Y",
		},
	}
}

func getCommonVolumesAndMounts() ([]core.Volume, []core.VolumeMount) {
	return nil, nil
}

func (r *MSSQLReconciler) installInitContainer(
	podTemplate *ofst.PodTemplateSpec,
	configSecret *core.LocalObjectReference,
) (*core.Container, []core.Volume, error) {
	var pt ofst.PodTemplateSpec
	if podTemplate != nil {
		pt = *podTemplate
	}
	initVolumes, mounts := getCommonVolumesAndMounts()

	if configSecret != nil {
	}

	return &core.Container{
		Name:            msapi.MSSQLInstallContainerName,
		Image:           "busybox",
		ImagePullPolicy: core.PullIfNotPresent,
		Command:         []string{"/bin/sh"},
		Env: func() []core.EnvVar {
			return []core.EnvVar{}
		}(),
		Args:         []string{},
		VolumeMounts: mounts,
		Resources:    pt.Spec.Resources,
	}, initVolumes, nil
}

func (r *MSSQLReconciler) getVolumeMounts(podTemplate *ofst.PodTemplateSpec) []core.VolumeMount {
	mounts := []core.VolumeMount{
		{
			Name:      msapi.MSSQLWorkDirectoryName,
			MountPath: msapi.MSSQLWorkDirectoryPath,
		},
	}
	return upsertCustomVolumeMounts(mounts, podTemplate)
}

func upsertCustomVolumeMounts(mounts []core.VolumeMount, podTemplate *ofst.PodTemplateSpec) []core.VolumeMount {
	var pt ofst.PodTemplateSpec
	if podTemplate != nil {
		pt = *podTemplate
	}
	mounts = coreutil.UpsertVolumeMount(pt.Spec.VolumeMounts, mounts...)
	return mounts
}

func (r *MSSQLReconciler) getVolumes(initVolumes []core.Volume, podTemplate *ofst.PodTemplateSpec) []core.Volume {
	var volumes []core.Volume
	volumes = coreutil.UpsertVolume(volumes, initVolumes...)

	volumes = coreutil.UpsertVolume(volumes, core.Volume{
		Name: msapi.MSSQLWorkDirectoryName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	})
	return upsertCustomVolumes(volumes, podTemplate)
}

func upsertCustomVolumes(volumes []core.Volume, podTemplate *ofst.PodTemplateSpec) []core.Volume {
	var pt ofst.PodTemplateSpec
	if podTemplate != nil {
		pt = *podTemplate
	}
	volumes = coreutil.UpsertVolume(pt.Spec.Volumes, volumes...)
	return volumes
}
