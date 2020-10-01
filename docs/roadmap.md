# Roadmap

## Aim of the project

* Enable application developers to connect their workloads to backing services **with little or no code changes**.

* Enable backing service authors to expose binding information by adding metadata to the the representative Kubernetes objects **with little or no code changes**.

* Provide a managed binding experience.

The Service Binding Operator is being designed to be adopted by application developers and backing service providers with a **low barrier for entry**.


# Key feature areas

### Collection of binding information
* Annotation "Decorators" in backing service resources.
* Descriptor "Decorators" in backing service resources.
* Duck type in backing service resources.
* Service Composition.
* Detection of binding resources.

### Project of binding informattion
* Custom binding variable generation.
* Projection of binding information in workloads.
* Projection in podSpec and non-PodSpec based workloads.

### Security
* Avoiding escalation of privilege.
* Narrow down the service account privileges.

### Release & Packaging
* Packaging of the Service Binding Controller/CRD.

----


# Collection of binding information

## Annotation "Decorators" in backing service resources

Supporting annotations as a way to decorate kubernetes resources enables the following:

* Application developers can use any kubernetes resource a potential backing service
* Backing service providers can annotate their CRDs to indicate what needs to be extracted for binding.

The project aims to support an annotation format such that there's a way to indicate whether a specific unit of binding information is to be bound as a volume mount or an environment variable.

## Improvements

Validate advanced volume mounting support for projecting sensitive information as "files".


## OLM Descriptor "Decorators" in backing service resources

Use of OLM Descriptors as a way to indicate what is interesting for binding enables operator-backed services to add metadata outside of their CRDs to indicate what needs to be extracted for binding.

## Improvements

Validate advanced volume mounting support for projecting sensitive information as "files".

## Duck type in backing service resources


Identify the information that a duck typed resource needs to contain and 
support binding secret generation based on the same.

( needs more information )


## Detection of binding resources

This feature enables operators that manage backing services but which don't
have any metadata in their CSV or CRD to use the Service Binding Operator to bind
together the service and applications. The Service Binding Operator binds all
information 'dependent' to the backing service CR by populating the binding
secret with information from Routes, Services, ConfigMaps, and Secrets owned
by the backing service CR.

The binding is initiated by the introduction of this API option in the backing service CR:
``` yaml
detectBindingResources : true
```
When this API option is set to true, the Service Binding Operator
automatically detects Routes, Services, ConfigMaps, and Secrets owned by the backing service CR.

The generated binding secret needs to be project-able both as a volume mount and environment variable.

## Improvements

Enable specification of mode of injection as volume mounts/"files" or environment variables.


## Service Resource Composition

Binding information for a backing service may need to composed from one or more Kubernetes resources.

```
services:
  - group: postgresql.example.dev
    version: v1alpha1
    kind: DatabaseInstance
    name: database
    id: postgresDBInstance
  - group: postgresql.example.dev
    version: v1alpha1
    kind: DatabaseUser
    name: user
    id: postgresDBuser
```

Service Resource composition enables the application author to enable creation of a 
binding secret from on or more service resources.

## Improvements

Enable support for specifying intent to bind information from specific service resources as volume mounts of environment variables.


# Project of binding informattion

## Custom binding information generation

Staying true to the promise of enabling developers to make minimum code changes to their app, the service binding API aims to empower users to construct custom binding information and inject the same into their workloads, thereby reducing the need for code changes.

Example, the backing service may expose the `host` and `port` individually but the application may need to use the same as a connection string.

The initial implementation involved use of the Go Templates to construct strings to be injected as environment variables.

The improved implementation enables users to compose custom key-value pairs with binding information from multiple service resources.

## Improvements 

Enable injection of custom variables as both environment variables and volume mounts.


## Projection of binding information in workloads

Binding information can be projected either as environment variables or as volume mounts ( "files" ). 

The Service Binding Operator project enables a multi-level decision resolution regarding how the binding information gets projected

1. The backing service author may suggest that specific binding information needs to be injected as files by annotating the CRD or adding a descriptor to indicate the same.

2. The actor who creates the backing service CR may choose to override the suggestion using annotations in the CR.

3. The application author who creates the `ServiceBinding` may choose to override the suggestion using a specific attribute in the `ServiceBinding` CR.

In a nutshell, the application author may choose to consume the recommendation of the backing service author with respect to projection as files or environment variables, or the application author may choose to override the recommendation based on her requirements.

Example, the backing service author may choose to indicate that certificate information needs to be projected as files whereas the username could be projected as environment variables. However, the application author may choose to override the recommendations and project everything as 'files'.


## Improvements

Enable projection of binding information as both environment variables and volume mounts


# Security

To ensure users don't use the service binding operator's service account as a way to escalate privileges, the actor creating the  `ServiceBinding` CR should not be able to indirectly read any Kubernetes object which it was otherwise unauthorized to do so.

## Improvements

* To accomplish the above, one needs to deploy a "Validating Admission Webhook" that would do Subject Access Reviews ( SARs ) on relevant objects before admitting the `ServiceBinding` object.

* Review and potentially reduce the permissions granted to the controller's service account.


# Release & Packaging

Even though the Service Binding Operator is compatible with vanilla Kubernetes without OLM, upstream releases are today delivered as an OpenShift Community OLM operator. 

**Improvements** 

As part of the release process, the Service Binding Controller/CRD needs to be made available in common distribution channels such as 

* A public Helm Repository
* An "install.yaml" generated on every release that could be installed as a `kubectl apply -f ... `
* A community Kubernetes Operator on [operatorhub.io](https://operatorhub.io)


---

The Roadmap document is a live document, PRs to include new features are welcome.
