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