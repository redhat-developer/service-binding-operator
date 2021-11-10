Feature: Inject custom env variable into application

    As a user of Service Binding Operator
    I want to inject into application context an env variable
    whose value might be generated from values available in service resources

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * CustomResourceDefinition backends.stable.example.com is available

    Scenario: Sequence from service resource is injected into application using custom env variables without specifying annotations
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
        * Generic test application is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
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
        Then Service Binding is ready
        And The application env var "TAGS" has value "[centos7-12.3 123]"

    Scenario: Map from service resource is injected into application using custom env variables without specifying annotations
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
            spec:
                host: example.common
                userLabels:
                    archive: "false"
                    environment: "demo"
            """
        * Generic test application is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
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
                   - name: USER_LABELS
                     value: '{{ .backend.spec.userLabels }}'
            """
        Then Service Binding is ready
        And The application env var "USER_LABELS" has value "map[archive:false environment:demo]"

    Scenario: Scalar from service resource is injected into application using custom env variables without specifying annotations
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
            spec:
                host: example.common
                userLabels:
                    archive: "false"
                    environment: "demo"
            """
        * Generic test application is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
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
                   - name: USER_LABELS_ARCHIVE
                     value: '{{ .backend.spec.userLabels.archive }}'
            """
        Then Service Binding is ready
        And The application env var "USER_LABELS_ARCHIVE" has value "false"

