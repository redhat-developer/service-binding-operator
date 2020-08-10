# Binding an Imported Java Spring Boot app to an In-cluster Operator Managed PostgreSQL Database

## Introduction

This scenario illustrates binding an imported Java application to an in-cluster operated managed PostgreSQL Database.

Note that this example app is configured to operate with OpenShift 4.3 or newer.
To use this example app with OpenShift 4.2, replace references to resource:`Deployment`s with `DeploymentConfig`s and group:`apps` with `apps.openshift.io`.

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

and install the `alpha` version.

Alternatively, you can perform the same task manually using the following command:

``` shell
make install-service-binding-operator-community
```

This makes the `ServiceBindingRequest` custom resource available, that the application developer will use later.

##### :bulb: Latest `master` version of the operator

It is also possible to install the latest `master` version of the operator instead of the one from `community-operators`. To enable that an `OperatorSource` has to be installed with the latest `master` version:

``` shell
cat <<EOS | kubectl apply -f -
---
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: redhat-developer-operators
  namespace: openshift-marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: redhat-developer
EOS
```

Alternatively, you can perform the same task manually using the following command before going to the Operator Hub:

``` shell
make install-service-binding-operator-source-master
```

or running the following command to install the operator completely:

``` shell
make install-service-binding-operator-master
```

#### Install the DB operator using an `OperatorSource`

Apply the following `OperatorSource`:

```shell
cat <<EOS |kubectl apply -f -
---
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: db-operators
  namespace: openshift-marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: pmacik
EOS
```

Alternatively, you can perform the same task with this make command:

```shell
make install-backing-db-operator-source
```

Then navigate to the `Operators`->`OperatorHub` in the OpenShift console and in the `Database` category select the `PostgreSQL Database` operator

![PostgreSQL Database Operator as shown in OperatorHub](../../assets/operator-hub-pgo-screenshot.png)

and install a `stable` version.

This makes the `Database` custom resource available, that the application developer will use later.

### Update the Database CRD

Annotations are created to describe the values that should be collected and made available by the Service Binding Request's intermediary secret.
 

```
oc get crd databases.postgresql.baiju.dev -o yaml
```
 
The database CRD contains the list of annotations:

```
metadata:
  annotations:
    servicebindingoperator.redhat.io/spec.dbName: binding:env:attribute
    servicebindingoperator.redhat.io/status.dbConfigMap-db.host: binding:env:object:configmap
    servicebindingoperator.redhat.io/status.dbConfigMap-db.name: binding:env:object:configmap
    servicebindingoperator.redhat.io/status.dbConfigMap-db.password: binding:env:object:configmap
    servicebindingoperator.redhat.io/status.dbConfigMap-db.port: binding:env:object:configmap
    servicebindingoperator.redhat.io/status.dbConfigMap-db.user: binding:env:object:configmap
    servicebindingoperator.redhat.io/status.dbConnectionIP: binding:env:attribute
    servicebindingoperator.redhat.io/status.dbConnectionPort: binding:env:attribute
    servicebindingoperator.redhat.io/status.dbCredentials-password: binding:env:object:secret
    servicebindingoperator.redhat.io/status.dbCredentials-user: binding:env:object:secret
    servicebindingoperator.redhat.io/status.dbName: binding:env:attribute

```

We can add more annotations to collect values from different kinds of data structures.  

To edit the CRD run the command:

```
oc edit crd databases.postgresql.baiju.dev
```

Add these annotations under `metadata.annotations` along with other annotations.

```
servicebindingoperator.redhat.io/spec.tags: binding:env:attribute
servicebindingoperator.redhat.io/spec.userLabels.archive: binding:env:attribute
servicebindingoperator.redhat.io/spec.secretName: binding:env:attribute
```

These annotations refer to data which can be either a number, string, boolean, or an object or a slice of arbitrary values. In the case of this example, 
- `spec.tags` represents a sequence
- `spec.userLabels` represents a mapping
- `spec.secretName` represents a sequence of mapping

### Application Developer

#### Create a namespace called `service-binding-demo`

The application and the DB needs a namespace to live in so let's create one for them:

```shell
cat <<EOS |kubectl apply -f -
---
kind: Namespace
apiVersion: v1
metadata:
  name: service-binding-demo
EOS
```

Alternatively, you can perform the same task with this make command:

```shell
make create-project
```

#### Import an application

