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

const (
	MSSQLContainerName                  = "mssql"
	MSSQLWorkDirectoryName              = "workdir"
	MSSQLWorkDirectoryPath              = "/work-dir"
	MSSQLInstallContainerName           = "copy-config"
	MSSQLDatabasePortName               = "db"
	MSSQLDatabasePort                   = 1433
	MSSQLUser                           = "sa"
	MSSQLDataDirectoryName              = "datadir"
	MSSQLDataDirectoryPath              = "/var/opt/mssql"
	MSSQLDefaultVolumeClaimTemplateName = MSSQLDataDirectoryName
)
