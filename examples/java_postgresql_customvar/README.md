# Binding an Imported Java Spring Boot app to an In-cluster Operator Managed PostgreSQL Database

## Introduction

This scenario illustrates binding an imported Java application to an in-cluster operated managed PostgreSQL Database.

Note that this example app is configured to operate with OpenShift 4.5 or newer.

## Actions to Perform by Users in 2 Roles

In this example there are 2 roles:

* Cluster Admin - Installs the operators to the cluster
* Application Developer - Imports a Node.js application, creates a DB instance, creates a request to bind the application and DB (to connect the DB and the application).

### Cluster Admin

The cluster admin needs to install 2 operators into the cluster:

* Service Binding Operator
* Backing Service Operator

A Backing Service Operator that is "bind-able," in other
words a Backing Service Operator that exposes binding information in secrets, config maps, status, and/or spec
attributes. The Backing Service Operator may represent a database or other services required by
applications. We'll use [postgresql-operator](https://github.com/operator-backing-service-samples/postgresql-operator) to
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

### Update the Database CRD

Annotations are created to describe the values that should be collected and made available by the Service Binding's intermediary secret.


```shell
kubectl get crd databases.postgresql.baiju.dev -o yaml
```

The database CRD contains the list of annotations:

```yaml
metadata:
  annotations:
    service.binding/db.host: path={.status.dbConfigMap},objectType=ConfigMap
    service.binding/db.name: path={.status.dbConfigMap},objectType=ConfigMap
    service.binding/db.password: path={.status.dbConfigMap},objectType=ConfigMap
    service.binding/db.port: path={.status.dbConfigMap},objectType=ConfigMap
    service.binding/db.user: path={.status.dbConfigMap},objectType=ConfigMap
    service.binding/dbConnectionIP: path={.status.dbConnectionIP}
    service.binding/dbConnectionPort: path={.status.dbConnectionPort}
    service.binding/dbName: path={.spec.dbName}
    service.binding/password: path={.status.dbCredentials},objectType=Secret
    service.binding/user: path={.status.dbCredentials},objectType=Secret
```

We can add more annotations to collect values from different kinds of data structures.

To edit the CRD run the command:

```shell
kubectl edit crd databases.postgresql.baiju.dev
```

Add these annotations under `metadata.annotations` along with other annotations.

```yaml
service.binding/tags: path={.spec.tags},elementType=sliceOfStrings
service.binding/userLabels: path={.spec.userLabels},elementType=map
service.binding/secretName: path={.spec.secretName},elementType=sliceOfMaps,sourceKey=type,sourceValue=secret
```

These annotations refer to data which can be either a number, string, boolean, or an object or a slice of arbitrary values. In the case of this example,
- `service.binding/tags` represents a sequence
- `service.binding/userLabels` represents a mapping
- `service.binding/secretName` represents a sequence of mapping


### Application Developer

#### Create a namespace called `service-binding-demo`

The application and the DB needs a namespace to live in so let's create one for them:

```shell
kubectl create namespace service-binding-demo
```

#### Import an application

