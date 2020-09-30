Feature: Unbind an application from a service

    As a user of Service Binding Operator
    I want to unbind an application from a service

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Unbind a generic test application from the backing service
        Given OLM Operator "backend" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: example-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.com
                username: foo
            """
        * Generic test application "generic-app-a-d-u" is running
        * Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-a-d-u
            spec:
                application:
                    name: generic-app-a-d-u
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: example-backend
                    id: backend
            """
        * The application env var "BACKEND_HOST" has value "example.com"
        * The application env var "BACKEND_USERNAME" has value "foo"

        When Service binding "binding-request-a-d-u" is deleted

        Then The env var "BACKEND_HOST" is not available to the application
        And The env var "BACKEND_USERNAME" is not available to the application
