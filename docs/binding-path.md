# Custom Containers Path and Secret Path

## Containers Path

If your application is using a custom resource and containers path should bind
at a certain location, SBO provides an API to achieve that.  Here is an example
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
    envVarPrefix: qiye111
    applicationSelector:
        resourceRef: example-appconfig
        group: stable.example.com
        version: v1
        resource: appconfigs
        bindingPath:
            containersPath: spec.containers
    backingServiceSelectors:
      - group: postgresql.baiju.dev
        version: v1alpha1
        kind: Database
        resourceRef: example-db
        id: zzz
        envVarPrefix: qiye
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
certain location, SBO provides an API to achieve that.  Here is an example CR
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
    envVarPrefix: qiye111
    applicationSelector:
        resourceRef: example-appconfig
        group: stable.example.com
        version: v1
        resource: appconfigs
        bindingPath:
            secretPath: spec.secret
    backingServiceSelectors:
      - group: postgresql.baiju.dev
        version: v1alpha1
        kind: Database
        resourceRef: example-db
        id: zzz
        envVarPrefix: qiye
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
