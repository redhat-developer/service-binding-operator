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

Service Binding manages the data plane for applications and backing services.
Service Binding Operator reads data made available by the control plane of
backing services and projects the data to applications according to the rules
provided via ServiceBinding resource.

![service-binding-intro](/docs/userguide/modules/ROOT/assets/images/intro-bindings.png)

### Why Service Bindings?

Today in Kubernetes, the exposure of secrets for connecting applications to
external services such as REST APIs, databases, event buses, and many more is
manual and bespoke.  Each service provider suggests a different way to access
their secrets, and each application developer consumes those secrets in a custom
way to their applications.  While there is a good deal of value to this
flexibility level, large development teams lose overall velocity dealing with
each unique solution.

Service Binding:
* Enables developers to connect their application to backing services with a
  consistent and predictable experience
* Removes error-prone manual configuration of binding information
* Provides service operators a low-touch administrative experience to provision
  and manage access to services
* Enriches development lifecycle with a consistent and declarative service
  binding method that eliminates environments discrepancies

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

## Service Binding Specification Alignment

* Service Binding Operator provides two different APIs.
    * binding.operators.coreos.com/v1alpha1: This API is compliant with the Service Binding Specification for Kubernetes.
    * servicebinding.io/v1alpha3 (tech preview): This API implements the Service Binding Specification for Kubernetes.

The [Service Binding Specification for Kubernetes](https://github.com/servicebinding/spec) is still evolving and maturing.  We are tracking changes to the spec as it approaches a stable release and are updating our APIs accordingly and as a result our APIs may change in the future.

## Getting started

### Installing in a Cluster

Follow [OperatorHub instructions](https://operatorhub.io/operator/service-binding-operator).

### Usage

To get started, consult the [quick start
tutorial](https://redhat-developer.github.io/service-binding-operator/userguide/getting-started/quick-start.html).
General documentation can be found
[here](https://redhat-developer.github.io/service-binding-operator/).

### Read more

Here are some more places to read about SBO in use:

* [Announcing Service Binding Operator 1.0 GA](https://developers.redhat.com/articles/2021/10/27/announcing-service-binding-operator-10-ga)
* [How to use service binding with RabbitMQ](https://developers.redhat.com/articles/2021/11/03/how-use-service-binding-rabbitmq)
* [Bind workloads to services easily with the Service Binding Operator and Red Hat OpenShift](https://developers.redhat.com/articles/2022/03/11/binding-workloads-services-made-easier-service-binding-operator-red-hat)
* [Drag&Drop JAR and Service Binding](https://www.youtube.com/watch?v=zb1m31i7EYA)
* [Bind a Kafka cluster to a Node.js application the easy way](https://developers.redhat.com/articles/2022/04/21/bind-kafka-cluster-nodejs-application-easy-way)

## Known bindable operators

The Service Binding Operator can automatically detect and bind to services
created by a limited selection of operators.  These operators do not support
binding directly.  Instead, the service binding operator is able to detect and
configure the operator's CRDs so that they become bindable.  The long-term
intention is to contribute upstream support for service binding and remove the
operators that gain native support for service bindings.  The operators that
currently fall in this category are:

* [OpsTree Redis](https://operatorhub.io/operator/redis-operator): bindable with
  `Redis.redis.redis.opstreelabs.in/v1beta1` services
* [CrunchyData Postgres](https://operatorhub.io/operator/postgresql): bindable
  with `PostgresCluster.postgres-operator.crunchydata.com/v1beta1` services
* [Cloud Native
  PostgreSQL](https://operatorhub.io/operator/cloud-native-postgresql): bindable
  with `Cluster.postgresql.k8s.enterprisedb.io/v1` services
* [Percona XtraDB
  Cluster](https://operatorhub.io/operator/percona-xtradb-cluster-operator):
  bindable with `PerconaXtraDBCluster.pxc.percona.com/v1-8-0` and `v1-9-0`
  services
* [Percona
  MongoDB](https://operatorhub.io/operator/percona-server-mongodb-operator):
  bindable with `PerconaServerMongoDB.psmdb.percona.com/v1-9-0` and `v1-10-0`
  services
  * NOTE: Provides administrative access to the cluster by default
* [RabbitMQ Cluster](https://github.com/rabbitmq/cluster-operator): bindable
  with `RabbitmqCluster.rabbitmq.com/v1beta1` services

OpenShift Streams for Apache Kafka are also bindable, although getting binding
to work requires a little more effort.  See [here][kafka] for more details.

## Roadmap

The direction of this project is tracked under
[milestones](https://github.com/redhat-developer/service-binding-operator/milestones)
posted here on GitHub.

## Community, discussion, contribution, and support

The Service Binding community meets weekly on Thursdays at 1:00 PM UTC via
[Google Meet](https://meet.google.com/wsc-jjsy-eih), and the meeting agenda is
maintained
[here](https://docs.google.com/document/d/1HwhAKqpM6l4Ur3h3IApDFzbH2Y_xvj_n1x1pEdwRuSY/edit?usp=sharing).
If you have a topic you wish to discuss at this meeting, please feel free to add
a discussion topic to the agenda.

Please file bug reports on
[Github](https://github.com/redhat-developer/service-binding-operator/issues/new).
For any other questions, reach out on
[service-binding-support@redhat.com](https://www.redhat.com/mailman/listinfo/service-binding-support).

Join the
[service-binding-operator](https://app.slack.com/client/T09NY5SBT/C019LQYGC5C)
channel in the [Kubernetes Workspace](https://slack.k8s.io/) for any discussions
and collaboration with the community.

[kafka]: https://developers.redhat.com/articles/2021/07/27/connect-nodejs-applications-red-hat-openshift-streams-apache-kafka-service#prerequisites
