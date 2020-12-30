@olm
Feature: Bind a single application to multiple SBRs

    As a user of Service Binding operator
    I want to bind a single application to multiple SBRs

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed

    Scenario: Bind a single db instance by creating 2 SBRs to a single application
        Given Imported Nodejs application "nodejs-app" is running
        * DB "db-demo-1" is running
        * Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-1
            spec:
                application:
                    name: nodejs-app
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-1
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-2
            spec:
                application:
                    name: nodejs-app
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-1
            """
        Then Service Binding "binding-request-1" is ready
        And Service Binding "binding-request-2" is ready
        And "nodejs-app" deployment must contain SBR name "binding-request-1"
        And "nodejs-app" deployment must contain SBR name "binding-request-2"
