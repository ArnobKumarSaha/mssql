
## Initializing commands : 

- `kubebuilder init --domain kubedb.com --repo kubedb.dev/mssql --owner "Appscode Inc."`
- `kubebuilder create api --group microsoft --version v1alpha1 --kind MSSQL`

## Testing commands: 
- make `manifests` `install` `run` to run locally.
- `make docker-build docker-push IMG=arnobkumarsaha/mssql:dev` + `make deploy IMG=arnobkumarsaha/mssql:dev`
to run in `In-CLuster` method.


## TODO :
- deploy with webhook enabled.