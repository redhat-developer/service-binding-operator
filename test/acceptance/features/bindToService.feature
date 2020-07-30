Feature: Bind an application to a service

    As a user of Service Binding Operator
    I want to bind applications to services it depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed

    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Application, DB and Service Binding Request
        Given Imported Nodejs application "nodejs-rest-http-crud-a-d-s" is running
        * DB "db-demo-a-d-s" is running
        When Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-a-d-s
            spec:
                applicationSelector:
                    resourceRef: nodejs-rest-http-crud-a-d-s
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-a-d-s
            """
        Then application should be re-deployed
        And application should be connected to the DB "db-demo-a-d-s"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-a-d-s" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-a-d-s" should be changed to "True"

    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Application, Service Binding Request and DB
        Given Imported Nodejs application "nodejs-rest-http-crud-a-s-d" is running
        * Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-a-s-d
            spec:
                applicationSelector:
                    resourceRef: nodejs-rest-http-crud-a-s-d
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-a-s-d
            """
        When DB "db-demo-a-s-d" is running
        Then application should be re-deployed
        And application should be connected to the DB "db-demo-a-s-d"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-a-s-d" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-a-s-d" should be changed to "True"

    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: DB, Application and Service Binding Request
        Given DB "db-demo-d-a-s" is running
        * Imported Nodejs application "nodejs-rest-http-crud-d-a-s" is running
        When Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-d-a-s
            spec:
                applicationSelector:
                    resourceRef: nodejs-rest-http-crud-d-a-s
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-d-a-s
            """
        Then application should be re-deployed
        And application should be connected to the DB "db-demo-d-a-s"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-d-a-s" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-d-a-s" should be changed to "True"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: DB, Service Binding Request and Application
        Given DB "db-demo-d-s-a" is running
        * Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-d-s-a
            spec:
                applicationSelector:
                    resourceRef: nodejs-rest-http-crud-d-s-a
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-d-s-a
            """
        When Imported Nodejs application "nodejs-rest-http-crud-d-s-a" is running
        Then application should be re-deployed
        And application should be connected to the DB "db-demo-d-s-a"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-d-s-a" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-d-s-a" should be changed to "True"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Service Binding Request, Application and DB
        Given Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-s-a-d
            spec:
                applicationSelector:
                    resourceRef: nodejs-rest-http-crud-s-a-d
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-s-a-d
            """
        * Imported Nodejs application "nodejs-rest-http-crud-s-a-d" is running
        When DB "db-demo-s-a-d" is running
        Then application should be re-deployed
        And application should be connected to the DB "db-demo-s-a-d"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-s-a-d" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-s-a-d" should be changed to "True"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Service Binding Request, DB and Application
        Given Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-s-d-a
            spec:
                applicationSelector:
                    resourceRef: nodejs-rest-http-crud-s-d-a
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-s-d-a
            """
        * DB "db-demo-s-d-a" is running
        When Imported Nodejs application "nodejs-rest-http-crud-s-d-a" is running
        Then application should be re-deployed
        And application should be connected to the DB "db-demo-s-d-a"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-s-d-a" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-s-d-a" should be changed to "True"

    @negative
    Scenario: Attempt to bind a non existing application to PostgreSQL database
        Given DB "db-demo-missing-app" is running
        * Imported Nodejs application "nodejs-missing-app" is not running
        When Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-missing-app
            spec:
                applicationSelector:
                    resourceRef: nodejs-missing-app
                    group: apps
                    version: v1
                    resource: deployments
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-missing-app
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-missing-app" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-missing-app" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding Request "binding-request-missing-app" should be changed to "ApplicationNotFound"

    @negative
    Scenario: Service Binding Request without application selector
        Given Imported Nodejs application "nodejs-empty-app" is running
        And DB "db-demo-empty-app" is running
        When Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-empty-app
            spec:
                backingServiceSelector:
                    group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-empty-app
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-empty-app" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-empty-app" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding Request "binding-request-empty-app" should be changed to "EmptyApplicationSelector"

    Scenario: Bind two backend services to a single application
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
