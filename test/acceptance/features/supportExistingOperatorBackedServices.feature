@olm
@supported-operator
@disable.arch.ppc64le
@disable.arch.s390x
@disable.arch.arm64
Feature: Support a number of existing operator-backed services out of the box

  As a user of Service Binding operator
  I would like to be able to bind my application to a number of existing operator-backed services
  without a need to tweak their k8s resources

  Background:
    Given Namespace [TEST_NAMESPACE] is used
    * Service Binding Operator is running

  Scenario: Bind test application to Redis instance provisioned by Opstree Redis operator
    Given Opstree Redis operator is running
    * Generic test application is running
    * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: redis-secret
            stringData:
                password: redisSecret!
            """

    * The Custom Resource is present
          """
          apiVersion: redis.redis.opstreelabs.in/v1beta1
          kind: Redis
          metadata:
            name: redis-standalone
          spec:
            kubernetesConfig:
              image: quay.io/opstree/redis:v6.2.5
              imagePullPolicy: IfNotPresent
              resources:
                requests:
                  cpu: 101m
                  memory: 128Mi
                limits:
                  cpu: 101m
                  memory: 128Mi
              serviceType: ClusterIP
              redisSecret:
                name: redis-secret
                key: password
            storage:
              volumeClaimTemplate:
                spec:
                  # storageClassName: standard
                  accessModes: ["ReadWriteOnce"]
                  resources:
                    requests:
                      storage: 1Gi
            redisExporter:
              enabled: false
              image: quay.io/opstree/redis-exporter:1.0
          """
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: redis.redis.opstreelabs.in
                version: v1beta1
                kind: Redis
                name: redis-standalone
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Kind Redis with apiVersion redis.redis.opstreelabs.in/v1beta1 is listed in bindable kinds
    And Content of file "/bindings/$scenario_id/type" in application pod is
           """
           redis
           """
    And Content of file "/bindings/$scenario_id/host" in application pod is
           """
           redis-standalone
           """
    And Content of file "/bindings/$scenario_id/password" in application pod is
           """
           redisSecret!
           """

  @external-feedback
  Scenario: Bind test application to Postgres provisioned by Crunchy Data Postgres operator
    Given Crunchy Data Postgres operator is running
    * Generic test application is running
    * The Custom Resource is present
          """
          apiVersion: postgres-operator.crunchydata.com/v1beta1
          kind: PostgresCluster
          metadata:
            name: hippo
          spec:
            image: registry.developers.crunchydata.com/crunchydata/crunchy-postgres:centos8-13.4-1
            postgresVersion: 13
            instances:
              - name: instance1
                dataVolumeClaimSpec:
                  accessModes:
                  - "ReadWriteOnce"
                  resources:
                    requests:
                      storage: 1Gi
            backups:
              pgbackrest:
                image: registry.developers.crunchydata.com/crunchydata/crunchy-pgbackrest:centos8-2.35-0
                repos:
                - name: repo1
                  volume:
                    volumeClaimSpec:
                      accessModes:
                      - "ReadWriteOnce"
                      resources:
                        requests:
                          storage: 1Gi
          """
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: postgres-operator.crunchydata.com
                version: v1beta1
                kind: PostgresCluster
                name: hippo
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Kind PostgresCluster with apiVersion postgres-operator.crunchydata.com/v1beta1 is listed in bindable kinds
    And Content of file "/bindings/$scenario_id/type" in application pod is
           """
           postgresql
           """
    And Content of file "/bindings/$scenario_id/provider" in application pod is
           """
           crunchydata
           """
    And Content of file "/bindings/$scenario_id/host" in application pod is
           """
           hippo-primary.$NAMESPACE.svc
           """
    And Content of file "/bindings/$scenario_id/database" in application pod is
           """
           hippo
           """
    And Content of file "/bindings/$scenario_id/username" in application pod is
           """
           hippo
           """
    And File "/bindings/$scenario_id/password" exists in application pod
    And File "/bindings/$scenario_id/ca.crt" exists in application pod
    And File "/bindings/$scenario_id/tls.crt" exists in application pod
    And File "/bindings/$scenario_id/tls.key" exists in application pod
    And Application can connect to the projected Postgres database

  @disable-github-actions
  @disable-openshift-4.11
  @disable-openshift-4.12
  Scenario: Bind test application to Mysql provisioned by Percona Mysql operator and connect
    Given Percona Mysql operator is running
    * Generic test application is running
    * The Custom Resource is present
          """
          apiVersion: pxc.percona.com/v1
          kind: PerconaXtraDBCluster
          metadata:
            name: minimal-cluster
          spec:
            crVersion: 1.11.0
            secretsName: minimal-cluster-secrets
            allowUnsafeConfigurations: true
            upgradeOptions:
              apply: 8.0-recommended
              schedule: "0 4 * * *"
            pxc:
              size: 1
              image: percona/percona-xtradb-cluster:8.0.27-18.1
              volumeSpec:
                persistentVolumeClaim:
                  resources:
                    requests:
                      storage: 6G
            haproxy:
              enabled: true
              size: 1
              image: percona/percona-xtradb-cluster-operator:1.11.0-haproxy
            logcollector:
              enabled: true
              image: percona/percona-xtradb-cluster-operator:1.11.0-logcollector
          """
    * Condition ready=True for PerconaXtraDBCluster/minimal-cluster resource is met
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: pxc.percona.com
                version: v1
                kind: PerconaXtraDBCluster
                name: minimal-cluster
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Kind PerconaXtraDBCluster with apiVersion pxc.percona.com/v1 is listed in bindable kinds
    And Content of file "/bindings/$scenario_id/type" in application pod is
           """
           mysql
           """
    And Content of file "/bindings/$scenario_id/database" in application pod is
           """
           mysql
           """|
    And Content of file "/bindings/$scenario_id/host" in application pod is
           """
           minimal-cluster-haproxy.$NAMESPACE
           """
    And Content of file "/bindings/$scenario_id/port" in application pod is
           """
           3306
           """
    And Content of file "/bindings/$scenario_id/username" in application pod is
           """
           root
           """
    And File "/bindings/$scenario_id/password" exists in application pod
    And Application can connect to the projected MySQL database

  Scenario: Bind test application to Mysql provisioned by Percona Mysql operator
    Given Percona Mysql operator is running
    * Generic test application is running
    * The Custom Resource is present
          """
          apiVersion: pxc.percona.com/v1
          kind: PerconaXtraDBCluster
          metadata:
            name: minimal-cluster
          spec:
            crVersion: 1.11.0
            secretsName: minimal-cluster-secrets
            allowUnsafeConfigurations: true
            upgradeOptions:
              apply: 8.0-recommended
              schedule: "0 4 * * *"
            pxc:
              size: 1
              image: percona/percona-xtradb-cluster:8.0.27-18.1
              volumeSpec:
                persistentVolumeClaim:
                  resources:
                    requests:
                      storage: 6G
            haproxy:
              enabled: true
              size: 1
              image: percona/percona-xtradb-cluster-operator:1.11.0-haproxy
            logcollector:
              enabled: true
              image: percona/percona-xtradb-cluster-operator:1.11.0-logcollector
          """
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: pxc.percona.com
                version: v1
                kind: PerconaXtraDBCluster
                name: minimal-cluster
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Kind PerconaXtraDBCluster with apiVersion pxc.percona.com/v1 is listed in bindable kinds
    And Content of file "/bindings/$scenario_id/type" in application pod is
           """
           mysql
           """
    And Content of file "/bindings/$scenario_id/database" in application pod is
           """
           mysql
           """|
    And Content of file "/bindings/$scenario_id/host" in application pod is
           """
           minimal-cluster-haproxy.$NAMESPACE
           """
    And Content of file "/bindings/$scenario_id/port" in application pod is
           """
           3306
           """
    And Content of file "/bindings/$scenario_id/username" in application pod is
           """
           root
           """
    And File "/bindings/$scenario_id/password" exists in application pod

  @external-feedback
  Scenario: Bind test application to Postgres instance provisioned by Cloud Native Postgres operator
    Given Cloud Native Postgres operator is running
    * Generic test application is running
    * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
              name: cluster-example-app-user
            type: kubernetes.io/basic-auth
            stringData:
              password: secret
              username: guest
            """
    * The Custom Resource is present
            """
            apiVersion: postgresql.k8s.enterprisedb.io/v1
            kind: Cluster
            metadata:
              name: postgres
            spec:
              instances: 1
              primaryUpdateStrategy: unsupervised
              storage:
                size: 1Gi
            """
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: postgresql.k8s.enterprisedb.io
                version: v1
                kind: Cluster
                name: postgres
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Kind Cluster with apiVersion postgresql.k8s.enterprisedb.io/v1 is listed in bindable kinds
    And Content of file "/bindings/$scenario_id/type" in application pod is
           """
           postgresql
           """
    And Content of file "/bindings/$scenario_id/provider" in application pod is
           """
           enterprisedb
           """
    And Content of file "/bindings/$scenario_id/host" in application pod is
           """
           postgres-rw
           """
    And Content of file "/bindings/$scenario_id/username" in application pod is
           """
           app
           """
    And Content of file "/bindings/$scenario_id/database" in application pod is
           """
           app
           """
    And File "/bindings/$scenario_id/password" exists in application pod
    And Application can connect to the projected Postgres database

  Scenario: Bind test application to RabbitMQ instance provisioned by RabbitMq operator
    Given RabbitMQ operator is running
    * Generic test application is running
    * The Custom Resource is present
          """
          apiVersion: rabbitmq.com/v1beta1
          kind: RabbitmqCluster
          metadata:
            name: hello-world
          """
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: rabbitmq.com
                version: v1beta1
                kind: RabbitmqCluster
                name: hello-world
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Kind RabbitmqCluster with apiVersion rabbitmq.com/v1beta1 is listed in bindable kinds
    And Content of file "/bindings/$scenario_id/type" in application pod is
           """
           rabbitmq
           """
    And Content of file "/bindings/$scenario_id/host" in application pod is
           """
           hello-world.$NAMESPACE.svc
           """
    And Content of file "/bindings/$scenario_id/port" in application pod is
           """
           5672
           """
    And File "/bindings/$scenario_id/username" exists in application pod
           """
           root
           """
    And File "/bindings/$scenario_id/password" exists in application pod

  @crdv1beta1
  Scenario: Bind test application to MongoDB provisioned by Percona's MongoDB operator
    Given Percona MongoDB operator is running
    * Generic test application is running
    * The Custom Resource is present
          """
          apiVersion: psmdb.percona.com/v1-9-0
          kind: PerconaServerMongoDB
          metadata:
              name: mongo-cluster
          spec:
              crVersion: 1.9.0
              image: percona/percona-server-mongodb:4.4.8-9
              allowUnsafeConfigurations: true
              upgradeOptions:
                  apply: 4.4-recommended
                  schedule: "0 2 * * *"
              secrets:
                  users: mongo-cluster-secrets
              replsets:
                  - name: rs0
                    size: 1
                    volumeSpec:
                        persistentVolumeClaim:
                            resources:
                                requests:
                                    storage: 1Gi
              sharding:
                  enabled: false
          """
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: psmdb.percona.com
                version: v1-9-0
                kind: PerconaServerMongoDB
                name: mongo-cluster
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Kind PerconaServerMongoDB with apiVersion psmdb.percona.com/v1-9-0 is listed in bindable kinds
    And Content of file "/bindings/$scenario_id/type" in application pod is
           """
           mongodb
           """
    And Content of file "/bindings/$scenario_id/username" in application pod is
           """
           userAdmin
           """
    And File "/bindings/$scenario_id/password" exists in application pod
    And Content of file "/bindings/$scenario_id/host" in application pod is
            """
            mongo-cluster-rs0.$NAMESPACE.svc.cluster.local
            """
