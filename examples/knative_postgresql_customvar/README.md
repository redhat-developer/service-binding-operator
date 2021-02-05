# Binding an Imported Quarkus app deployed as Knative service with an In-cluster Operator Managed PostgreSQL Database

## Introduction

This scenario illustrates binding a Quarkus application deployed as Knative service with an in-cluster operated managed PostgreSQL Database.

## Actions to Perform by Users in 2 Roles

In this example there are 2 roles:

* Cluster Admin - Installs the operators and Serverless plugin to the OpenShift cluster
* Application Developer - Imports a Node.js application, creates a DB instance, creates a request to bind the application and DB (to connect the DB and the application).

### Cluster Admin

The cluster admin needs to install operators, knative serving and a builder image into the cluster:

* Service Binding Operator
* Backing Service Operator
* Serverless Operator
  * Serverless Operator
  * Serverless UI

A Backing Service Operator that is "bind-able," in other
words a Backing Service Operator that exposes binding information in secrets,
config maps, status, and/or spec attributes. The Backing Service Operator
may represent a database or other services required by applications. We'll
use [postgresql-operator](https://github.com/operator-backing-service-samples/postgresql-operator) to
demonstrate a sample use case.

#### Install the Service Binding Operator

Navigate to the `Operators`->`OperatorHub` in the OpenShift console and in the `Developer Tools` category select the `Service Binding Operator` operator

![Service Binding Operator as shown in OperatorHub](../../assets/operator-hub-sbo-screenshot.png)

and install the `beta` version.

This makes the `ServiceBinding` custom resource available, that the application developer will use later.

#### Install the DB operator using a `CatalogSource`

Apply the following `CatalogSource`:

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

Then navigate to the `Operators`->`OperatorHub` in the OpenShift console and in the `Database` category select the `PostgreSQL Database` operator

![PostgreSQL Database Operator as shown in OperatorHub](../../assets/operator-hub-pgo-screenshot.png)

and install a `beta` version.

This makes the `Database` custom resource available, that the application developer will use later.

#### Install the OpenShift Serverless Operator

Navigate to the `Operators`->`OperatorHub` in the OpenShift console and in the `Cloud Provider` category select the `OpenShift Serverless Operator` operator.

##### Install Serverless UI

```shell
kubectl apply -f - << EOD
---
apiVersion: v1
kind: Namespace
metadata:
  name: knative-serving
---
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeServing
metadata:
  name: knative-serving
  namespace: knative-serving
EOD
```

This enables `Serverless` view in the UI. Note that installing the Serverless features will require around seven minutes for the command to run. You can check the status of this install with this command:

```shell
kubectl get knativeserving knative-serving -n knative-serving -o yaml
```

When everything is installed, let's refresh the OpenShift Console Page to make the Serverless features visible.

### Application Developer

#### Create a namespace called `service-binding-demo`

The application and the DB needs a namespace to live in so let's create one for them:

```shell
kubectl create namespace service-binding-demo
```

#### Create a DB instance for the application

Now we utilize the DB operator that the cluster admin has installed. To create a DB instance just create a `Database` custom resource in the `service-binding-demo` namespace called `db-demo`:

```shell
kubectl apply -f - << EOD
---
apiVersion: postgresql.baiju.dev/v1alpha1
kind: Database
metadata:
  name: db-demo
  namespace: service-binding-demo
spec:
  image: docker.io/postgres
  imageName: postgres
  dbName: db-demo
EOD
```

#### Import an application

In this example we will import an arbitrary [Quarkus application](https://github.com/sbose78/using-spring-data-jpa-quarkus).

In the OpenShift Console switch to the Developer perspective. (We need to make sure we've selected the `service-binding-demo` project). Navigate to the `+Add` page from the menu and then click on the `[Container Image]` button. Fill in the form with the following:

* `Image name from external registry` = `quay.io/pmacik/using-spring-data-jqa-quarkus:latest`
* (Optional) `Runtime Icon` = `quarkus`
* `Application Name` = `knative-app`
* `Name` = `knative-app`
* `Select the resource type to generate`->`Knative Service` = checked
* `Advanced Options`->`Create a route to the application` = checked

and click on the `[Create]` button.

Notice, that during the import no DB config was mentioned or requested.

After the application is imported we can check the `Services` under `Serverless` view to see the deployed application. The application should fail at this point with `Reason` to be the "connection refused" error. That indicates that the application is not connected to the DB.

#### Express an intent to bind the DB with the application

Now, the only thing that remains is to connect the DB and the application. We let the Service Binding Operator to make the connection for us.

Create the following `ServiceBinding`:

```shell
kubectl apply -f - << EOD
---
apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: binding-request
  namespace: service-binding-demo
spec:
  application:
    group: serving.knative.dev
    version: v1beta1
    resource: services
    name: knative-app
  services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: db-demo
    id: knav
  mappings:
    - name: JDBC_URL
      value: jdbc:postgresql://{{ .knav.status.dbConnectionIP }}:{{ .knav.status.dbConnectionPort }}/{{ .knav.status.dbName }}
    - name: DB_USER
      value: "{{ .knav.status.dbCredentials.user }}"
    - name: DB_PASSWORD
      value: "{{ .knav.status.dbCredentials.password }}"
EOD
```

There are 3 parts in the request:

* `application` - used to search for the application based on the name that we set earlier and the `group`, `version` and `resource` of the application to be a knative `Service`.
* `services` - used to find the backing service - our operator-backed DB instance called `db-demo`.
* `mappings` - used to create custom environment variables and files constructed using a templating engine from out-of-the-box bound information.

That causes the application to be re-deployed.

Once the new version is up, go to the application's route to check the UI. Now, it works!

When the `ServiceBinding` was created the Service Binding Operator's controller injected the DB connection information into the
application as environment variables via an intermediate `Secret` called `binding-request`:

```shell
kubectl get servicebinding binding-request -o yaml
```
```yaml
spec:
  template:
    spec:
      containers:
        - envFrom:
          - secretRef:
              name: binding-request
```

#### Check the status of Service Binding

`ServiceBinding Status` depicts the status of the Service Binding operator. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

To check the status of Service Binding, run the command:

```
kubectl get servicebinding binding-request -n service-binding-demo -o yaml
```

Status of Service Binding on successful binding:

```yaml
status:
  conditions:
  - lastHeartbeatTime: "2020-10-15T13:23:36Z"
    lastTransitionTime: "2020-10-15T13:23:23Z"
    status: "True"
    type: CollectionReady
  - lastHeartbeatTime: "2020-10-15T13:23:36Z"
    lastTransitionTime: "2020-10-15T13:23:23Z"
    status: "True"
    type: InjectionReady
  secret: binding-request
```

where

* Conditions represent the latest available observations of Service Binding's state
* Secret represents the name of the secret created by the Service Binding Operator


Conditions have two types `CollectionReady` and `InjectionReady`

where

* `CollectionReady` type represents collection of secret from the service
* `InjectionReady` type represents an injection of the secret into the application

Conditions can have the following type, status and reason:

| Type            | Status | Reason               | Type           | Status | Reason                   |
| --------------- | ------ | -------------------- | -------------- | ------ | ------------------------ |
| CollectionReady | False  | EmptyServiceSelector | InjectionReady | False  |                          |
| CollectionReady | False  | ServiceNotFound      | InjectionReady | False  |                          |
| CollectionReady | True   |                      | InjectionReady | False  | EmptyApplicationSelector |
| CollectionReady | True   |                      | InjectionReady | False  | ApplicationNotFound      |
| CollectionReady | True   |                      | InjectionReady | True   |                          |

That's it, folks!
