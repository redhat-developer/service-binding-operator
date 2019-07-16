# Connect Applications with Operator-backed Services

## Introduction

The goal of the Service Binding Operator is to enable application authors to 
import an application and run it on OpenShift with operator-backed services such as 
databases, without having to perform manual configuration of secrets, 
configmaps, etc. 

In order for the Service Binding Operator to bind an application to a backing
service, the backing service operator must specify the information required 
by the application to bind to the operator's service. The information must
be specified in the operator's OLM descriptor from which it will be extracted
to bind the application to the operator. The information could be specified in
the "status" and/or "spec" section of the OLM in plaintext or as a reference 
to a secret.

In order to make an imported application (for example, a NodeJS 
application) connect to a backing services (for example, a database):

* The app author (developer) creates a ServiceBindingRequest and specifies
  * The resource that needs the binding information. The resource can 
be specified by label selectors. 
  * The backing service's resource reference that the imported application 
needs to be bound to.

* The Service Binding Controller then:
  * Reads backing service operator OLM descriptor to discover the binding
    attributes
  * Creates a binding secret for the backing service, example, 
    an operator-managed database.
  * Injects environment variables into the applications's DeploymentConfig,
    Deployment or Replicaset

Here is an example of the "bind-able" operator OLM Descriptor - in this case 
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
 
 Clone the repository and run `make local` in an existing 
 `kube:admin` openshift CLI session.

 Alternatively, install the operator using

 ```yaml:
 --- 
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata: 
  name: redhat-developer-operators
  namespace: openshift-marketplace
spec: 
  endpoint: "https://quay.io/cnr"
  registryNamespace: redhat-developer
  type: appregistry

```

## Example Scenario

*What the Cluster Admin does*

* A cluster admin installs into the cluster a Backing Service Operator that
 is "bind-able," in other words a Backing Service Operator that exposes binding
 information in secrets, status, and/or spec attributes. The Backing Service 
 Operator may represent a database or other services required by applications.
 We'll use  https://github.com/baijum/postgresql-operator to demonstrate a 
 sample use case.

* The cluster-admin installs the database operator
  using an `OperatorSource`
```yaml:
    --- 
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata: 
  name: db-operators
  namespace: openshift-marketplace
spec: 
  endpoint: "https://quay.io/cnr"
  registryNamespace: pmacik
  type: appregistry
```

  * For example, Creation of a PostgreSQL operator-managed database 
 exposes the following binding information: 

```yaml:
    --- 
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
  dbConnectionIP: "172.30.24.255"
  dbConnectionPort: 5432
  dbCredentials: demo-database-postgresql
  dbName: postgres
    
```

* The cluster-admin installs the Service Binding operator
  using an `OperatorSource` if not already done using 
  `make local`

```yaml:
     --- 
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata: 
  name: redhat-developer-operators
  namespace: openshift-marketplace
spec: 
  endpoint: "https://quay.io/cnr"
  registryNamespace: redhat-developer
  type: appregistry
```


*What the Application Author (Developer) does*

* The application an application - During the import, the application doesn't
make any other choices than the ones she makes in the normal import flow. 

* After the application import, the application author creates a 
ServiceBindingRequest. The Custom Resource includes non-sensitive 
information (connection URL) and a secret for sensitive information 
(username and password).

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
    resourceRef: db-demo
    resourceKind: databases.postgresql.baiju.dev
    resourceVersion: v1alpha1
```

The controller injects the binding information as specified in the OLM descriptor below, into the application's `DeploymentConfig` as environment variables:

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


