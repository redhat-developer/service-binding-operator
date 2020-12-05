@rbac
Feature: Forbid user to bind services to application
  if she/he does not have permissions to read binding data or
  modify application resource

  Background:
    Given Namespace [TEST_NAMESPACE] is used
    * No user has access to the namespace
    * Service Binding Operator is running
    * The custom resource is present
          """
          apiVersion: apiextensions.k8s.io/v1beta1
          kind: CustomResourceDefinition
          metadata:
            name: backends.stable.example.com
          spec:
            group: stable.example.com
            versions:
              - name: v1
                served: true
                storage: true
            scope: Namespaced
            names:
              plural: backends
              singular: backend
              kind: Backend
              shortNames:
                - bk
          """
    * Generic test application is running


  Scenario: Service cannot be bound to application if user cannot read service resource from another namespace
    Given Namespace "backend-ns" exists
    * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                namespace: backend-ns
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - 123
            """
    * User acceptance-tests-dev has 'edit' role in test namespace
    * User acceptance-tests-dev cannot read resource backends/backend in namespace backend-ns
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                    namespace: backend-ns
                    id: backend
                customEnvVar:
                   - name: TAGS
                     value: '{{ .backend.spec.tags }}'
            """
    Then Service Binding CollectionReady.status is "False"
    And Service Binding CollectionReady.reason is "NoServiceReadPermission"
    And The env var "TAGS" is not available to the application

  Scenario: Service cannot be bound to application if user cannot read service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - 123
            """
    * User acceptance-tests-dev has 'edit' role in test namespace
    * User acceptance-tests-dev cannot read resource backends/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                    id: backend
                customEnvVar:
                   - name: TAGS
                     value: '{{ .backend.spec.tags }}'
            """
    Then Service Binding CollectionReady.status is "False"
    And Service Binding CollectionReady.reason is "NoServiceReadPermission"
    And The env var "TAGS" is not available to the application

  Scenario: Service cannot be bound to application if user cannot read secret referred from service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/username: path={.status.data},objectType=Secret,valueKey=username
            spec:
                host: example.common
            status:
                data: $scenario_id
            """
    * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: acmeuser

            """
    * User acceptance-tests-dev has 'service-binding-edit' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    * User acceptance-tests-dev cannot read resource secrets/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                    id: backend
            """
    Then Service Binding CollectionReady.status is "False"
    And Service Binding CollectionReady.reason is "NoSecretReadPermission"
    And The env var "BACKEND_USERNAME" is not available to the application

  Scenario: Service cannot be bound to application if user cannot read config referred from service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/username: path={.status.data},objectType=ConfigMap,valueKey=username
            spec:
                host: example.common
            status:
                data: $scenario_id
            """
    * The ConfigMap is present
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: $scenario_id
            data:
                username: acmeuser

            """
    * User acceptance-tests-dev has 'service-binding-edit' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    * User acceptance-tests-dev cannot read resource configmaps/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                    id: backend
            """
    Then Service Binding CollectionReady.status is "False"
    And Service Binding CollectionReady.reason is "NoConfigMapReadPermission"
    And The env var "BACKEND_USERNAME" is not available to the application

  Scenario: Service cannot be bound to application if user cannot modify application resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - 123
            """
    * The Custom Resource is present
            """
            apiVersion: rbac.authorization.k8s.io/v1
            kind: ClusterRole
            metadata:
              name: backends-view
            rules:
              - apiGroups:
                  - stable.example.com
                resources:
                  - backends
                verbs:
                  - get
                  - list

            """
    * User acceptance-tests-dev has 'view' role in test namespace
    * User acceptance-tests-dev has 'service-binding-edit' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                    id: backend
                customEnvVar:
                   - name: TAGS
                     value: '{{ .backend.spec.tags }}'
            """
    Then Service Binding CollectionReady.status is "True"
    And Service Binding InjectionReady.status is "False"
    And Service Binding InjectionReady.reason is "NoAppModifyPermission"
    And The env var "TAGS" is not available to the application
