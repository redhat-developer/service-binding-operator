# Multiple Services Binding Example

This document describes step-by-step the actions to create the required
infrastructure to demonstrate multiple services binding built on existing
Postgres and ETCD examples.

As *Cluster Administrator*, the reader will install both the "PostgreSQL
Database" and the "ETCD" operators, as described below.

Once the cluster setup is finished, the reader will create a Postgres
database and a ETCD cluster, and bind services to a Node.js application as
a *Developer*.

## Cluster Configuration

### Create a New Project

Create a new project, in this example it is called `multiple-services-demo`.

### Install the Postgres Operator

Switch to the *Administrator* perspective.

Add an extra OperatorSource by pushing the "+" button on the top right corner
and pasting the following:

```yaml
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
```

Go to "Operators > OperatorHub", search for "Postgres" and install "PostgreSQL
Database" provided by Red Hat.

Select "A specific namespace on the cluster" in "Installation Mode", select the
"multiple-services-demo" namespace in "Installed Namespace" and push "Subscribe".

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

### Install the ETCD Operator

Go to "Operators > OperatorHub", search for "etcd" and install "etcd" provided by
CNCF.

Select "A specific namespace on the cluster" in "Installation Mode", select the
"multiple-services-demo" namespace in "Installed Namespace" and push "Subscribe".

## Application Configuration

Switch to the *Developer* perspective.

Create the Postgres database `db-demo` by pushing the `(+)` button on the top right
corner and pasting the following:

```yaml
---
apiVersion: postgresql.baiju.dev/v1alpha1
kind: Database
metadata:
  name: db-demo
  namespace: multiple-services-demo
spec:
  image: docker.io/postgres
  imageName: postgres
  dbName: db-demo
```

Create the ETCD cluster `etcd-demo` by pushing the `(+)` button on the top right
corner and paste the following:

```yaml
---
apiVersion: etcd.database.coreos.com/v1beta2
kind: EtcdCluster
metadata:
  name: etcd-demo
  namespace: multiple-services-demo
spec:
  size: 3
  version: "3.2.13"
```

Import the application by pushing the `+Add` button on the left side of the
screen, and then the `From Git` card. Fill the `Git Repo URL` with
`https://github.com/akashshinde/node-todo.git`; the repository will be
validated and the appropriate `Builder Image` and `Builder Image Version`
will be selected. Push the `Create` button to create the application.

Create the ServiceBinding `node-todo-git` by pushing the `(+)` button
on the top right corner and pasting the following:

```yaml
---
apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: node-todo-git
spec:
  application:
    name: node-todo-git
    group: apps
    version: v1
    resource: deployments
  services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: db-demo
  - group: etcd.database.coreos.com
    version: v1beta2
    kind: EtcdCluster
    name: etcd-demo
  detectBindingResources: true
```

Once the binding is processed, the secret can be verified by executing
```shell
kubectl get secrets node-todo-git -o yaml
```
```yaml
apiVersion: v1
data:
  DATABASE_DB_HOST: MTcyLjMwLjU2LjM0
  DATABASE_DB_NAME: ZGItZGVtbw==
  DATABASE_DB_PASSWORD: cGFzc3dvcmQ=
  DATABASE_DB_PORT: NTQzMg==
  DATABASE_DB_USER: cG9zdGdyZXM=
  DATABASE_DBCONNECTIONIP: MTcyLjMwLjU2LjM0
  DATABASE_DBCONNECTIONPORT: NTQzMg==
  DATABASE_DBNAME: ZGItZGVtbw==
  DATABASE_IMAGE: ZG9ja2VyLmlvL3Bvc3RncmVz
  DATABASE_IMAGENAME: cG9zdGdyZXM=
  DATABASE_PASSWORD: cGFzc3dvcmQ=
  DATABASE_USER: cG9zdGdyZXM=
  ETCDCLUSTER_CLUSTERIP: MTcyLjMwLjIwOC4yMw==
  ETCDCLUSTER_DB_HOST: MTcyLjMwLjU2LjM0
  ETCDCLUSTER_DB_NAME: ZGItZGVtbw==
  ETCDCLUSTER_DB_PASSWORD: cGFzc3dvcmQ=
  ETCDCLUSTER_DB_PORT: NTQzMg==
  ETCDCLUSTER_DB_USER: cG9zdGdyZXM=
  ETCDCLUSTER_PASSWORD: Y0dGemMzZHZjbVE9
  ETCDCLUSTER_USER: Y0c5emRHZHlaWE09
kind: Secret
metadata:
  ...
  name: node-todo-git
  namespace: multiple-services-demo
  ...
type: Opaque
```
#### Check the status of Service Binding

`ServiceBinding Status` depicts the status of the Service Binding operator. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

To check the status of Service Binding, run the command:

```
kubectl get servicebinding binding-request -o yaml
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
  secret: binding-request-72ddc0c540ab3a290e138726940591debf14c581
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
