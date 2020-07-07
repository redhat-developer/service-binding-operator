# Testing PodSpec path

Based on the note by @qibobo 

Create the application CRD:

```
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
```

Create the application CR using the CRD:

```
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
```

Create the operator source:

```
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: db-operators
  namespace: openshift-marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: pmacik
```

Install the Postgres Operator and then create Database CR:

```
apiVersion: postgresql.baiju.dev/v1alpha1
kind: Database
metadata:
  name: db-demo
spec:
  image: docker.io/postgres
  imageName: postgres
  dbName: db-demo
```

Create the SBR with a custom bind path:

```
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
  name: binding-request-sample
spec:
  envVarPrefix: qiye111
  applicationSelector:
    resourceRef: demo-appconfig
    group: stable.example.com
    version: v1
    resource: appconfigs 
    bindingPath:
      podSpecPath:
        containers: spec.spec.containers
        volumes: spec.spec.volumes
  backingServiceSelectors:
    group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    resourceRef: db-demo
    id: zzz
    envVarPrefix: qiye
```

Check the secret `binding-request-sample` has been injected:

```
kubectl get appconfigs.stable.example.com -o yaml
apiVersion: v1
items:
- apiVersion: stable.example.com/v1
  kind: AppConfig
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"stable.example.com/v1","kind":"AppConfig","metadata":{"annotations":{},"name":"demo-appconfig","namespace":"default"},"spec":{"Command":"some command","image":"my-image","spec":{"containers":[{"image":"yusufkaratoprak/kubernetes-gosample:latest","name":"hello-world","ports":[{"containerPort":8090}]}]},"uri":"some uri"}}
    creationTimestamp: "2020-04-30T01:07:21Z"
    generation: 12
    name: demo-appconfig
    namespace: default
    resourceVersion: "4909355"
    selfLink: /apis/stable.example.com/v1/namespaces/default/appconfigs/demo-appconfig
    uid: e0175a4c-a442-4b7c-a2c7-de132eb4dd7a
  spec:
    Command: some command
    image: my-image
    spec:
      containers:
      - env:
        - name: ServiceBindingOperatorChangeTriggerEnvVar
          value: "2020-05-22T08:08:51Z"
        envFrom:
        - secretRef:
            name: binding-request-sample
        image: yusufkaratoprak/kubernetes-gosample:latest
        name: hello-world
        ports:
        - containerPort: 8090
        resources: {}
    uri: some uri
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```

Delete the SBR and secret has been removed.

```
spec:
      containers:
      - env:
        - name: ServiceBindingOperatorChangeTriggerEnvVar
          value: "2020-05-22T08:08:51Z"
        image: yusufkaratoprak/kubernetes-gosample:latest
        name: hello-world
        ports:
        - containerPort: 8090
        resources: {}
```
