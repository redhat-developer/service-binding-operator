# Service Binding Operator: Connect Applications with Services

## Operator Goal - Binding Imported Applications with Operator-Backed Services

The goal of the Service Binding Operator is to enable application authors to 
import an application and run it on OpenShift with backing services such as 
databases, without having to perform manual configuration of secrets, 
configmaps, etc. 

In order for the Service Binding Operator to bind an application to a backing
service, the backing service operator must specify the information required 
to for the application to bind to the operator's service. The information must
be specified in the operator's OLM descriptor from which it will be extracted
to bind the application to the operator. The information can be specified in a
secret, or in the "status" and/or "spec" section of the OLM.

In order to make an imported application (for example, as a NodeJS 
application) connect to a backing services (for example as a database):

* The app author (developer) creates a ServiceBindingRequest and specifies
a label in the form of "connects-to." For example, to bind to an
operator-backed database service created using a CR named "my-db", the label
will be "connects-to=my-db".

* The Service Binding Operator then:
  * Reads backing service operator OLM descriptor to discover the binding
attributes
  * Creates a binding secret for the database
  * Injects environment variables into the applications's DeploymentConfig,
Deployment, Replicaset

Here is an example of the "bind-able" operator OLM Descriptor - in ths case 
for a PostgreSQL database backing operator:

```yaml:
statusDescriptors:
description: Name of the Secret to hold the DB user and password
  displayName: DB Password Credentials
  path: dbCredentials

  x-descriptors:
    - urn:alm:descriptor:io.kubernetes:Secret
    - urn:alm:descriptor:servicebindingrequest:env:object:secret:user
    - urn:alm:descriptor:servicebindingrequest:env:object:secret:password

description: Database connection IP address
  displayName: DB IP address
  path: dbConnectionIP
  x-descriptors:
    - urn:alm:descriptor:servicebindingrequest:env:attribute
    - description: Database connection port
      displayName: DB port
      path: dbConnectionPort
      x-descriptors:
    - urn:alm:descriptor:servicebindingrequest:env:attribute
```

## Quick Start

To build:
```
operator-sdk build quay.io/<username>/service-binding-operator:1
```

To push image:
```
docker login quay.io -u <username>
docker push quay.io/<username>/service-binding-operator:1
```

To run the operator locally:
```
kubectl apply -f deploy/service_account.yaml
kubectl apply -f deploy/role.yaml
kubectl apply -f deploy/role_binding.yaml
kubectl apply -f deploy/crds/apps_v1alpha1_servicebindingrequest_crd.yaml

And then you can: 
operator-sdk up local --verbose --namespace default
```

## Example Use Scenario

Cluster Admin:
* A cluster admin installs into the cluster a Backing Service Operator that
is "bind-able," in other words a Backing Service Operator that exposes binding
information in secrets, status, and/or spec attributes.
* The Backing Service Operator may represent a database or other services
required by applications.
* For example, this PostgreSQL operator exposes binding information: 
https://github.com/baijum/postgresql-operator

```yaml:
apiVersion: postgresql.baiju.dev/v1alpha1
kind: Database
metadata:
  name: my-db
  namespace: openshift-operators
spec:
  dbName: postgres
  image: docker.io/postgres
  imageName: postgres
status:
  dbConnectionIP: 172.30.24.255
  dbConnectionPort: 5432
  dbCredentials: demo-database-postgresql
  dbName: postgres
```

Application Author (Developer)

* The application an application - During the import, the application doesn't
make any other choices than the ones she makes in the normal import flow. 
* The application pod is created automatically
* After the application import, the application author creates a 
ServiceBindingRequest for the Backing Service Operator. The Custom Resource 
includes a configmap for non-sensitive information (connection URL) and a 
secret for sensitive information (username and password). The application 
author also specifies a label of "connects-to=backing-service-name" in order
to bind to the Backing Service Operator's service.

Here's an example of a Service Binding Request - for the PostgreSQL database:

```yaml:
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
  name: binding-request
  namespace: openshift-operators
spec:
  applicationSelector:
    matchLabels:
    connects-to: db-demo
    environment: demo
    resourceKind: DeploymentConfig
  backingSelector:
    objectName: db-demo
    resourceName: postgresql.baiju.dev
    resourceVersion: v1alpha1
```

* Note that the creation of a backing service Custom Resource takes form of
an OpenShift API call to create the Custom Resource, followed by the operator
reconciling and creating the backing service.
* At this point, the following actions take place - all of which are hidden
from the application author:
  * The Service Binding Operator looks for OLM descriptors on the backing
service to identify the bind-able attributes
  * The Backing Service Operator creates the Secret and ConfigMap
  * The Service Binding Operator controller modifies the imported 
application's pods deployment/pod spec by injecting the bindable Secret and
ConfigMap required to access the database provided by the Backing Service
Operator through setting environment variables in these resources such as
Deployment (Kubernetes and OpenShift), DeploymentConfig (Openshift),
StatefulSet (Kubernetes and Openshift).
  * If the App Author is watching the status of the imported app's pods,
he/she will observe the pod being created and starting, then the pod is
restarted after the Secret and ConfigMap are injected.
* Now, the application can access and make use of the service provided by
the Backing Service Operator.

