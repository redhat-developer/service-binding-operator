Feature: Bind multiple applications to a single service

    As a user of Service Binding Operator
    I want to bind multiple applications to a single service that depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Successfully bind two applications to a single service
        Given Test applications "gen-app-a-s-f-1" and "gen-app-a-s-f-2" is running
        * The common label "app-custom=test" is set for both apps
        * CustomResourceDefinition backends.stable.example.com is available
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: backend-secret
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-service
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: backend-secret
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-service
                application:
                    labelSelector:
                      matchLabels:
                        app-custom: test
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond" in both apps
        And The application env var "BACKEND_PASSWORD" has value "hunter2" in both apps
