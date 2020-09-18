---
title: Implementing Service Binding Specification for Kubernetes
authors:
  - "@arturdzm"
  - "@navidsh"
reviewers:
  - TBD
approvers:
  - TBD
creation-date: 2020-09-09
last-updated: 2020-09-09
status: provisional|implementable|implemented|deferred|rejected|withdrawn|replaced
see-also:
  - N/A
replaces:
  - N/A
superseded-by:
  - N/A
---

# Implementing Service Binding Specification for Kubernetes

## Release Signoff Checklist

- [ ] Enhancement is `implementable`
- [ ] Design details are appropriately documented from clear requirements
- [ ] Test plan is defined
- [ ] Graduation criteria for alpha, beta, GA maturity
- [ ] User-facing documentation is created in [docs](/docs/)

## Open Questions [optional]

This is where to call out areas of the design that require closure before deciding to implement the
design. For instance:

> 1. This prevents us from doing cross-namespace binding. Should we do this?

## Summary

This document proposes a mechanism for the Service Binding Operator to implement the
[Service Binding Specification for Kubernetes](https://github.com/k8s-service-bindings/spec).

## Motivation

The [Service Binding Specification for Kubernetes](https://github.com/k8s-service-bindings/spec)
aims to create a mechanism that is widely applicable. The benefit of a Kubernetes-wide service binding
specification is that all of the actors in an ecosystem, including application developers and service
providers, can work towards a clearly defined abstraction at the edge of their expertise and depend on
other parties to complete the chain.

On the other hand, the Service Binding Operator has a well-defined mechanism for binding applications
and services together. This mechanism is different from how the Service Binding Specification for
Kubernetes lays out binding between entities. Generally, most features defined in the specification
are a subset of features defined in the Service Binding Operator.

There are major differences between the specification and the current mechanism the Service Binding
Operator employs such as how the application is defined in the CRD, how a binding secret is
projected in an application container and etc. The [Service Binding feature comparison document](./service-binding-feature-comparison.pdf) lists the differences between the two.

One of the major gaps between the CRDs defined by the two approaches is that the Service Binding
Operator supports defining multiple services within a CR. Whereas the specification allows defining
one service per CR. Here are two main reasons why the specification is in favor of defining one
service per CR:

- It is easier for tools created around the service binding to add and remove services by simply
  adding or removing independent CRs instead of having to deal with editing an array of services
  within a CR. In addition, editing the service array in a CR might cause CR to be corrupted unless
  concurrent updates are well synchronized across a tool or tools. However, by defining a single
  service per CR, adding or removing services simply becomes `kubectl apply ...` and
  `kubectl delete ...`.

- Defining one service per CR allows cleaner association of attributes with the service. For example,\
  when we identify a set of extra mappings or environment variables in a `ServiceBinding`, which
  service does that get associated with if we have multiple services? Or if we want to change the
  service name or type that gets mounted, we again run into the issue that they all require a way
  to identify the service you want to apply those attributes against.

Therefore, allowing users of the Service Binding Operator to choose between the specification and the
current mechanism is likely to increase the user base of the operator, having a more vibrant community
of service providers.

### Goals

Goals for this proposal are listed as follows:

- Allow users of the Service Binding Operator to have a clear separation in choosing between the
specification and the current mechanism for binding applications to services.

### Non-Goals

- Implementation details about how the specification should be implemented.

## Proposal

To avoid tangling the implementation between the current codebase and the codebase to support the
specification, this proposal suggests creating two different channels for the Service Binding
Operator (e.g. `current` and `spec`) to avoid mixing of implementation details between the two
codebases. This way users can freely choose what set of resources they expect the operator to act on at
the operator install time.

### User Stories

#### Story 1

When installing the Service Binding Operator, a user can choose which of the two channels
to install the Service Binding Operator from. Depending on the channel, the operator installs different
sets of Custom Resource Definitions (CRDs) and would watch for different sets of resources in the
cluster.

### Implementation Details/Notes/Constraints

Providing implementation details is not the goal of this proposal.

## Design Details

### Upgrade / Downgrade Strategy

Since this proposal suggests the operator to be released under two different channels, upgrade/downgrade
scenarios do not apply.

## Drawbacks

- Having two different channels and sets of CRDs could confuse users and thus decrease UX.
- Other tools in the ecosystem potentially need to support both implementations/channels (CRD, APIs) and
thus more effort on their side.

## Alternatives

The alternative is having a separate operator to implement the specification.