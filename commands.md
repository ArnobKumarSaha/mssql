
## Initializing commands : 

- `kubebuilder init --domain kubedb.com --repo kubedb.dev/mssql --owner "Appscode Inc."`
- `kubebuilder create api --group microsoft --version v1alpha1 --kind MSSQL`

NB : `domain` will be concatenated with the group name. `repo` will be set as the module name on go.mod file.
And `owner` will be set on the license's boilerplate.

## Testing commands: 
- make `manifests` `install` `run` to run locally.
- `make docker-build docker-push IMG=arnobkumarsaha/mssql:dev` + `make deploy IMG=arnobkumarsaha/mssql:dev`
to run in `In-CLuster` method.


## TODO :
- deploy with webhook enabled.