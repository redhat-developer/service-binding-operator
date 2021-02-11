# Application Workload Author's Guide



- [Introduction to Service Binding](#Introduction-to-Service-binding)

- [Backing Service providing binding metadata](#Backing-service-providing-binding-metadata)

- [Backing Service not providing binding metadata](#Backing-service-not-providing-binding-metadata)
  * [Annotate Service Resources](#Annotate-service-resources)
  * [Detect Binding Resources](#Detect-binding-resources)
  * [Compose custom binding variables](#Compose-custom-binding-variables)

- [Configuring custom naming strategy for binding names](#Custom-Naming-Strategy)
- [Accessing the binding data from the application](#Accessing-the-binding-data-from-the-application)
- [Binding non podSpec based application workloads](#Binding-non-podSpec-based-application-workloads)

- [Custom Containers Path and Secret Path](#Custom-Containers-Path-and-Secret-Path)
  * [Containers Path](#Containers-Path)
  * [Secret Path](#Secret-Path)

- [Binding Metadata in Annotations](#Binding-Metadata-in-Annotations)
  * [Requirements for specifying binding information in a backing service CRD / Kubernetes resource](#Requirements-for-specifying-binding-information-in-a-backing-service-CRD-/-Kubernetes-resource)
  * [Data model : Building blocks for expressing binding information](#Data-model-:-Building-blocks-for-expressing-binding-information)
  * [A Sample CR : The Kubernetes resource that the application would bind to](#A-Sample-CR-:-The-Kubernetes-resource-that-the-application-would-bind-to)
  * [Scenarios](#Scenarios)

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

As shown above, you may also directly use a `ConfigMap` or a `Secret` itself as a service resource that would be used as a source of binding information.

# Backing Service providing binding metadata

If the backing service author has provided binding metadata in the corresponding CRD,
then Service Binding acknowledges it and automatically creates a binding secret with
the information relevant for binding.

The backing service may provide binding information as
* Metadata in the CRD as annotations
* Metadata in the OLM bunde manifest file as Descriptors
* Secret or ConfigMap

If the backing service provides binding metadata, you may use the resource as is
to express an intent to bind your workload with one or more service resources, by creating a `ServiceBinding`.

# Backing Service not providing binding metadata

If the backing service hasn't provided any binding metadata, the application author may annotate the Kubernetes resource representing the backing service such that the managed binding secret generated has the necessary binding information.

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
* A specific attribute in a `ConfigMap` referenced in the Kubernetes resource.

As an example, if the Cockroachdb authors do not provide any binding metadata in the CRD, you, as an application author may annotate the CR/kubernetes resource that manages the backing service ( cockroach DB ).

The backing service could be represented as any one of the following:
* Custom Resources.
* Kubernetes Resources, such as `Ingress`, `ConfigMap` and `Secret`.
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

  mappings:
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

This would generate a binding secret and inject it into the workload as an environment variable or a file based on `bindAsFile` flag in the `ServiceBinding` resource.


Here's how the environment variables look like:

```
$ env | grep COCKROACHDB

COCKROACHDB_CLUSTERIP=172.10.2.3
COCKROACHDB_CONF_PORT=8090

```

Here's how the mount paths look like:

```
bindings
├── <Service-binding-name>
│   ├── COCKROACHDB_CLUSTERIP
│   ├── COCKROACHDB_CONF_PORT
```

Instead of `/bindings`, you can specify a custom binding root path by specifying the same in `spec.mountPath`, example,

``` yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: binding-request
  namespace: service-binding-demo
spec:
  bindAsFiles: true
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
  mounthPath: '/bindings/accounts-db' # User configurable binding root
```

Here's how the mount paths would look like, where applicable:

```
bindings
├── accounts-db
│   ├── COCKROACHDB_CLUSTERIP
│   ├── COCKROACHDB_CONF_PORT
```

Setting `spec.bindAsFiles` to `true` (default: `false`) enables injecting gathered bindings as files into the application/workload.

For determining the folder where bindings should be injected, we can specify the destination using `spec.mountPath` or we can use `SERVICE_BINDING_ROOT` environment variable. If both are set then the `SERVICE_BINDING_ROOT` environment variable takes the higher precedence.

The following table summarizes how the final bind path is computed:

| spec.mountPath  | SERVICE_BINDING_ROOT | Final Bind Path                      |
| --------------- | ---------------------| -------------------------------------|
| nil             | non-existent         | /bindings/ServiceBinding_Name        |
| nil             | /some/path/root      | /some/path/root/ServiceBinding_Name  |
| /home/foo       | non-existent         | /home/foo                            |
| /home/foo       | /some/path/root      | /some/path/root/ServiceBinding_Name  |

## Custom naming strategy
Binding names declared through annotations or CSV descriptors are processed before injected into the application according to the following strategy
 - names are upper-cased
 - service resource kind is upper-cased and prepended to the name

example:
```yaml
DATABASE_HOST: example.com
```
`DATABASE` is backend service `Kind` and `HOST` is the binding name.

With custom naming strategy/templates we can build custom binding names.

Naming strategy defines a way to prepare binding names through
ServiceBinding request.

1. We have a nodejs application which connects to database.
2. Application mostly requires host address and exposed port information.
3. Application access this information through binding names.


###How ?
Following fields are part of `ServiceBinding` request.
- Application
```yaml
  application:
    name: nodejs-app
    group: apps
    version: v1
    resource: deployments
```

- Backend/Database Service
```yaml
namingStrategy: 'POSTGRES_{{ .service.kind | upper }}_{{ .name | upper }}_ENV'
services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: db-demo
```

Considering following are the fields exposed by above service to use for binding
1. host
2. port

We have applied `POSTGRES_{{ .service.kind | upper }}_{{ .name | upper }}_ENV` naming strategy
1. `.name` refers to the binding name specified in the crd annotation or descriptor.
2. `.service` refer to the services in the `ServiceBinding` request.
3. `upper` is the string function used to postprocess the string while compiling the template string.
4. `POSTGRES` is the prefix used.
5. `ENV` is the suffix used.

Following would be list of binding names prepared by above `ServiceBinding`

```yaml
POSTGRES_DATABASE_HOST_ENV: example.com
POSTGRES_DATABASE_PORT_ENV: 8080
```

We can define how that key should be prepared defining string template in `namingStrategy`

#### Naming Strategies

There are few naming strategies predefine.
1. `none` - When this is applied, we get binding names in following form - `{{ .name }}`
```yaml
host: example.com
port: 8080
```


2. `uppercase` - This is by uppercase set when no `namingStrategy` is defined and `bindAsFiles` set to false - `{{ .service.kind | upper}}_{{ .name | upper }}`

```yaml
DATABASE_HOST: example.com
DATABASE_PORT: 8080
```

3. `lowercase` - This is by default set when `bindAsFiles` set to true -`{{ .name | lower }}`

```yaml
host: example.com
port: 8080
```

#### Predefined string post processing functions
1. `upper` - Capatalize all letters
2. `lower` - Lowercase all letters
3. `title` - Title case all letters.


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

A detailed documentation could be found [here](#Custom-Containers-Path-and-Secret-Path).

# Custom Containers Path and Secret Path

## Containers Path

If your application is using a custom resource and containers path should bind
at a custom location, SBO provides an API to achieve that.  Here is an example
CR with containers in a custom location:

```
apiVersion: "stable.example.com/v1"
kind: AppConfig
metadata:
    name: example-appconfig
spec:
    containers:
    - name: hello-world
      image: yusufkaratoprak/kubernetes-gosample:latest
      ports:
      - containerPort: 8090
```

In the above CR, the containers path is at `spec.containers`.  You can specify
this path in the `ServiceBindingRequest` config at
`spec.applicationSelector.bindingPath.containersPath`:

```
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
    name: binding-request
spec:
    namePrefix: qiye111
    applicationSelector:
        name: example-appconfig
        group: stable.example.com
        version: v1
        resource: appconfigs
        bindingPath:
            containersPath: spec.containers
    backingServiceSelectors:
      - group: postgresql.baiju.dev
        version: v1alpha1
        kind: Database
        name: example-db
        id: zzz
        namePrefix: qiye
```

After reconciliation, the `spec.containers` is going to be updated with
`envFrom` and `secretRef` like this:

```
apiVersion: stable.example.com/v1
kind: AppConfig
metadata:
    name: example-appconfig
spec:
  containers:
  - env:
    - name: ServiceBindingOperatorChangeTriggerEnvVar
      value: "31793"
    envFrom:
    - secretRef:
        name: binding-request
    image: yusufkaratoprak/kubernetes-gosample:latest
    name: hello-world
    ports:
    - containerPort: 8090
    resources: {}
```

## Secret Path

If your application is using a custom resource and secret path should bind at a
custom location, SBO provides an API to achieve that.  Here is an example CR
with secret in a custom location:

```
apiVersion: "stable.example.com/v1"
kind: AppConfig
metadata:
    name: example-appconfig
spec:
    secret: some-value
```

In the above CR, the secret path is at `spec.secret`.  You can specify
this path in the `ServiceBindingRequest` config at
`spec.applicationSelector.bindingPath.secretPath`:


```
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
    name: binding-request
spec:
    namePrefix: qiye111
    applicationSelector:
        name: example-appconfig
        group: stable.example.com
        version: v1
        resource: appconfigs
        bindingPath:
            secretPath: spec.secret
    backingServiceSelectors:
      - group: postgresql.baiju.dev
        version: v1alpha1
        kind: Database
        name: example-db
        id: zzz
        namePrefix: qiye
```

After reconciliation, the `spec.secret` is going to be updated with
`binding-request` as the value:

```
apiVersion: "stable.example.com/v1"
kind: AppConfig
metadata:
    name: example-appconfig
spec:
    secret: binding-request
```

# Binding Metadata in Annotations

 During a binding operation, annotations from relevant Kubernetes resources are extracted to gather information about what is interesting for binding. This information is eventually used to bind the application with the backing service by populating the binding Secret.

## Requirements for specifying binding information in a backing service CRD / Kubernetes resource

1. Extract a string from the Kubernetes resource.
2. Extract a string from the Kubernetes resource, and map it to custom name in the binding Secret.
3. Extract an entire configmap/Secret from the Kubernetes resource.
4. Extract a specific field from the configmap/Secret from the Kubernetes resource, and bind it as an environment variable.
5. Extract a specific field from the configmap/Secret from the Kubernetes resource and and bind it as a volume mount.
6. Extract a specific field from the configmap/Secret from the Kubernetes resource and map it to different name in the binding Secret.
7. Extract a “slice of maps” from the Kubernetes resource and generate multiple fields in the binding Secret.
8. Extract a "slice of strings" from a Kubernetes resource and indicate the content in a specific index in the slice to be relevant for binding.


## Data model : Building blocks for expressing binding information

* `path`: A template representation of the path to the element in the Kubernetes resource. The value of `path` could be specified in either [JSONPath](https://kubernetes.io/docs/reference/kubectl/jsonpath/) or [GO templates](https://golang.org/pkg/text/template/)

* `elementType`: Specifies if the value of the element referenced in `path` is of type `string` / `sliceOfStrings` / `sliceOfMaps`. Defaults to `string` if omitted.

* `objectType`: Specifies if the value of the element indicated in `path` refers to a `ConfigMap`, `Secret` or a plain string in the current namespace!  Defaults to `Secret` if omitted and `elementType` is a non-`string`.

* `bindAs`: Specifies if the element is to be bound as an environment variable or a volume mount using the keywords `envVar` and `volume`, respectively. Defaults to `envVar` if omitted.

* `sourceKey`: Specifies the key in the configmap/Secret that is be added to the binding Secret. When used in conjunction with `elementType`=`sliceOfMaps`, `sourceKey` specifies the key in the slice of maps whose value would be used as a key in the binding Secret. This optional field is the operator author intends to express that only when a specific field in the referenced `Secret`/`ConfigMap` is bindable.

* `sourceValue`: Specifies the key in the slice of maps whose value would be used as the value, corresponding to the value of the `sourceKey` which is added as the key, in the binding Secret. Mandatory only if `elementType` is `sliceOfMaps`.


## A Sample CR : The Kubernetes resource that the application would bind to

```
    apiVersion: apps.kube.io/v1beta1
    kind: Database
    metadata:
      name: my-cluster
    spec:
    ...
    status:
      bootstrap:
        - type: plain
          url: myhost2.example.com
          name: hostGroup1
        - type: tls
          url: myhost1.example.com:9092
          name: hostGroup2
      data:
        dbConfiguration: database-config  # configmap
        dbCredentials: database-cred-Secret # Secret
        url: db.stage.ibm.com
```



## Scenarios


1. ### Use everything from the Secret  “status.data.dbCredentials”

    Requirement : *Extract an entire configmap/Secret from the Kubernetes resource*


    Annotation:

    ```
    “service.binding/dbcredentials”:”path={.status.data.dbcredentials},objectType=Secret”
    ```


    Descriptor:

    ```
    - path: data.dbcredentials
      x-descriptors:
        - urn:alm:descriptor:io.kubernetes:Secret
        - service.binding
    ```


2. ### Use everything from the ConfigMap “status.data.dbConfiguration”


    Requirement : *Extract an entire configmap/Secret from the Kubernetes resource*

    Annotation

    ```
    “service.binding/dbConfiguration”: "path={.status.data.dbConfiguration},objectType=ConfigMap”
    ```


    Descriptor

    ```
    - path: data.dbConfiguration
      x-descriptors:
        - urn:alm:descriptor:io.kubernetes:ConfigMap
        - service.binding
    ```

3. ### Use “certificate” from the ConfigMap “status.data.dbConfiguration” as an environment variable

    Requirement : *Extract a specific field from the configmap/Secret from the Kubernetes resource and use it as an environment variable.*


    Annotation

    ```
    “service.binding/certificate”:
    "path={.status.data.dbConfiguration},objectType=ConfigMap"
    ```


    Descriptor


    ```
    - path: data.dbConfiguration
      x-descriptors:
        - urn:alm:descriptor:io.kubernetes:ConfigMap
        - service.binding:certificate:bindAs=envVar
    ```


4. ### Use “certificate” from the ConfigMap “status.data.dbConfiguration” as a volume mount

    Requirement : *Extract a specific field from the configmap/Secret from the Kubernetes resource and use it as a volume mount.*


    Annotation

    ```
    “service.binding/certificate”:
    "path={.status.data.dbConfiguration},bindAs=volume,objectType=ConfigMap"
    ```


    Descriptor

    ```
    - path: data.dbConfiguration
      x-descriptors:
        - urn:alm:descriptor:io.kubernetes:ConfigMap
        - service.binding:certificate:bindAs=volume
    ```


5. ### Use “db_timeout” from the ConfigMap “status.data.dbConfiguration” as “timeout” in the binding Secret.

    Requirement: *Extract a specific field from the configmap/Secret from the Kubernetes resource and map it to different name in the binding Secret*

    Annotation

    ```
    “service.binding/timeout”:
    “path={.status.data.dbConfiguration},objectType=ConfigMap,sourceKey=db_timeout”
    ```


    Descriptor

    ```
    - path: data.dbConfiguration
      x-descriptors:
        - urn:alm:descriptor:io.kubernetes:ConfigMap
        - service.binding:timeout:sourceKey=db_timeout
    ```

6. ### Use the attribute “status.data.url”

    Requirement: *Extract a string from the Kubernetes resource.*

    Annotation

    ```
    “service.binding/url”:"path={.status.data.url}"
    ```

    Descriptor

    ```
    - path: data.url
      x-descriptors:
        - service.binding
    ```

7. ### Use the attribute “status.data.connectionURL” as uri in the binding Secret

    Requirement: *Extract a string from the Kubernetes resource, and map it to custom name in the binding Secret.*

    Annotation

    ```
    “service.binding/uri: "path={.status.data.connectionURL}”
    ```



    Descriptor

    ```
    - path: data.connectionURL
      x-descriptors:
        - service.binding:uri
    ```

8. ### Use specific elements from the CR’s “status.bootstrap” to produce key/value pairs in the  binding Secret

    Requirement: *Extract a “slice of maps” from the Kubernetes resource and generate multiple fields in the binding Secret.*

    Annotation

    ```
    “service.binding/endpoints”:
    "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url"
    ```


    Descriptor

    ```
    - path: bootstrap
      x-descriptors:
        - service.binding:endpoints:elementType=sliceOfMaps:sourceKey=type:sourceValue=url
    ```

9. ### Use Go template to produce key/value pairs in the binding Secret <kbd>EXPERIMENTAL</kbd>

    Requirement: *Extract binding information from the Kubernetes resource using Go templates and generate multiple fields in the binding Secret.*

    A sample Kafka CR:

    ```
    apiVersion: kafka.strimzi.io/v1alpha1
    kind: Kafka
    metadata:
      name: my-cluster
    ...
    status:
      listeners:
        - type: plain
          addresses:
            - host: my-cluster-kafka-bootstrap.service-binding-demo.svc
              port: 9092
            - host: my-cluster-kafka-bootstrap.service-binding-demo.svc
              port: 9093
        - type: tls
          addresses:
            - host: my-cluster-kafka-bootstrap.service-binding-demo.svc
              port: 9094
    ```

    Go Template:
    ```
    {{- range $idx1, $lis := .status.listeners -}}
      {{- range $idx2, $adr := $el1.addresses -}}
        {{ $lis.type }}_{{ $idx2 }}={{ printf "%s:%s\n" "$adr.host" "$adr.port" | b64enc | quote }}
      {{- end -}}
    {{- end -}}
    ```

    The above Go template produces the following string when executed on the sample Kafka CR:
    ```
    plain_0="<base64 encoding of my-cluster-kafka-bootstrap.service-binding-demo.svc:9092>"
    plain_1="<base64 encoding of my-cluster-kafka-bootstrap.service-binding-demo.svc:9093>"
    tls_0="<base64 encoding of my-cluster-kafka-bootstrap.service-binding-demo.svc:9094>"
    ```

    The string can then be parsed into key-value pairs to be added into the final binding secret. The Go template above can be written as one-liner and added as `{{GO TEMPLATE}}` in the annotation and descriptor below.

    Annotation

    ```
    “service.binding:
    "path={.status.listeners},elementType=template,source={{GO TEMPLATE}}"
    ```

    Descriptor

    ```
    - path: listeners
      x-descriptors:
        - service.binding:elementType=template:source={{GO TEMPLATE}}
    ```