In this example we will import an arbitrary [Java Spring Boot application](https://github.com/ldimaggi/java-rest-http-crud).

In the OpenShift Console switch to the Developer perspective. (Make sure you have selected the `service-binding-demo` project). Navigate to the `+Add` page from the menu and then click on the `[From Git]` button. Fill in the form with the following:

* `Project` = `service-binding-demo`
* `Git Repo URL` = `https://github.com/ldimaggi/java-rest-http-crud`
* `Builder Image` = `Java`
* `Application Name` = `java-app`
* `Name` = `java-app`

* `Select the resource type to generate` = Deployment
* `Create a route to the application` = checked

and click on the `[Create]` button.

Notice, that during the import no DB config was mentioned or requested.

When the application is running navigate to its route to verify that it is up. Try the application's UI to add a fruit - it causes an error proving that the DB is not connected.

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
  tags:
    - "centos7-12.3"
    - "centos7-12.4"
  userLabels:
    archive: "false"
    environment: "demo"
  secretName:
    - type: "primarySecretName"
      secret: "example-primaryuser"
    - type: "secondarySecretName"
      secret: "example-secondaryuser"
    - type: "rootSecretName"
      secret: "example-rootuser"
EOD
```

#### Express an intent to bind the DB and the application

Now, the only thing that remains is to connect the DB and the application. We let the Service Binding Operator to 'magically' do the connection for us.

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
    name: java-app
    group: apps
    version: v1
    resource: deployments
  services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: db-demo
    id: postgresDB
  mappings:
  - name: JDBC_URL
    value: 'jdbc:postgresql://{{ .postgresDB.status.dbConnectionIP }}:{{ .postgresDB.status.dbConnectionPort }}/{{ .postgresDB.status.dbName }}'
  - name: DB_USER
    value: '{{ .postgresDB.status.dbCredentials.user }}'
  - name: DB_PASSWORD
    value: '{{ .postgresDB.status.dbCredentials.password }}'
EOD
```

There are 2 parts in the request:

* `application` - used to search for the application based on the name that we set earlier and the `group`, `version` and `resource` of the application to be a `Deployment` named `java-app`.
* `services` - used to find the backing service - our operator-backed DB instance called `db-demo`.
* `mappings` - used to create custom environment variables constructed using a templating engine from out-of-the-box bound information.

That causes the application to be re-deployed.
Once the new version is up, go to the application's route to check the UI. Now, it works!

### Check the status of Service Binding

`ServiceBinding Status` depicts the status of the Service Binding operator. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

To check the status of Service Binding, run the command:

```
kubectl get servicebinding binding-request -n service-binding-demo -o yaml
```

Status of Service Binding on successful binding:

```yaml
status:
  conditions:
  - lastHeartbeatTime: "2020-08-12T07:05:22Z"
    lastTransitionTime: "2020-08-12T06:39:13Z"
    status: "True"
    type: CollectionReady
  - lastHeartbeatTime: "2020-08-12T07:05:22Z"
    lastTransitionTime: "2020-08-12T06:39:13Z"
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



### Check secret injection into the application deployment

When the `ServiceBinding` was created the Service Binding Operator's controller injected the DB connection information into the application's `Deployment` as environment variables via an intermediate `Secret` called `binding-request`:

```shell
kubectl get deployment java-app -n service-binding-demo -o yaml
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

### Check the environment variable creation in the application pod

To list all the pods, run the command:

```
$ kubectl get pods -n service-binding-demo
NAME                                   READY   STATUS      RESTARTS   AGE
db-demo-postgresql-6574fc44bd-5qjct   1/1     Running     0          7m9s
java-app-1-build                      0/1     Completed   0          21m
java-app-67bdf56459-vfxqt             1/1     Running     0          6m46s
```

To ssh into the application pod `java-app-67bdf56459-vfxqt`, run command:

```
oc exec -it java-app-67bdf56459-vfxqt -- /bin/bash
```

Print all the environment variables that were injected into the application pod by Service Binding Operator:

```
[jboss@java-app-67bdf56459-vfxqt ~]$ printenv | grep DATABASE_ | sort
DATABASE_DBCONNECTIONIP=172.30.197.39
DATABASE_DBCONNECTIONPORT=5432
DATABASE_DBNAME=db-demo
DATABASE_DB_HOST=172.30.197.39
DATABASE_DB_NAME=db-demo
DATABASE_DB_PASSWORD=password
DATABASE_DB_PORT=5432
DATABASE_DB_USER=postgres
DATABASE_IMAGE=docker.io/postgres
DATABASE_IMAGENAME=postgres
DATABASE_PASSWORD=password
DATABASE_SECRETNAME_PRIMARYSECRETNAME=example-primaryuser
DATABASE_SECRETNAME_ROOTSECRETNAME=example-rootuser
DATABASE_SECRETNAME_SECONDARYSECRETNAME=example-secondaryuser
DATABASE_TAGS_0=centos7-12.3
DATABASE_TAGS_1=centos7-12.4
DATABASE_USER=postgres
DATABASE_USERLABELS_ARCHIVE=false
DATABASE_USERLABELS_ENVIRONMENT=demo
```

Notice, distinct environment variables are produced for sibling properties of the `tags`, `secretName` and `userLabels` in the variable name.
