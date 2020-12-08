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
screen, and then the `Container image` card. Fill the `Image name from external registry` with
`quay.io/redhat-developer/sbo-generic-test-app:20200923`. Openshift will fill rest things correctly.
Name of application will be `sbo-generic-test-app`.
Check `Create a route to the application` option at the bottom.
Push the `Create` button to create the application.

Create the ServiceBinding `service-binding-multiple-services` by pushing the `(+)` button
on the top right corner and pasting the following:

```yaml
---
apiVersion: operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: service-binding-multiple-services
spec:
  application:
    name: sbo-generic-test-app
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

***After the binding is processed***

#### Check the status of Service Binding

`ServiceBinding Status` depicts the status of the Service Binding operator. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status

To check the status of Service Binding, run the command:

```
kubectl get servicebinding service-binding-multiple-services -o yaml
```

Status of Service Binding on successful binding:

```yaml
status:
  conditions:
  - lastHeartbeatTime: "2020-12-08T09:03:12Z"
    lastTransitionTime: "2020-12-08T09:02:44Z"
    status: "True"
    type: CollectionReady
  - lastHeartbeatTime: "2020-12-08T09:03:12Z"
    lastTransitionTime: "2020-12-08T09:02:44Z"
    status: "True"
    type: InjectionReady
  - lastHeartbeatTime: "2020-12-08T09:03:12Z"
    lastTransitionTime: "2020-12-08T09:02:44Z"
    status: "True"
    type: Ready
  secret: service-binding-multiple-services
```

#### Verify the data injected

You can check env var by executing `curl route-of-the-application/env`.
You can get route by `oc get route` command.

```shell
curl sbo-generic-test-app-test-multiple-services.apps.dev-svc-4.6-120807.devcluster.openshift.com/env
```

```json
{
"DATABASE_DBCONNECTIONIP": "172.30.228.141",
"DATABASE_DBCONNECTIONPORT": "5432",
"DATABASE_DBNAME": "db-demo",
"DATABASE_DB_HOST": "172.30.228.141",
"DATABASE_DB_NAME": "db-demo",
"DATABASE_DB_PASSWORD": "password",
"DATABASE_DB_PORT": "5432",
"DATABASE_DB_USER": "postgres",
"DATABASE_PASSWORD": "password",
"DATABASE_USER": "postgres",
"DB_DEMO_POSTGRESQL_PORT": "tcp://172.30.228.141:5432",
"DB_DEMO_POSTGRESQL_PORT_5432_TCP": "tcp://172.30.228.141:5432",
"DB_DEMO_POSTGRESQL_PORT_5432_TCP_ADDR": "172.30.228.141",
"DB_DEMO_POSTGRESQL_PORT_5432_TCP_PORT": "5432",
"DB_DEMO_POSTGRESQL_PORT_5432_TCP_PROTO": "tcp",
"DB_DEMO_POSTGRESQL_SERVICE_HOST": "172.30.228.141",
"DB_DEMO_POSTGRESQL_SERVICE_PORT": "5432",
"DB_DEMO_POSTGRESQL_SERVICE_PORT_DB_DEMO_POSTGRESQL": "5432",
"ETCDCLUSTER_CLUSTERIP": "172.30.205.77",
"ETCDCLUSTER_DB_HOST": "172.30.228.141",
"ETCDCLUSTER_DB_NAME": "db-demo",
"ETCDCLUSTER_DB_PASSWORD": "password",
"ETCDCLUSTER_DB_PORT": "5432",
"ETCDCLUSTER_DB_USER": "postgres",
"ETCDCLUSTER_PASSWORD": "cGFzc3dvcmQ=",
"ETCDCLUSTER_USER": "cG9zdGdyZXM=",
"ETCD_DEMO_CLIENT_PORT": "tcp://172.30.205.77:2379",
"ETCD_DEMO_CLIENT_PORT_2379_TCP": "tcp://172.30.205.77:2379",
"ETCD_DEMO_CLIENT_PORT_2379_TCP_ADDR": "172.30.205.77",
"ETCD_DEMO_CLIENT_PORT_2379_TCP_PORT": "2379",
"ETCD_DEMO_CLIENT_PORT_2379_TCP_PROTO": "tcp",
"ETCD_DEMO_CLIENT_SERVICE_HOST": "172.30.205.77",
"ETCD_DEMO_CLIENT_SERVICE_PORT": "2379",
"ETCD_DEMO_CLIENT_SERVICE_PORT_CLIENT": "2379",
"ETCD_RESTORE_OPERATOR_PORT": "tcp://172.30.149.218:19999",
"ETCD_RESTORE_OPERATOR_PORT_19999_TCP": "tcp://172.30.149.218:19999",
"ETCD_RESTORE_OPERATOR_PORT_19999_TCP_ADDR": "172.30.149.218",
"ETCD_RESTORE_OPERATOR_PORT_19999_TCP_PORT": "19999",
"ETCD_RESTORE_OPERATOR_PORT_19999_TCP_PROTO": "tcp",
"ETCD_RESTORE_OPERATOR_SERVICE_HOST": "172.30.149.218",
"ETCD_RESTORE_OPERATOR_SERVICE_PORT": "19999",
"HOME": "/",
"HOSTNAME": "sbo-generic-test-app-865978dbc-8fgjs",
"KUBERNETES_PORT": "tcp://172.30.0.1:443",
"KUBERNETES_PORT_443_TCP": "tcp://172.30.0.1:443",
"KUBERNETES_PORT_443_TCP_ADDR": "172.30.0.1",
"KUBERNETES_PORT_443_TCP_PORT": "443",
"KUBERNETES_PORT_443_TCP_PROTO": "tcp",
"KUBERNETES_SERVICE_HOST": "172.30.0.1",
"KUBERNETES_SERVICE_PORT": "443",
"KUBERNETES_SERVICE_PORT_HTTPS": "443",
"NSS_SDB_USE_CACHE": "no",
"PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
"PWD": "/",
"SBO_GENERIC_TEST_APP_PORT": "tcp://172.30.95.166:8080",
"SBO_GENERIC_TEST_APP_PORT_8080_TCP": "tcp://172.30.95.166:8080",
"SBO_GENERIC_TEST_APP_PORT_8080_TCP_ADDR": "172.30.95.166",
"SBO_GENERIC_TEST_APP_PORT_8080_TCP_PORT": "8080",
"SBO_GENERIC_TEST_APP_PORT_8080_TCP_PROTO": "tcp",
"SBO_GENERIC_TEST_APP_SERVICE_HOST": "172.30.95.166",
"SBO_GENERIC_TEST_APP_SERVICE_PORT": "8080",
"SBO_GENERIC_TEST_APP_SERVICE_PORT_8080_TCP": "8080",
"SHLVL": "1",
"ServiceBindingOperatorChangeTriggerEnvVar": "45838",
"TERM": "xterm"
}
```

That's it, folks!
