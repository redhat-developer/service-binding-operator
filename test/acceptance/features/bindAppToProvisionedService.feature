Feature: Bind application to provisioned service

  As a user I would like to bind my applications to provisioned services, as defined by the binding spec

  Background:
    Given Namespace [TEST_NAMESPACE] is used
    * Service Binding Operator is running
    * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: provisioned-secret-1
            stringData:
                username: foo
                password: bar
            """
    * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: provisioned-secret-2
            stringData:
                username: foo
                password: bar
                type: db
            """
    * OLM Operator "provisioned_backend" is running
    * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: ProvisionedBackend
            metadata:
                name: provisioned-service-1
            spec:
                foo: bar
            status:
                binding:
                    name: provisioned-secret-1
            """
    * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: ProvisionedBackend
            metadata:
                name: provisioned-service-2
            spec:
                foo: bar
            status:
                binding:
                    name: provisioned-secret-2
            """

  @external-feedback
  Scenario: Bind application to provisioned service
    Given Generic test application is running
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              services:
              - group: stable.example.com
                version: v1
                kind: ProvisionedBackend
                name: provisioned-service-1
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And jq ".status.secret" of Service Binding should be changed to "provisioned-secret-1"
    And Content of file "/bindings/$scenario_id-binding/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/$scenario_id-binding/password" in application pod is
            """
            bar
            """

  @openshift
  @external-feedback
  Scenario: Bind provisioned service to application deployed as deployment config
    Given Generic test application is running as deployment config
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              services:
              - group: stable.example.com
                version: v1
                kind: ProvisionedBackend
                name: provisioned-service-1
              application:
                name: $scenario_id
                group: apps.openshift.io
                version: v1
                resource: deploymentconfigs
          """
    Then Service Binding is ready
    And jq ".status.secret" of Service Binding "$scenario_id-binding" should be changed to "provisioned-secret-1"
    And Content of file "/bindings/$scenario_id-binding/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/$scenario_id-binding/password" in application pod is
            """
            bar
            """


  @negative
  @external-feedback
  Scenario: Fail binding to provisioned service if secret name is not provided
    Given The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: provisionedbackends.stable.example.com
                annotations:
                    "servicebinding.io/provisioned-service": "true"
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                    schema:
                        openAPIV3Schema:
                            type: object
                            properties:
                                apiVersion:
                                    type: string
                                kind:
                                    type: string
                                metadata:
                                    type: object
                                spec:
                                    type: object
                                    properties:
                                        foo:
                                            type: string
                scope: Namespaced
                names:
                    plural: provisionedbackends
                    singular: provisionedbackend
                    kind: ProvisionedBackend
                    shortNames:
                      - pbk
            """
    * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: ProvisionedBackend
            metadata:
                name: $scenario_id-provisioned-backend
            spec:
                foo: bar
            """
    * Generic test application is running
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              services:
              - group: stable.example.com
                version: v1
                kind: ProvisionedBackend
                name: $scenario_id-provisioned-backend
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding should be changed to "False"
    And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding should be changed to "False"
    And jq ".status.conditions[] | select(.type=="CollectionReady").reason" of Service Binding should be changed to "ErrorReadingBinding"

  @negative
  @external-feedback
  Scenario: Fail binding to provisioned service if secret name is provided but the secret does not exist
    * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: ProvisionedBackend
            metadata:
                name: $scenario_id
            spec:
                foo: bar
            status:
                binding:
                    name: provisioned-secret-imaginary
            """
    * Generic test application is running
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: stable.example.com
                version: v1
                kind: ProvisionedBackend
                name: $scenario_id
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding should be changed to "False"
    And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding should be changed to "False"
    And jq ".status.conditions[] | select(.type=="CollectionReady").reason" of Service Binding should be changed to "ErrorReadingSecret"

  @external-feedback
  Scenario: Bind application to provisioned service that has binding annotations as well
    Given OLM Operator "provisioned_backend_with_annotations" is running
    * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: ProvisionedBackend
            metadata:
                name: $scenario_id-provisioned-backend
                annotations:
                    "service.binding/foo": "path={.spec.foo}"
            spec:
                foo: bla
            status:
                binding:
                    name: provisioned-secret-1
            """
    * Generic test application is running
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              services:
              - group: stable.example.com
                version: v1
                kind: ProvisionedBackend
                name: $scenario_id-provisioned-backend
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding is ready
    And Content of file "/bindings/$scenario_id-binding/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/$scenario_id-binding/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/$scenario_id-binding/foo" in application pod is
            """
            bla
            """

  @spec
  @smoke
  @external-feedback
  Scenario: SPEC Bind application to provisioned service
    Given Generic test application is running
    When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              service:
                apiVersion: stable.example.com/v1
                kind: ProvisionedBackend
                name: provisioned-service-2
              workload:
                name: $scenario_id
                apiVersion: apps/v1
                kind: Deployment
          """
    Then Service Binding is ready
    And jq ".status.binding.name" of Service Binding "$scenario_id-binding" should be changed to "provisioned-secret-2"
    And Content of file "/bindings/$scenario_id-binding/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/$scenario_id-binding/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/$scenario_id-binding/type" in application pod is
            """
            db
            """

  @spec
  @external-feedback
  Scenario: SPEC Bind application to provisioned service and inject type/provider from values set on service binding
    Given Generic test application is running
    When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              type: mysql
              provider: foovendor
              service:
                apiVersion: stable.example.com/v1
                kind: ProvisionedBackend
                name: provisioned-service-1
              workload:
                name: $scenario_id
                apiVersion: apps/v1
                kind: Deployment
          """
    Then Service Binding is ready
    And Content of file "/bindings/$scenario_id-binding/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/$scenario_id-binding/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/$scenario_id-binding/type" in application pod is
            """
            mysql
            """
    And Content of file "/bindings/$scenario_id-binding/provider" in application pod is
            """
            foovendor
            """

  @spec
  @external-feedback
  Scenario: SPEC Bind application to provisioned service and inject binding into folder specified by .spec.name
    Given Generic test application is running
    When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              name: foo-bindings
              service:
                apiVersion: stable.example.com/v1
                kind: ProvisionedBackend
                name: provisioned-service-2
              workload:
                name: $scenario_id
                apiVersion: apps/v1
                kind: Deployment
          """
    Then Service Binding is ready
    And jq ".status.binding.name" of Service Binding "$scenario_id-binding" should be changed to "provisioned-secret-2"
    And Content of file "/bindings/foo-bindings/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/foo-bindings/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/foo-bindings/type" in application pod is
            """
            db
            """

  Scenario: Bind application to provisioned service and inject binding into folder specified by .spec.name
    Given Generic test application is running
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              name: foo
              services:
                - group: stable.example.com
                  version: v1
                  kind: ProvisionedBackend
                  name: provisioned-service-2
              application:
                name: $scenario_id
                group: apps
                version: v1
                kind: Deployment
          """
    Then Service Binding is ready
    And Content of file "/bindings/foo/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/foo/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/foo/type" in application pod is
            """
            db
            """


  @spec
  @external-feedback
  Scenario: SPEC Inject specified bindings as env vars
    Given Generic test application is running
    When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              env:
                - name: "FOO"
                  key: username
                - name: "BAR"
                  key: password
              service:
                apiVersion: stable.example.com/v1
                kind: ProvisionedBackend
                name: provisioned-service-2
              workload:
                name: $scenario_id
                apiVersion: apps/v1
                kind: Deployment
          """
    Then Service Binding is ready
    And jq ".status.binding.name" of Service Binding "$scenario_id-binding" should be changed to "provisioned-secret-2"
    And Content of file "/bindings/$scenario_id-binding/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/$scenario_id-binding/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/$scenario_id-binding/type" in application pod is
            """
            db
            """
    And The application env var "FOO" has value "foo"
    And The application env var "BAR" has value "bar"
