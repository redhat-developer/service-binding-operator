@rbac
Feature: Prevent users to bind services to application
  if she/he does not have permissions to read binding data or
  modify application resource

  Background:
    Given Namespace [TEST_NAMESPACE] is used
    * No user has access to the namespace
    * Service Binding Operator is running
    * CustomResourceDefinition backends.stable.example.com is available
    * Generic test application is running


  Scenario: Service cannot be bound to application if user cannot read service resource from another namespace
    Given Namespace "$scenario_id-ns" exists
    * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                namespace: $scenario_id-ns
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - "123"
            """
    * User acceptance-tests-dev has 'edit' role in test namespace
    * User acceptance-tests-dev cannot read resource backends/backend in namespace backend-ns
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
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
                    name: $scenario_id-backend
                    namespace: $scenario_id-ns
                    id: backend
                mappings:
                   - name: TAGS
                     value: '{{ .backend.spec.tags }}'
            """
    Then Service Binding CollectionReady.status is "False"

  Scenario: Service cannot be bound to application if user cannot read service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - "123"
            """
    * User acceptance-tests-dev has 'edit' role in test namespace
    * User acceptance-tests-dev cannot read resource backends/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
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
                    name: $scenario_id-backend
                    id: backend
                mappings:
                   - name: TAGS
                     value: '{{ .backend.spec.tags }}'
            """
    Then Service Binding CollectionReady.status is "False"

  Scenario: Service cannot be bound to application if user cannot read secret referred from service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/username: path={.status.bindings},objectType=Secret,valueKey=username
            spec:
                host: example.common
            status:
                bindings: $scenario_id-secret
            """
    * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: acmeuser

            """
    * User acceptance-tests-dev has 'service-binding-editor-role' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    * User acceptance-tests-dev cannot read resource secrets/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
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
                    name: $scenario_id-backend
                    id: backend
            """
    Then Service Binding CollectionReady.status is "False"

  Scenario: Service cannot be bound to application if user cannot read config referred from service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/username: path={.status.bindings},objectType=ConfigMap,valueKey=username
            spec:
                host: example.common
            status:
                bindings: $scenario_id
            """
    * The ConfigMap is present
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: $scenario_id-secret
            data:
                username: acmeuser
            """
    * User acceptance-tests-dev has 'service-binding-editor-role' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    * User acceptance-tests-dev cannot read resource configmaps/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
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
                    name: $scenario_id-backend
                    id: backend
            """
    Then Service Binding CollectionReady.status is "False"

  Scenario: Service cannot be bound to application if user cannot modify application resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - "123"
            """
    * User acceptance-tests-dev has 'view' role in test namespace
    * User acceptance-tests-dev has 'service-binding-editor-role' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
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
                    name: $scenario_id-backend
                    id: backend
                mappings:
                   - name: TAGS
                     value: '{{ .backend.spec.tags }}'
            """
    Then Service Binding CollectionReady.status is "True"
    And Service Binding InjectionReady.status is "False"

  @spec
  Scenario: SPEC Service cannot be bound to application if user cannot read service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - "123"
            """
    * User acceptance-tests-dev has 'edit' role in test namespace
    * User acceptance-tests-dev cannot read resource backends/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                type: foo
                workload:
                    name: $scenario_id
                    apiVersion: apps/v1
                    kind: Deployment
                service:
                    apiVersion: stable.example.com/v1
                    kind: Backend
                    name: $scenario_id-backend
            """
    Then Service Binding CollectionReady.status is "False"

  @spec
  Scenario: SPEC Service cannot be bound to application if user cannot read secret referred from service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/username: path={.status.bindings},objectType=Secret,valueKey=username
            spec:
                host: example.common
            status:
                bindings: $scenario_id-secret
            """
    * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: acmeuser

            """
    * User acceptance-tests-dev has 'service-binding-editor-role' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    * User acceptance-tests-dev cannot read resource secrets/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                type: foo
                workload:
                    name: $scenario_id
                    apiVersion: apps/v1
                    kind: Deployment
                service:
                    apiVersion: stable.example.com/v1
                    kind: Backend
                    name: $scenario_id-backend
            """
    Then Service Binding CollectionReady.status is "False"

  @spec
  Scenario: SPEC Service cannot be bound to application if user cannot read config referred from service resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/username: path={.status.bindings},objectType=ConfigMap,valueKey=username
            spec:
                host: example.common
            status:
                bindings: $scenario_id-secret
            """
    * The ConfigMap is present
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: $scenario_id-secret
            data:
                username: acmeuser
            """
    * User acceptance-tests-dev has 'service-binding-editor-role' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    * User acceptance-tests-dev cannot read resource configmaps/$scenario_id in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                type: foo
                workload:
                    name: $scenario_id
                    apiVersion: apps/v1
                    kind: Deployment
                service:
                    apiVersion: stable.example.com/v1
                    kind: Backend
                    name: $scenario_id-backend
            """
    Then Service Binding CollectionReady.status is "False"

  @spec
  Scenario: SPEC Service cannot be bound to application if user cannot modify application resource
    Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - "123"
            """
    * User acceptance-tests-dev has 'view' role in test namespace
    * User acceptance-tests-dev has 'service-binding-editor-role' role in test namespace
    * User acceptance-tests-dev has 'backends-view' role in test namespace
    When user acceptance-tests-dev applies Service Binding
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                type: foo
                workload:
                    name: $scenario_id
                    apiVersion: apps/v1
                    kind: Deployment
                service:
                    apiVersion: stable.example.com/v1
                    kind: Backend
                    name: $scenario_id-backend
            """
    Then Service Binding CollectionReady.status is "True"
    And Service Binding InjectionReady.status is "False"
