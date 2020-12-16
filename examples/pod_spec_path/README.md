# Testing PodSpec path

Based on the note by @qibobo

Assuming Service Binding Operator is installed from OperatorHub via `beta` channel.

Create the application CRD:

```shell
kubectl apply -f - << EOD
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: appconfigs.stable.example.com
spec:
  group: stable.example.com
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: appconfigs
    singular: appconfig
    kind: AppConfig
    shortNames:
    - ac
EOD
```

Create the application CR using the CRD:

```shell
kubectl apply -f - << EOD
---
apiVersion: "stable.example.com/v1"
kind: AppConfig
metadata:
  name: demo-appconfig
spec:
  uri: "some uri"
  Command: "some command"
  image: my-image
  spec:
    containers:
    - name: hello-world
      # Image from dockerhub, This is the import path for the Go binary to build and run.
      image: yusufkaratoprak/kubernetes-gosample:latest
      ports:
      - containerPort: 8090
EOD
```

Create the DB operator's catalog source:

```shell
kubectl apply -f - << EOD
---
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
    name: sample-db-operators
    namespace: openshift-marketplace
spec:
    sourceType: grpc
    image: quay.io/redhat-developer/sample-db-operators-olm:v1
    displayName: Sample DB Operators
EOD
```

Install the Postgres Operator in OperatorHub via the `beta` channel and then create Database CR:

```shell
kubectl apply -f - << EOD
---
apiVersion: postgresql.baiju.dev/v1alpha1
kind: Database
metadata:
  name: db-demo
spec:
  image: docker.io/postgres
  imageName: postgres
  dbName: db-demo
EOD
```

Create the Service Binding with a custom bind path:

```shell
kubectl apply -f - << EOD
---
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
    name: binding-request-sample
spec:
    namePrefix: qiye111
    application:
        name: demo-appconfig
        group: stable.example.com
        version: v1
        resource: appconfigs
        bindingPath:
            containersPath: spec.spec.containers
    services:
      - group: postgresql.baiju.dev
        version: v1alpha1
        kind: Database
        name: db-demo
        id: zzz
        namePrefix: qiye
EOD
```

Check the secret `binding-request-sample` has been injected:

```shell
kubectl get appconfigs.stable.example.com demo-appconfig -o yaml
```
```yaml
apiVersion: stable.example.com/v1
kind: AppConfig
metadata:
  ...
  name: demo-appconfig
  namespace: default
  ...
spec:
  Command: some command
  image: my-image
  spec:
    containers:
    - env:
      - name: ServiceBindingOperatorChangeTriggerEnvVar
        value: "106757"
      envFrom:
      - secretRef:
          name: binding-request-sample
      image: yusufkaratoprak/kubernetes-gosample:latest
      name: hello-world
      ports:
      - containerPort: 8090
      resources: {}
  uri: some uri
```
