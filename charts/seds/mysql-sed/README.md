This helm chart defines a MySQL Service Endpoint Definition (SED). When the SED is installed it will provide the user with the oportunity to provide connection information as well as credentials to authenticate. The following are the values that can be customized when the SED chart is installed:

1. Host
1. Port
1. Username
1. Password
1. Databasename

The SED Chart will render a secret with the connection information. This secret is compliant with the Service Binding Specification [Well Known Secret Entries](https://github.com/servicebinding/spec#well-known-secret-entries). Therefore, the secret rendered by MySQL SED Chart is a bindable service endpoint that can be projected to workloads using the Service [Binding Direct Secret Reference](https://github.com/servicebinding/spec#well-known-secret-entries).