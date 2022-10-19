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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
)

const (
	ResourceCodeMSSQL     = "ms"
	ResourceKindMSSQL     = "MSSQL"
	ResourceSingularMSSQL = "mssql"
	ResourcePluralMSSQL   = "mssqls"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mssqls,singular=mssql,shortName=ms,categories={datastore,kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// MSSQL is the Schema for the mssqls API
type MSSQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MSSQLSpec   `json:"spec,omitempty"`
	Status MSSQLStatus `json:"status,omitempty"`
}

type MSSQLSpec struct {
	// Version of MSSQL to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a MSSQL database.
	Replicas *int32 `json:"replicas,omitempty"`

	// https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-editions-and-components-2019?view=sql-server-ver16#-editions
	// +kubebuilder:default="Developer"
	// +optional
	Edition MSSQLEdition `json:"edition"`

	// StorageType can be durable (default) or ephemeral
	StorageType dbapi.StorageType `json:"storageType,omitempty"`

	// Storage spec to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// EphemeralStorage spec to specify the configuration of ephemeral storage type.
	EphemeralStorage *core.EmptyDirVolumeSource `json:"ephemeralStorage,omitempty"`

	// SSLMode for both standalone and clusters. (default, disabled.)
	SSLMode string `json:"sslMode,omitempty"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// Database authentication secret
	// +optional
	AuthSecret *dbapi.SecretReference `json:"authSecret,omitempty"`

	// ConfigSecret is an optional field to provide custom configuration file for database
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []dbapi.NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 10, timeoutSeconds: 10, failureThreshold: 1}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`
}

// +kubebuilder:validation:Enum=Developer;Express;Standard;Enterprise
type MSSQLEdition string

const (
	MSSQLEditionDeveloper  MSSQLEdition = "Developer"
	MSSQLEditionExpress    MSSQLEdition = "Express"
	MSSQLEditionStandard   MSSQLEdition = "Standard"
	MSSQLEditionEnterprise MSSQLEdition = "Enterprise"
)

type MSSQLStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase string `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true

// MSSQLList contains a list of MSSQL
type MSSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MSSQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MSSQL{}, &MSSQLList{})
}