In this example we will import an arbitrary [Java Spring Boot application](https://github.com/ldimaggi/java-rest-http-crud).

In the OpenShift Console switch to the Developer perspective. (Make sure you have selected the `service-binding-demo` project). Navigate to the `+ADD` page from the menu and then click on the `[Import from Git]` button. Fill in the form with the following:

* `Git Repo URL` = `https://github.com/ldimaggi/java-rest-http-crud`
* `Project` = `service-binding-demo`
* `Application`->`Create New Application` = `java-app`
* `Name` = `java-app`
* `Builder Image` = `Java`
* `Create a route to the application` = checked
* `Select the resource type to generate` = Deployment

and click on the `[Create]` button.

Notice, that during the import no DB config was mentioned or requestd.

When the application is running navigate to its route to verify that it is up. Try the application's UI to add a fruit - it causes an error proving that the DB is not connected.

#### Create a DB instance for the application

Now we utilize the DB operator that the cluster admin has installed. To create a DB instance just create a `Database` custom resource in the `service-binding-demo` namespace called `db-demo`:

```shell
cat <<EOS |kubectl apply -f -
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
          - primarySecretName: "example-primaryuser"
            secondarySecretName: "example-secondaryuser"
          - rootSecretName: "example-secretuser"
EOS
```

Alternatively, you can perform the same task with this make command:

```shell
make create-backing-db-instance
```

#### Express an intent to bind the DB and the application

Now, the only thing that remains is to connect the DB and the application. We let the Service Binding Operator to 'magically' do the connection for us.

Create the following `ServiceBindingRequest`:

```shell
cat <<EOS |kubectl apply -f -
---
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
  name: binding-request
  namespace: service-binding-demo
spec:
  applicationSelector:
    resourceRef: java-app
    group: apps
    version: v1
    resource: deployments
  backingServiceSelectors:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    resourceRef: db-demo
    id: postgresDB
  customEnvVar:
  - name: JDBC_URL
    value: 'jdbc:postgresql://{{ .postgresDB.status.dbConnectionIP }}:{{ .postgresDB.status.dbConnectionPort }}/{{ .postgresDB.status.dbName }}'
  - name: DB_USER
    value: '{{ .postgresDB.status.dbCredentials.user }}'
  - name: DB_PASSWORD
    value: '{{ .postgresDB.status.dbCredentials.password }}'
  - name: TAGS
    value: '{{ .postgresDB.spec.tags 0 }}'
  - name: ARCHIVE_USERLABEL
    value: '{{ .postgresDB.spec.userLabels.archive }}'
  - name: SECONDARY_SECRETNAME
    value: '{{ .postgresDB.spec.secretName }}'
EOS
```

Alternatively, you can perform the same task with this make command:

```shell
make create-service-binding-request
```

There are 2 parts in the request:

* `applicationSelector` - used to search for the application based on theresourceRef that we set earlier and the `group`, `version` and `resource` of the application to be a `Deployment` named `java-app`.
* `backingServiceSelector` - used to find the backing service - our operator-backed DB instance called `db-demo`.

That causes the application to be re-deployed.
Once the new version is up, go to the application's route to check the UI. Now, it works!

### Check the status of Service Binding Request

`ServiceBindingRequestStatus` depicts the status of the Service Binding operator. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

To check the status of Service Binding Request, run the command:

```
oc get sbr binding-request -n service-binding-demo -o yaml
```

Status of Service Binding Request on successful binding:

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

* Conditions represent the latest available observations of Service Binding Request's state
* Secret represents the name of the secret created by the Service Binding Operator


Conditions have two types `CollectionReady` and `InjectionReady` 

where

* `CollectionReady` type represents collection of secret from the service
* `InjectionReady` type represents injection of secret into the application

Conditions can have the following type, status and reason:

| Type            | Status | Reason               | Type           | Status | Reason                   |
| --------------- | ------ | -------------------- | -------------- | ------ | ------------------------ |
| CollectionReady | False  | EmptyServiceSelector | InjectionReady | False  |                          |
| CollectionReady | False  | ServiceNotFound      | InjectionReady | False  |                          |
| CollectionReady | True   |                      | InjectionReady | False  | EmptyApplicationSelector |
| CollectionReady | True   |                      | InjectionReady | False  | ApplicationNotFound      |
| CollectionReady | True   |                      | InjectionReady | True   |                          |



### Check secret injection into the application deployment

When the `ServiceBindingRequest` was created the Service Binding Operator's controller injected the DB connection information into the application's `Deployment` as environment variables via an intermediate `Secret` called `binding-request`:

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
➜  ~ oc get pods                               
NAME                                   READY   STATUS      RESTARTS   AGE
db-demo-postgresql-78b8466897-ht8x9    1/1     Running     0          15m
java-rest-http-crud-1-build            0/1     Completed   0          51m
java-rest-http-crud-666b8597cc-w7456   1/1     Running     0          15m
```

To ssh into the application pod `java-rest-http-crud-666b8597cc-w7456`, run command:

```
oc exec -it java-rest-http-crud-666b8597cc-w7456 /bin/bash
```

Print all the environment variables that were injected into the application pod by Service Binding Operator:

```
[jboss@java-rest-http-crud-666b8597cc-w7456 ~]$ printenv | grep DATABASE_
DATABASE_SECRET_USER=postgres
DATABASE_TAGS_0=centos7-12.3
DATABASE_TAGS_1=123
DATABASE_SECRET_PASSWORD=password
DATABASE_CONFIGMAP_DB_NAME=db-demo
DATABASE_DBNAME=db-demo
DATABASE_CONFIGMAP_DB_HOST=172.25.88.63
DATABASE_DBCONNECTIONPORT=5432
DATABASE_USERLABELS_ARCHIVE=false
DATABASE_SECRETNAME_0_PRIMARYSECRETNAME=example-primaryuser
DATABASE_SECRETNAME_0_SECONDARYSECRETNAME=example-secondaryuser
DATABASE_SECRETNAME_1_ROOTSECRETNAME=example-secretuser
DATABASE_CONFIGMAP_DB_PORT=5432
DATABASE_DBCONNECTIONIP=172.25.88.63
DATABASE_CONFIGMAP_DB_PASSWORD=password
```

Notice, distinct environment variables are produced for sibling properties with index number 0 and 1 in the variable name.
