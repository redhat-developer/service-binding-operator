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
                    name: db-demo-2
            """

        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-1" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-1" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-2" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-2" should be changed to "True"
        And "nodejs-app" deployment must contain SBR name "binding-request-1"
        And "nodejs-app" deployment must contain SBR name "binding-request-2"

        Scenario: Bind two backend services by creating 1 SBR to a single application
        Given Generic test application "myapp-1sbr" is running
        * OLM Operator "backend" is running
        * The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: internal-db-1sbr
            annotations:
                servicebindingoperator.redhat.io/spec.host_internal_db: 'binding:env:attribute'
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
                servicebindingoperator.redhat.io/spec.host_external_db: 'binding:env:attribute'
        spec:
            host_external_db: external.db.stable.example.com
        """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-1sbr
            spec:
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

        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-1sbr" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-1sbr" should be changed to "True"
        And The application env var "BACKEND_HOST_INTERNAL_DB" has value "internal.db.stable.example.com"
        And The application env var "BACKEND_HOST_EXTERNAL_DB" has value "external.db.stable.example.com"
