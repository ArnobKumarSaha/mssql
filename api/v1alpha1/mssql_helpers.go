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

package v1alpha1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metautil "kmodules.xyz/client-go/meta"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
)

func (in MSSQL) ResourceShortCode() string {
	return ResourceCodeMSSQL
}

func (in MSSQL) ResourceKind() string {
	return ResourceKindMSSQL
}

func (in MSSQL) ResourceSingular() string {
	return ResourceSingularMSSQL
}

func (in MSSQL) ResourcePlural() string {
	return ResourcePluralMSSQL
}

func (in MSSQL) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", in.ResourcePlural(), kubedb.GroupName)
}

func (in MSSQL) Owner() *metav1.OwnerReference {
	return metav1.NewControllerRef(&in, GroupVersion.WithKind(in.ResourceKind()))
}

func (in MSSQL) OffshootName() string {
	return in.Name
}

func (in MSSQL) PrimaryServiceName() string {
	return in.OffshootName()
}

func (in MSSQL) GoverningServiceName() string {
	return metautil.NameWithSuffix(in.PrimaryServiceName(), "pods")
}

func (in MSSQL) offshootLabels(selector, override map[string]string) map[string]string {
	selector[metautil.ComponentLabelKey] = dbapi.ComponentDatabase
	return metautil.FilterKeys(kubedb.GroupName, selector, metautil.OverwriteKeys(nil, in.Labels, override))
}

func (in MSSQL) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		metautil.NameLabelKey:      in.ResourceFQN(),
		metautil.InstanceLabelKey:  in.Name,
		metautil.ManagedByLabelKey: kubedb.GroupName,
	}
	return metautil.OverwriteKeys(selector, extraSelectors...)
}

func (in MSSQL) OffshootLabels() map[string]string {
	return in.offshootLabels(in.OffshootSelectors(), nil)
}

func (in MSSQL) ServiceLabels(alias dbapi.ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := dbapi.GetServiceTemplate(in.Spec.ServiceTemplates, alias)
	return in.offshootLabels(metautil.OverwriteKeys(in.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

func (in MSSQL) GetAuthSecretName() string {
	if in.Spec.AuthSecret != nil && in.Spec.AuthSecret.Name != "" {
		return in.Spec.AuthSecret.Name
	}
	return metautil.NameWithSuffix(in.OffshootName(), "auth")
}

func (in MSSQL) PodControllerLabels(podControllerLabels map[string]string, extraLabels ...map[string]string) map[string]string {
	return in.offshootLabels(metautil.OverwriteKeys(in.OffshootSelectors(), extraLabels...), podControllerLabels)
}

func (in MSSQL) PodLabels(podTemplateLabels map[string]string, extraLabels ...map[string]string) map[string]string {
	return in.offshootLabels(metautil.OverwriteKeys(in.OffshootSelectors(), extraLabels...), podTemplateLabels)
}
