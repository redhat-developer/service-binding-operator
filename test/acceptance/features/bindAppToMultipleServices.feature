Feature: Bind a single application to multiple services

    As a user of Service Binding operator
    I want to bind a single application to multiple services

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed

    Scenario: Bind two backend services by creating 2 SBRs to a single application
        Given Imported Nodejs application "nodejs-app" is running
        * DB "db-demo-1" is running
        * DB "db-demo-2" is running
        * Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-1
            spec:
                applicationSelector:
                    resourceRef: nodejs-app
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-1
            """
        When Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-2
            spec:
                applicationSelector:
                    resourceRef: nodejs-app
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-2
            """
        Then "nodejs-app" deployment must contain SBR name "binding-request-1" and "binding-request-2"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-1" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-1" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-2" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-2" should be changed to "True"
