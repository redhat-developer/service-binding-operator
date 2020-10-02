# The Service Binding Operator
## Connecting Applications with Services on Kubernetes and OpenShift

<p align="center">
    <a alt="GoReport" href="https://goreportcard.com/report/github.com/redhat-developer/service-binding-operator">
        <img alt="GoReport" src="https://goreportcard.com/badge/github.com/redhat-developer/service-binding-operator">
    </a>
    <a href="https://godoc.org/github.com/redhat-developer/service-binding-operator">
        <img alt="GoDoc Reference" src="https://godoc.org/github.com/redhat-developer/service-binding-operator?status.svg">
    </a>
    <a href="https://codecov.io/gh/redhat-developer/service-binding-operator">
        <img alt="Codecov.io - Code Coverage" src="https://codecov.io/gh/redhat-developer/service-binding-operator/branch/master/graph/badge.svg">
    </a>
</p>

## Introduction

The goal of the Service Binding Operator is to enable application authors to
import an application and run it on Kubernetes with services
such as databases represented as Kubernetes objects including Operator-backed and chart-based backing services, without having to perform manual configuration of `Secrets`,
`ConfigMaps`, etc.

To make a service bindable, the service provider needs to express
the information needed by applications to bind with the services. In other words, the service provider must express the
information that's “interesting” to applications.

There are multiple methods for making backing services
bindable, including the backing service provider providing metadata as
annotations on the resource. Details on the methods for making backing services bindable
are available in the [Backing Service Provider Best Practices Guide](docs/BackingServiceBestPractices.md)

To make an imported application (for example, a NodeJS application)
connect to a backing service (for example, a database):

* The app author (developer) creates a `ServiceBinding` and specifies:
  * The resource that needs the binding information. The resource can be
    specified by label selectors;
  * The backing service's resource reference that the imported application
    needs to be bound to;

* The Service Binding Controller then:
  * Reads backing service operator CRD annotations to discover the
    binding attributes
  * Creates a binding secret for the backing service, example, an operator-managed database;
  * Injects environment variables into the applications' `Deployment`, `DeploymentConfig`,
    `Replicaset`, `KnativeService` or anything that uses a standard PodSpec;

### Example
#### Binding a Java Application with a Database

``` yaml
apiVersion: operators.coreos.com/v1alpha1
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
```

## User guide


* [Application Workload Author's Guide](docs/application-author.md)
* [Backing Service Provider Best Practices Guide](docs/BackingServiceBestPractices.md)




## Dependencies

| Dependency                                | Supported versions           |
| ----------------------------------------- | ---------------------------- |
| [Kubernetes](https://kubernetes.io/)      |  v1.17.\* or higher.        |


## Quick Start

Clone the repository and run `make local` in an existing `kube:admin` OpenShift
CLI session. Alternatively, install the operator using:

``` bash
cat <<EOS |kubectl apply -f -
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


## Key Features

* Support Binding with backing services represented by Kubernetes resources including third-party CRD-backed resources.
* Support binding with multiple-backing services.
* Extract binding information based on annotations present in CRDs/CRs/resources. 
* Extract binding values based on annotations present in OLM descriptors.
* Project binding values as volume mounts.
* Project binding values as environment variables.
* Binding of PodSpec-based workloads.
* Binding of non-PodSpec-based Kubernetes resources.
* Custom binding variables composed from one or more backing services.
* Auto-detect binding resources in the absence of binding decorators.


## Getting Started

The best way to get started with the Service Binding Operator is to see it in action.

A number of example scenarios for using the operator are included in this
repo. The examples are found in the "/examples" directory. Each of these
examples illustrates a usage scenario for the operator. Each example also
includes a README file with step-by-step instructions for how to run the
example.

The following section in this README file includes links to the current set of examples.

## Example Scenarios

The following example scenarios are available:

[Binding an Imported app with an In-cluster Operator Managed PostgreSQL Database](examples/nodejs_postgresql/README.md)

[Binding an Imported app with an Off-cluster Operator Managed AWS RDS Database](examples/nodejs_awsrds_varprefix/README.md)

[Binding an Imported Java Spring Boot app with an In-cluster Operator Managed PostgreSQL Database](examples/java_postgresql_customvar/README.md)

[Binding an Imported Quarkus app deployed as Knative service with an In-cluster Operator Managed PostgreSQL Database](examples/knative_postgresql_customvar/README.md)

[Binding an Imported app with an In-cluster Operator Managed ETCD Database](examples/nodejs_etcd_operator/README.md)

[Binding an Imported app to an Off-cluster Operator Managed IBM Cloud Service](examples/nodejs_ibmcloud_operator/README.md)

[Binding an Imported app in one namespace with an In-cluster Managed PostgreSQL Database in another namespace](examples/nodejs_postgresql_namespaces/README.md)

[Binding an Imported app to a Route/Ingress](examples/route_k8s_resource/README.md)

## Community, discussion, contribution, and support

The Service Binding community meets weekly on Thursdays at 11:15 AM UTC via [Google Meet](https://meet.google.com/jti-pujt-zqs).

Meeting Agenda is maintained [here](https://github.com/redhat-developer/service-binding-operator/issues?q=is%3Aissue+is%3Aopen+label%3Ameeting)

Please file bug reports on [Github](https://github.com/redhat-developer/service-binding-operator/issues/new). For any other questions, reach out on [service-binding-support@redhat.com](https://www.redhat.com/mailman/listinfo/service-binding-support).

Join the [service-binding-operator](https://app.slack.com/client/T09NY5SBT/C019LQYGC5C) channel in the [Kubernetes Workspace](https://slack.k8s.io/) for any discussions and collaboration with the community.
