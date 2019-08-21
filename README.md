# Connecting Applications with Operator-backed Services

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

The goal of the Service Binding Operator is to enable application authors to import an application
and run it on OpenShift with operator-backed services such as databases, without having to perform
manual configuration of secrets, configmaps, etc.

In order for the Service Binding Operator to bind an application to a backing service, the backing
service operator must specify the information required by the application to bind to the
operator's service. The information must be specified in the operator's OLM (Operator Lifecycle
Manager) descriptor from which it will be extracted to bind the application to the operator. The
information could be specified in the "status" and/or "spec" section of the OLM in plaintext or as
a reference to a secret.

In order to make an imported application (for example, a NodeJS application) connect to a backing
services (for example, a database):

* The app author (developer) creates a `ServiceBindingRequest` and specifies:
  * The resource that needs the binding information. The resource can be specified by label
    selectors;
  * The backing service's resource reference that the imported application needs to be bound to;

* The Service Binding Controller then:
  * Reads backing service operator OLM descriptor to discover the binding attributes
  * Creates a binding secret for the backing service, example, an operator-managed database;
  * Injects environment variables into the applications's `DeploymentConfig`, `Deployment` or
    `Replicaset`;

Here is an example of the *bind-able* operator OLM Descriptor -- in this case for a PostgreSQL
database backing operator:

``` yaml
---
[...]
statusDescriptors:
  description: Name of the Secret to hold the DB user and password
    displayName: DB Password Credentials
    path: dbCredentials
    x-descriptors:
      - urn:alm:descriptor:io.kubernetes:Secret
      - binding:env:object:secret:user
      - binding:env:object:secret:password
  description: Database connection IP address
    displayName: DB IP address
    path: dbConnectionIP
    x-descriptors:
      - binding:env:attribute
```

## Quick Start

Clone the repository and run `make local` in an existing `kube:admin` openshift CLI session.
Alternatively, install the operator using:

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

## Getting Started

The best way to get started with the Service Binding Operator is to see it in action. 

We've included a number of examples scenarios for using the operator in this repo. The examples are found in the "/examples" directory. Each of these examples illustrates a usage scenario for the operator. Each example also includes a README file with step-by-step instructions for how to run the example. 

We'll add more examples in the future. The following section in this README file includes links to the current set of examples. 

## Example Scenarios

The following example scenarios are available:

[Binding an Imported app to an In-cluster Operator Managed PostgreSQL Database](examples/bind_imported_app_to_incluster_operator_managed_PostgreSQL_db/README.md)

[Binding an Imported app to an Off-cluster Operator Managed AWS RDS Database](examples/bind_imported_app_to_offcluster_operator_managed_AWS_RDS_db/README.md)



