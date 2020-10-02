# Application Workload Author's Guide



- [Introduction to Service Binding](#Introduction-to-Service-binding)

- [Backing Service providing binding metadata](#Backing-service-providing-binding-metadata)

- [Backing Service not providing binding metadata](#Backing-service-not-providing-binding-metadata)
  * [Annotate Service Resources](##Annotate-service-resources)
  * [Detect Binding Resources](#Detect-binding-resources)
  * [Compose custom binding variables](#Compose-custom-binding-variables)

- [Accessing the binding data from the application](#Accessing-the-binding-data-from-the-application)

- [Binding non-podSpec-based application workloads]([#Binding-non-podSpec-based-application-workloads])

<!-- toc -->


# Introduction to Service Binding

A Service Binding involves connecting an application to one or more backing services using a binding secret generated for the purpose of storing information to be consumed by the application.

``` yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: binding-request
  namespace: service-binding-demo

spec:
  ## Workload where binding information is
  ## injected.

  application:
    name: java-app
    group: apps
    version: v1
    resource: deployments

  ## One or more Service resources from which
  ## binding information is collected
  services:
  - group: database.example.com
    version: v1alpha1
    kind: DBInstance
    name: db

  - group: database.example.com
    version: v1alpha1
    kind: DBCredentials
    name: db

  - group: "route.openshift.io"
    version: v1
    kind: Route
    name: auth-service
```

The application workload expects binding metadata to be present on the Kubernetes Resources representing the backing service.

As shown above, you may also directly use a `Configmap` or a `Secret` itsef as a service resource that would be used as a source of binding information.

# Backing Service providing binding metadata

If the backing service author has provided binding metadata in the corresponding CRD,
then Service Binding acknowleges it and automatically creates a binding secret with
the information relevant for binding.

The backing service may provide binding information as
* Metadata in the CRD as annotations
* Metadata in the OLM bunde manifest file as Descriptiors
* Secret or ConfigMap

If the backing service provides binding metadata, you may use the resource as is
to express an intent to bind your workload with one or more service resources, by creating a `ServiceBinding`.

# Backing Service not providing binding metadata

If the backing service hasn't provided any binding metadata, the application author may annotate the Kubernetes resource representating the backing service such that the managed binding secret generated has with the necessary binding information.

As an application author, you have a couple of options to extract binding information 
from the backing service: 

* Decorate the backing service resource using annotations.
* Define custom binding variables.

In the following section, details of the above methods to make 
a backing service consumable for your application workload, is explained.

## Annotate Service Resources
---

The application author may consider specific elements of the backing service resource interesting for binding

* A specific attribute in the `spec` section of the Kubernetes resource.
* A specific attribute in the `status` section of the Kubernetes resource.
* A specific attribute in the `data` section of the Kubernetes resource.
* A specific attribute in a `Secret` referenced in the Kubernetes resource.
* A specific attribute in a `Configmap` referenced in the Kubernetes resource.

As an example, if the Cockroachdb authors do not provide any binding metadata in the CRD, you, as an application author may annotate the CR/kubernetes resource that manages the backing service ( cockroach DB ).

Please refer to the [documentation](docs/roadmap.md) to annotate objects for binding metadata.

The backing service could be represented as any one of the following:
* Custom Resources.
* Kubernetes Resources, such as `Ingress`, `Configmap` and `Secret`.
* OpenShift Resources, such as `Routes`.

## Compose custom binding variables
---

If the backing service doesn't expose binding metadata or the values exposed are not easily consumable, then an application author may compose custom binding variables using attributes in the Kubernetes resource representing the backing service.


## Custom binding variables
---


The *custom binding variables* feature enables application authors to request customized binding secrets using a combination of Go and jsonpath templating.

Example, the backing service CR may expose the host, port and database user in separate variables, but the application may need to consume this information as a connection string.




``` yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: multi-service-binding
  namespace: service-binding-demo
spec:

  application:
    name: java-app
    group: apps
    version: v1
    resource: deployments

 services:
  - group: postgresql.baiju.dev 
    version: v1alpha1
    kind: Database
    name: db-demo   <--- Database service
    id: postgresDB <--- Optional "id" field
  - group: ibmcloud.ibm.com
    version: v1alpha1
    kind: Binding
    name: mytranslator-binding <--- Translation service
    id: translationService

  customEnvVar:
    ## From the database service
    - name: JDBC_URL
      value: 'jdbc:postgresql://{{ .postgresDB.status.dbConnectionIP }}:{{ .postgresDB.status.dbConnectionPort }}/{{ .postgresDB.status.dbName }}'
    - name: DB_USER
      value: '{{ .postgresDB.status.dbCredentials.user }}'

    ## From the translator service
    - name: LANGUAGE_TRANSLATOR_URL
      value: '{{ index translationService.status.secretName "url" }}'
    - name: LANGUAGE_TRANSLATOR_IAM_APIKEY
      value: '{{ index translationService.status.secretName "apikey" }}'

    ## From both the services!
    - name: EXAMPLE_VARIABLE
      value: '{{ .postgresDB.status.dbName }}{{ translationService.status.secretName}}'
   
    ## Generate JSON. 
    - name: DB_JSON
      value: {{ json .postgresDB.status }}
    
```

This has been adopted in [IBM CodeEngine](https://cloud.ibm.com/docs/codeengine?topic=codeengine-kn-service-binding).


*In future releases, the above would be supported as volume mounts too.*

## Detect Binding Resources
---

The Service Binding Operator binds all information 'dependent' to the backing service CR by populating the binding secret with information from Routes, Services, ConfigMaps, and Secrets owned by the backing service CR if you express an intent to extract the same in case the backing service isn't annotated with the binding metadata.

[This](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#owners-and-dependents) is how owner and dependent relationships are set in Kubernetes.

The binding is initiated by the setting this `detectBindingResources: true` in the `ServiceBinding` CR's `spec`.

``` yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: etcdbinding
  namespace: service-binding-demo
spec:
  detectBindingResources: true
  application:
    name: java-app
    group: apps
    version: v1
    resource: deployments
  services:
  - group: etcd.database.coreos.com
    version: v1beta2
    kind: EtcdCluster
    name: etcd-cluster-example

```

When this API option is set to true, the Service Binding Operator automatically detects Routes, Services, ConfigMaps, and Secrets owned by the backing service CR and generates a binding secret out of it.



## Accessing the binding data from the application
---

The binding secret generated for the `ServiceBinding` may be associated with the workload as environment variables or volume mounts ("files").


``` yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: binding-request
  namespace: service-binding-demo
spec:
  application:
    name: java-app
    group: apps
    version: v1
    resource: deployments
  services:
  - group: charts.helm.k8s.io
    version: v1alpha1
    kind: Cockroachdb
    name: db-demo
    id: db_1
```

The generated binding secret would look like this:

``` yaml
kind: Secret
apiVersion: v1
metadata:
  name: example-servicebindingrequest
  namespace: pgo
data:
  COCKROACHDB_CLUSTERIP: MTcyLjMwLjEwMS4zNA==
  COCKROACHDB_CONF_PORT: MjYyNTc=
type: Opaque
```

This would generate a binding secret and inject it into the workload as an environment variable or a volume mount in the path `/var/data` depending on the intent expressed in the annotations on the Custom Resource.


Here's how the environment variables look like:

```
$ env | grep COCKROACHDB

COCKROACHDB_CLUSTERIP=172.10.2.3
COCKROACHDB_CONF_PORT=8090

```

Here's how the mount paths look like:

```
var
├── data
│   ├── COCKROACHDB_CLUSTERIP
│   ├── COCKROACHDB_CONF_PORT

```

Instead of `/var/data`, you can specify a custom binding root path by specifying the same in `spec.mountPathPrefix`, example,

``` yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: binding-request
  namespace: service-binding-demo
spec:
  application:
    name: java-app
    group: apps
    version: v1
    resource: deployments
  services:
  - group: charts.helm.k8s.io
    version: v1alpha1
    kind: Cockroachdb
    name: db-demo
    id: db_1
  mounthPathPrefix: '/bindings/accounts-db' # User configurable binding root
```

Here's how the mount paths would look like, where applicable:

```
bindings
├── accounts-db
│   ├── COCKROACHDB_CLUSTERIP
│   ├── COCKROACHDB_CONF_PORT
```


**Note**

*Injection of binding information as volume mounts is in the development phase and is not stable enough for use.*



# Binding non-podSpec-based application workloads

If your application is to be deployed as a non-podSPec-based workload such that the containers path should bind at a custom location, the `ServiceBinding` API provides an API to achieve that. 


``` yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: binding-request
  namespace: service-binding-demo
spec:
    application:
        resourceRef: example-appconfig
        group: stable.example.com
        version: v1
        resource: appconfigs
        bindingPath:
            secretPath: spec.secret # custom path to secret reference
            containersPath: spec.containers # custom path to containers reference
    ...
    ...
```

A detailed documentation could be found [here](docs/binding-path.md).

