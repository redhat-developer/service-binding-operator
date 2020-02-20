# Service Binding Operator Usage

This document describes how the Service Binding Operator v1.0.0 would help the developer at
different development stages.

## Bootstrapping

The first example would be a `webshop` application, which requires only a PostgreSQL database.
The binding between both resources are as follow:

```yaml
---
apiVersion: apps.openshift.io/v1
kind: ServiceBinding
metadata:
  name: webshop
spec:
  applications:
  - group: apps
    version: v1
    resource: deployments
    name: webshop
  services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: webshop
```

The following environment variables would be available to the `webshop` deployment (assuming
`Database` expose the `connectionUrl` property):

```
DATABASE_WEBSHOP_CONNECTIONURL=...
```

## Adding another dependency

Building in our example, let's consider the webshop has users and some bottlenecks have been
identified, leading the developers to add a caching service in their infrastructure; a Redis instance
has been created and is available to be connected with an application.

The following manifest illustrates the new binding situation:

```yaml
---
apiVersion: apps.openshift.io/v1
kind: ServiceBinding
metadata:
  name: webshop
spec:
  applications:
  - group: apps
    version: v1
    resource: deployments
    name: webshop
  services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: webshop
  - group: redis.isutton.dev
    version: v1alpha1
    kind: Redis
    name: webshop
```

The following environment variables would be available to the `webshop` deployment (assuming
`Database` and `Redis` expose the `connectionUrl` property):

```
DATABASE_WEBSHOP_CONNECTIONURL=...
REDIS_WEBSHOP_CONNECTIONURL=...
```

## Breaking The Monolith

The development team has decided to split some functionality, isolating the checkout into its own
workload called `webshot-checkout`. Since it has the same dependencies from the `webshop` workload,
it can be added in the same service binding manifest as well as the new `webshop-checkout` database.

The new binding situation can be described as follows:

```yaml
---
apiVersion: apps.openshift.io/v1
kind: ServiceBinding
metadata:
  name: webshop
spec:
  applications:
  - group: apps
    version: v1
    resource: deployments
    name: webshop
  - group: apps
    version: v1
    resource: deployments
    name: webshop-checkout
  services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: webshop
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: webshop-checkout
  - group: redis.isutton.dev
    version: v1alpha1
    kind: Redis
    name: webshop
```

The following environment variables would be available to the `webshop` deployment (assuming
`Database` and `Redis` expose the `connectionUrl` property):

```
DATABASE_WEBSHOP_CONNECTIONURL=...
DATABASE_WEBSHOP_CHECKOUT_CONNECTIONURL=...
REDIS_WEBSHOP_CONNECTIONURL=...
```

## Separating The Environments

Once the developers have finished migrating the checkout to the new workload and database, this
binding can be split in two: one for the checkout functionality, and another for the rest of the
webshop application.

The new binding situation would be as follows, with the existing binding manifest:

```yaml
---
apiVersion: apps.openshift.io/v1
kind: ServiceBinding
metadata:
  name: webshop
spec:
  applications:
  - group: apps
    version: v1
    resource: deployments
    name: webshop
  services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: webshop
  - group: redis.isutton.dev
    version: v1alpha1
    kind: Redis
    name: webshop
```

And the new `webshop-checkout` service binding, which still depends on data available on the
`webshop` database:

```yaml
---
apiVersion: apps.openshift.io/v1
kind: ServiceBinding
metadata:
  name: webshop-checkout
spec:
  applications:
  - group: apps
    version: v1
    resource: deployments
    name: webshop-checkout
  services:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: webshop
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    name: webshop-checkout
```
