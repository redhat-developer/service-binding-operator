@olm
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

  @crdv1beta1
  Scenario: Bind test application to Mysql provisioned by Percona Mysql operator
    Given Percona Mysql operator is running
    * Generic test application is running
    * The Custom Resource is present
          """
          apiVersion: pxc.percona.com/v1-8-0
          kind: PerconaXtraDBCluster
          metadata:
            name: minimal-cluster
          spec:
            crVersion: 1.8.0
            secretsName: minimal-cluster-secrets
            allowUnsafeConfigurations: true
            upgradeOptions:
              apply: 8.0-recommended
              schedule: "0 4 * * *"
            pxc:
              size: 1
              image: percona/percona-xtradb-cluster:8.0.23-14.1
              volumeSpec:
                persistentVolumeClaim:
                  resources:
                    requests:
                      storage: 6G
            haproxy:
              enabled: true
              size: 1
              image: perconalab/percona-xtradb-cluster-operator:main-haproxy
            logcollector:
              enabled: true
              image: perconalab/percona-xtradb-cluster-operator:main-logcollector
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
                version: v1-8-0
                kind: PerconaXtraDBCluster
                name: minimal-cluster
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Kind PerconaXtraDBCluster with apiVersion pxc.percona.com/v1-8-0 is listed in bindable kinds
    And Content of file "/bindings/$scenario_id/type" in application pod is
           """
           mysql
           """
    And Content of file "/bindings/$scenario_id/host" in application pod is
           """
           minimal-cluster-haproxy.$NAMESPACE
           """
    And Content of file "/bindings/$scenario_id/username" in application pod is
           """
           root
           """
    And File "/bindings/$scenario_id/password" exists in application pod
