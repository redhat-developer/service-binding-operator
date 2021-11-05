Feature: Bind a single application to multiple services

    As a user of Service Binding operator
    I want to bind a single application to multiple services

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * CustomResourceDefinition backends.stable.example.com is available

    Scenario: Bind two backend services by creating 2 SBRs to a single application
        Given Generic test application is running
        * The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: $scenario_id-backend-1
            annotations:
                service.binding/host: path={.spec.host}
        spec:
            host: foo
        """
        * The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: $scenario_id-backend-2
            annotations:
                service.binding/port: path={.spec.port}
        spec:
            port: bar
        """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding-1
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
                    name: $scenario_id-backend-1
            """
        Then Service Binding "$scenario_id-binding-1" is ready
        And The application env var "BACKEND_HOST" has value "foo"

        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding-2
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
                    name: $scenario_id-backend-2
            """
        Then Service Binding "$scenario_id-binding-2" is ready
        And The application env var "BACKEND_HOST" has value "foo"
        And The application env var "BACKEND_PORT" has value "bar"
        And The application got redeployed 2 times so far
        And The application does not get redeployed again with 5 minutes

    Scenario: Bind two backend services by creating 1 SBR to a single application
        Given Generic test application is running
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-internal
                annotations:
                    service.binding/host_internal_db: path={.spec.host_internal_db}
            spec:
                host_internal_db: internal.db.stable.example.com
            """
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-external
                annotations:
                    service.binding/host_external_db: path={.spec.host_external_db}
            spec:
                host_external_db: external.db.stable.example.com
            """
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
                mappings:
                - name: FOO
                  value: '{{ .db1.metadata.name }}_{{ .db2.metadata.name }}'
                - name: FOO2
                  value: '{{ .db1.metadata.name }}_{{ .db2.kind }}'
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-internal
                    id: db1
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-external
                    id: db2
            """
        Then Service Binding "$scenario_id-binding" is ready
        And The application env var "BACKEND_HOST_INTERNAL_DB" has value "internal.db.stable.example.com"
        And The application env var "BACKEND_HOST_EXTERNAL_DB" has value "external.db.stable.example.com"
        And The application env var "FOO" has value "$scenario_id-internal_$scenario_id-external"
        And The application env var "FOO2" has value "$scenario_id-internal_Backend"
