# Binding an Imported app to a Route/Ingress

## Introduction

Binding information can be present in standalone k8s objects like routes/ingress, services, deployments too. This scenario illustrates using any resource ( CR / non-CR ) which has a spec and a status as a backing service.

Binding metadata is being read from annotations on the backing service ( like CR, Route, Service, basically any kubernetes object with a spec and status, along with associated CRD or CSV.

Here's how the operator resolves the binding metadata:

1) Look up annotations in the CR or kubernetes resource,
2) Look up annotations in CRD
3) Look up descriptors in CSV ( overrides the CRD annotations ..)
Provide cumulative annotations : (1) and (2 & 3).


## Actions to Perform by Users in 2 Roles

In this example there are 2 roles:

* Cluster Admin - Installs the operators to the cluster
* Application Developer - Imports the application, creates an OpenShift Route and creates a request to bind the application and the route.

### Cluster Admin

The cluster admin needs to install Service Binding Operator into the cluster.

Navigate to the `Operators`->`OperatorHub` in the OpenShift console and in the `Developer Tools` category select the `Service Binding Operator` operator

![Service Binding Operator as shown in OperatorHub](../../assets/operator-hub-sbo-screenshot.png)

and install the `beta` version.

This makes the `ServiceBinding` custom resource available, that the application developer will use later.

### Application Developer

#### Create a namespace called `service-binding-demo`

The application and the DB needs a namespace to live in so let's create one for them:

```shell
kubectl create namespace service-binding-demo
```

#### Import an application

In this example we will import an arbitrary [hello-openshift application](https://hub.docker.com/r/openshift/hello-openshift).

For that we just need to create a `Deployment` called `hello-app` from the `openshift/hello-openshift` container image.

``` shell
$ kubectl create deployment -n service-binding-demo --image=openshift/hello-openshift hello-app
deployment.apps/hello-app created
```

We can verify that the app is running by checking the Deployment is available:

``` shell
$ kubectl get deployment hello-app -n service-binding-demo -o wide                             
NAME        READY   UP-TO-DATE   AVAILABLE   AGE    CONTAINERS        IMAGES                      SELECTOR
hello-app   1/1     1            1           3m2s   hello-openshift   openshift/hello-openshift   app=hello-app
```
#### Create a Route and annotate it:

Now let's create a kubernetes resource - `Route` (for our case) and annotate it with the value that we would like to be injected for binding. For this case it is the `spec.host`

``` shell
kubectl apply -f - << EOD
---
kind: Route
apiVersion: route.openshift.io/v1
metadata:
  name: example
  namespace: service-binding-demo
  annotations:
    openshift.io/host.generated: 'true'
    service.binding/host: path={.spec.host} #annotate here.
spec:
  host: example-sbo.apps.ci-ln-smyggvb-d5d6b.origin-ci-int-aws.dev.rhcloud.com
  path: /
  to:
    kind: Service
    name: example
    weight: 100
  port:
    targetPort: 80
  wildcardPolicy: None
EOD
```

Now create a Service Binding as below:

``` shell
kubectl apply -f - << EOD
---
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding

metadata:
  name: binding-request
  namespace: service-binding-demo

spec:
  application:
    group: apps
    resource: deployments
    name: hello-app
    version: v1

  services:
    - group: route.openshift.io
      version: v1
      kind: Route # <--- not NECESSARILY a CR
      name: example
      namespace: service-binding-demo
EOD
```

When the `ServiceBinding` was created the Service Binding Operator's controller injected the Route information that was annotated to be injected into the application's `Deployment` as environment variables via an intermediate `Secret` called `binding-request`.

```shell
kubectl get sbr binding-request -n service-binding-demo -o yaml
```
```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  ...
  name: binding-request
  namespace: service-binding-demo
  ...
...
status:
  conditions:
  - lastHeartbeatTime: "2020-10-19T11:35:53Z"
    lastTransitionTime: "2020-10-19T11:35:37Z"
    status: "True"
    type: CollectionReady
  - lastHeartbeatTime: "2020-10-19T11:35:53Z"
    lastTransitionTime: "2020-10-19T11:35:37Z"
    status: "True"
    type: InjectionReady
  secret: binding-request
```

Check the contents of `Secret` - `binding-request` by executing:

```shell
kubectl get secrets binding-request -n service-binding-demo -o yaml
```

for the following result:

```yaml
apiVersion: v1
data:
  ROUTE_HOST: ZXhhbXBsZS1zYm8uYXBwcy5jaS1sbi1zbXlnZ3ZiLWQ1ZDZiLm9yaWdpbi1jaS1pbnQtYXdzLmRldi5yaGNsb3VkLmNvbQ==
kind: Secret
metadata:
  ...
  name: binding-request
  namespace: service-binding-demo
  ...
...
```

The secret value is actually encoded with base64 so to get the actual value we need to decode it properly:

```shell
kubectl get secret binding-request -n service-binding-demo -o jsonpath='{.data.ROUTE_HOST}' | base64 --decode
```
for the following result:
```
example-sbo.apps.ci-ln-smyggvb-d5d6b.origin-ci-int-aws.dev.rhcloud.com
```
