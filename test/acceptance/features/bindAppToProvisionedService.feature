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

  Scenario: Bind application to provisioned service
    Given Generic test application "myaop-provision-srv" is running
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: bind-provisioned-service-1
          spec:
              services:
              - group: stable.example.com
                version: v1
                kind: ProvisionedBackend
                name: provisioned-service-1
              application:
                name: myaop-provision-srv
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding "bind-provisioned-service-1" is ready
    And jq ".status.secret" of Service Binding "bind-provisioned-service-1" should be changed to "provisioned-secret-1"
    And Content of file "/bindings/bind-provisioned-service-1/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/bind-provisioned-service-1/password" in application pod is
            """
            bar
            """

  @openshift
  Scenario: Bind provisioned service to application deployed as deployment config
    Given Generic test application is running as deployment config
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
                name: provisioned-service-1
              application:
                name: $scenario_id
                group: apps.openshift.io
                version: v1
                resource: deploymentconfigs
          """
    Then Service Binding is ready
    And jq ".status.secret" of Service Binding should be changed to "provisioned-secret-1"
    And Content of file "/bindings/$scenario_id/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/$scenario_id/password" in application pod is
            """
            bar
            """


  @negative
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
                name: provisioned-service-2
            spec:
                foo: bar
            """
    * Generic test application "myaop-provision-srv2" is running
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: bind-provisioned-service-2
          spec:
              services:
              - group: stable.example.com
                version: v1
                kind: ProvisionedBackend
                name: provisioned-service-2
              application:
                name: myaop-provision-srv2
                group: apps
                version: v1
                resource: deployments
          """
    Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "bind-provisioned-service-2" should be changed to "False"
    And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "bind-provisioned-service-2" should be changed to "False"
    And jq ".status.conditions[] | select(.type=="CollectionReady").reason" of Service Binding "bind-provisioned-service-2" should be changed to "ErrorReadingBinding"

  Scenario: Bind application to provisioned service that has binding annotations as well
    Given OLM Operator "provisioned_backend_with_annotations" is running
    * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: ProvisionedBackend
            metadata:
                name: provisioned-service-3
                annotations:
                    "service.binding/foo": "path={.spec.foo}"
            spec:
                foo: bla
            status:
                binding:
                    name: provisioned-secret-1
            """
    * Generic test application "myaop-provision-srv3" is running
    When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: bind-provisioned-service-3
          spec:
              services:
              - group: stable.example.com
                version: v1
                kind: ProvisionedBackend
                name: provisioned-service-3
              application:
                name: myaop-provision-srv3
                group: apps
                version: v1
                resource: deployments
          """
    Then Service Binding "bind-provisioned-service-3" is ready
    And Content of file "/bindings/bind-provisioned-service-3/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/bind-provisioned-service-3/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/bind-provisioned-service-3/foo" in application pod is
            """
            bla
            """

  @spec
  @smoke
  Scenario: SPEC Bind application to provisioned service
    Given Generic test application "spec-myapp-provision-srv" is running
    When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: spec-bind-provisioned-service-1
          spec:
              service:
                apiVersion: stable.example.com/v1
                kind: ProvisionedBackend
                name: provisioned-service-2
              workload:
                name: spec-myapp-provision-srv
                apiVersion: apps/v1
                kind: Deployment
          """
    Then Service Binding "spec-bind-provisioned-service-1" is ready
    And jq ".status.binding.name" of Service Binding "spec-bind-provisioned-service-1" should be changed to "provisioned-secret-2"
    And Content of file "/bindings/spec-bind-provisioned-service-1/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/spec-bind-provisioned-service-1/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/spec-bind-provisioned-service-1/type" in application pod is
            """
            db
            """

  @spec
  Scenario: SPEC Bind application to provisioned service and inject type/provider from values set on service binding
    Given Generic test application "spec-myapp-provision-srv4" is running
    When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: spec-bind-provisioned-service-4
          spec:
              type: mysql
              provider: foovendor
              service:
                apiVersion: stable.example.com/v1
                kind: ProvisionedBackend
                name: provisioned-service-1
              workload:
                name: spec-myapp-provision-srv4
                apiVersion: apps/v1
                kind: Deployment
          """
    Then Service Binding "spec-bind-provisioned-service-4" is ready
    And Content of file "/bindings/spec-bind-provisioned-service-4/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/spec-bind-provisioned-service-4/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/spec-bind-provisioned-service-4/type" in application pod is
            """
            mysql
            """
    And Content of file "/bindings/spec-bind-provisioned-service-4/provider" in application pod is
            """
            foovendor
            """

  @spec
  Scenario: SPEC Bind application to provisioned service and inject binding into folder specified by .spec.name
    Given Generic test application "spec-myapp-provision-srv3" is running
    When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: spec-bind-provisioned-service-3
          spec:
              name: foo-bindings
              service:
                apiVersion: stable.example.com/v1
                kind: ProvisionedBackend
                name: provisioned-service-2
              workload:
                name: spec-myapp-provision-srv3
                apiVersion: apps/v1
                kind: Deployment
          """
    Then Service Binding "spec-bind-provisioned-service-3" is ready
    And jq ".status.binding.name" of Service Binding "spec-bind-provisioned-service-3" should be changed to "provisioned-secret-2"
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
              name: $scenario_id
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
  Scenario: SPEC Inject specified bindings as env vars
    Given Generic test application "spec-myapp-provision-srv8" is running
    When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: spec-bind-provisioned-service-8
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
                name: spec-myapp-provision-srv8
                apiVersion: apps/v1
                kind: Deployment
          """
    Then Service Binding "spec-bind-provisioned-service-8" is ready
    And jq ".status.binding.name" of Service Binding "spec-bind-provisioned-service-8" should be changed to "provisioned-secret-2"
    And Content of file "/bindings/spec-bind-provisioned-service-8/username" in application pod is
            """
            foo
            """
    And Content of file "/bindings/spec-bind-provisioned-service-8/password" in application pod is
            """
            bar
            """
    And Content of file "/bindings/spec-bind-provisioned-service-8/type" in application pod is
            """
            db
            """
    And The application env var "FOO" has value "foo"
    And The application env var "BAR" has value "bar"
