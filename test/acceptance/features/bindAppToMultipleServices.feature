Feature: Bind a single application to multiple services

    As a user of Service Binding operator
    I want to bind a single application to multiple services

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * CustomResourceDefinition backends.stable.example.com is available

    Scenario: Bind two backend services by creating 2 SBRs to a single application
        Given Generic test application "myapp-2-sbrs" is running
        * The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: myapp-2-sbrs-service-1
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
            name: myapp-2-sbrs-service-2
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
                name: binding1-myapp-2-sbrs
            spec:
                bindAsFiles: false
                application:
                    name: myapp-2-sbrs
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: myapp-2-sbrs-service-1
            """
        Then Service Binding "binding1-myapp-2-sbrs" is ready
        And "myapp-2-sbrs" deployment must contain reference to secret existing in service binding
        And The application env var "BACKEND_HOST" has value "foo"

        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding2-myapp-2-sbrs
            spec:
                bindAsFiles: false
                application:
                    name: myapp-2-sbrs
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: myapp-2-sbrs-service-2
            """
        Then Service Binding "binding2-myapp-2-sbrs" is ready
        And "myapp-2-sbrs" deployment must contain reference to secret existing in service binding
        And The application env var "BACKEND_HOST" has value "foo"
        And The application env var "BACKEND_PORT" has value "bar"
        And The application got redeployed 2 times so far
        And The application does not get redeployed again with 5 minutes

    Scenario: Bind two backend services by creating 1 SBR to a single application
        Given Generic test application "myapp-1sbr" is running
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: internal-db-1sbr
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
                name: external-db-1sbr
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
                name: binding-request-1sbr
            spec:
                bindAsFiles: false
                application:
                    name: myapp-1sbr
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: internal-db-1sbr
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: external-db-1sbr
            """
        Then Service Binding "binding-request-1sbr" is ready
        And The application env var "BACKEND_HOST_INTERNAL_DB" has value "internal.db.stable.example.com"
        And The application env var "BACKEND_HOST_EXTERNAL_DB" has value "external.db.stable.example.com"
