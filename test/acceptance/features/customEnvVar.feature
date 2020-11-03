Feature: Inject custom env variable into application

  As a user of Service Binding Operator
  I want to inject into application context an env variable
  whose value might be generated from values available in service resources

  Background:
    Given Namespace [TEST_NAMESPACE] is used
    * Service Binding Operator is running

  Scenario: Sequence from service resource is injected into application using custom env variables without specifying annotations
    Given OLM Operator "backend" is running
    * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-with-tag-sequence
            spec:
                host: example.common
                tags:
                    - "centos7-12.3"
                    - 123
            """
    * Generic test application "foo" is running
    When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: custom-env-var-from-sequence
            spec:
                application:
                    name: foo
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-with-tag-sequence
                    id: backend
                customEnvVar:
                   - name: TAGS
                     value: '{{ .backend.spec.tags }}'
            """
    Then The application env var "TAGS" has value "[centos7-12.3 123]"
    And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "custom-env-var-from-sequence" should be changed to "True"
    And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "custom-env-var-from-sequence" should be changed to "True"
    And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "custom-env-var-from-sequence" should be changed to "True"

  Scenario: Map from service resource is injected into application using custom env variables without specifying annotations
    Given OLM Operator "backend" is running
    * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-with-user-labels-map
            spec:
                host: example.common
                userLabels:
                    archive: "false"
                    environment: "demo"
            """
    * Generic test application "foo2" is running
    When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: custom-env-var-from-map
            spec:
                application:
                    name: foo2
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-with-user-labels-map
                    id: backend
                customEnvVar:
                   - name: USER_LABELS
                     value: '{{ .backend.spec.userLabels }}'
            """
    Then The application env var "USER_LABELS" has value "map[archive:false environment:demo]"
    And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "custom-env-var-from-map" should be changed to "True"
    And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "custom-env-var-from-map" should be changed to "True"
    And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "custom-env-var-from-map" should be changed to "True"


  Scenario: Scalar from service resource is injected into application using custom env variables without specifying annotations
    Given OLM Operator "backend" is running
    * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-with-user-labels-archive
            spec:
                host: example.common
                userLabels:
                    archive: "false"
                    environment: "demo"
            """
    * Generic test application "foo3" is running
    When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: custom-env-var-from-scalar
            spec:
                application:
                    name: foo3
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-with-user-labels-archive
                    id: backend
                customEnvVar:
                   - name: USER_LABELS_ARCHIVE
                     value: '{{ .backend.spec.userLabels.archive }}'
            """
    Then The application env var "USER_LABELS_ARCHIVE" has value "false"
    And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "custom-env-var-from-scalar" should be changed to "True"
    And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "custom-env-var-from-scalar" should be changed to "True"
    And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "custom-env-var-from-scalar" should be changed to "True"
