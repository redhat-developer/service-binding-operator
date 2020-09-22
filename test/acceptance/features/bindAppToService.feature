Feature: Bind an application to a service

    As a user of Service Binding Operator
    I want to bind applications to services it depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed

    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Application, DB and Service Binding
        Given Imported Nodejs application "nodejs-rest-http-crud-a-d-s" is running
        * DB "db-demo-a-d-s" is running
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-a-d-s
            spec:
                application:
                    name: nodejs-rest-http-crud-a-d-s
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-a-d-s
            """

        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-a-d-s" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-a-d-s" should be changed to "True"
        And application should be re-deployed
        And application should be connected to the DB "db-demo-a-d-s"
        And Secret "binding-request-a-d-s" contains "DATABASE_DBNAME" key with value "db-demo-a-d-s"
        And Secret "binding-request-a-d-s" contains "DATABASE_SECRET_USER" key with value "postgres"
        And Secret "binding-request-a-d-s" contains "DATABASE_SECRET_PASSWORD" key with value "password"
        And Secret "binding-request-a-d-s" contains "DATABASE_CONFIGMAP_DB_PASSWORD" key with value "password"
        And Secret "binding-request-a-d-s" contains "DATABASE_CONFIGMAP_DB_NAME" key with value "db-demo-a-d-s"
        And Secret "binding-request-a-d-s" contains "DATABASE_CONFIGMAP_DB_PORT" key with value "5432"
        And Secret "binding-request-a-d-s" contains "DATABASE_CONFIGMAP_DB_USER" key with value "postgres"
        And Secret "binding-request-a-d-s" contains "DATABASE_CONFIGMAP_DB_HOST" key with dynamic IP addess as the value
        And Secret "binding-request-a-d-s" contains "DATABASE_DBCONNECTIONIP" key with dynamic IP addess as the value
        And Secret "binding-request-a-d-s" contains "DATABASE_DBCONNECTIONPORT" key with value "5432"

    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Application, Service Binding and DB
        Given Imported Nodejs application "nodejs-rest-http-crud-a-s-d" is running
        * Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-a-s-d
            spec:
                application:
                    name: nodejs-rest-http-crud-a-s-d
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-a-s-d
            """
        When DB "db-demo-a-s-d" is running
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-a-s-d" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-a-s-d" should be changed to "True"
        And application should be re-deployed
        And application should be connected to the DB "db-demo-a-s-d"


    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: DB, Application and Service Binding
        Given DB "db-demo-d-a-s" is running
        * Imported Nodejs application "nodejs-rest-http-crud-d-a-s" is running
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-d-a-s
            spec:
                application:
                    name: nodejs-rest-http-crud-d-a-s
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-d-a-s
            """

        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-d-a-s" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-d-a-s" should be changed to "True"
        And application should be re-deployed
        And application should be connected to the DB "db-demo-d-a-s"


    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: DB, Service Binding and Application
        Given DB "db-demo-d-s-a" is running
        * Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-d-s-a
            spec:
                application:
                    name: nodejs-rest-http-crud-d-s-a
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-d-s-a
            """
        When Imported Nodejs application "nodejs-rest-http-crud-d-s-a" is running
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-d-s-a" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-d-s-a" should be changed to "True"
        And application should be re-deployed
        And application should be connected to the DB "db-demo-d-s-a"


    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Service Binding, Application and DB
        Given Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-s-a-d
            spec:
                application:
                    name: nodejs-rest-http-crud-s-a-d
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-s-a-d
            """
        * Imported Nodejs application "nodejs-rest-http-crud-s-a-d" is running
        When DB "db-demo-s-a-d" is running
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-s-a-d" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-s-a-d" should be changed to "True"
        And application should be re-deployed
        And application should be connected to the DB "db-demo-s-a-d"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Service Binding, DB and Application
        Given Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-s-d-a
            spec:
                application:
                    name: nodejs-rest-http-crud-s-d-a
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-s-d-a
            """
        * DB "db-demo-s-d-a" is running
        When Imported Nodejs application "nodejs-rest-http-crud-s-d-a" is running
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-s-d-a" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-s-d-a" should be changed to "True"
        And application should be re-deployed
        And application should be connected to the DB "db-demo-s-d-a"


    @negative
    Scenario: Attempt to bind a non existing application to PostgreSQL database
        Given DB "db-demo-missing-app" is running
        * Imported Nodejs application "nodejs-missing-app" is not running
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-missing-app
            spec:
                application:
                    name: nodejs-missing-app
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-missing-app
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-missing-app" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-missing-app" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "binding-request-missing-app" should be changed to "ApplicationNotFound"

    @negative
    Scenario: Service Binding without application selector
        Given Imported Nodejs application "nodejs-empty-app" is running
        And DB "db-demo-empty-app" is running
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-empty-app
            spec:
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-empty-app
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-empty-app" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-empty-app" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "binding-request-empty-app" should be changed to "EmptyApplication"
        And Secret "binding-request-empty-app" contains "DATABASE_DBNAME" key with value "db-demo-empty-app"
        And Secret "binding-request-empty-app" contains "DATABASE_SECRET_USER" key with value "postgres"
        And Secret "binding-request-empty-app" contains "DATABASE_SECRET_PASSWORD" key with value "password"
        And Secret "binding-request-empty-app" contains "DATABASE_CONFIGMAP_DB_PASSWORD" key with value "password"
        And Secret "binding-request-empty-app" contains "DATABASE_CONFIGMAP_DB_NAME" key with value "db-demo-empty-app"
        And Secret "binding-request-empty-app" contains "DATABASE_CONFIGMAP_DB_PORT" key with value "5432"
        And Secret "binding-request-empty-app" contains "DATABASE_CONFIGMAP_DB_USER" key with value "postgres"
        And Secret "binding-request-empty-app" contains "DATABASE_CONFIGMAP_DB_HOST" key with dynamic IP addess as the value
        And Secret "binding-request-empty-app" contains "DATABASE_DBCONNECTIONIP" key with dynamic IP addess as the value
        And Secret "binding-request-empty-app" contains "DATABASE_DBCONNECTIONPORT" key with value "5432"



    Scenario: Backend Service status update gets propagated to the binding secret
        Given OLM Operator "backend" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo
                annotations:
                    servicebindingoperator.redhat.io/status.ready: 'binding:env:attribute'
                    servicebindingoperator.redhat.io/spec.host: 'binding:env:attribute'
            spec:
                host: example.common
            """
        * Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-backend
            spec:
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo
                    id: SBR
                customEnvVar:
                  - name: CustomReady
                    value: '{{ .SBR.status.ready }}'
                  - name: CustomHost
                    value: '{{ .SBR.spec.host }}'
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-backend" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-backend" should be changed to "False"
        Then Secret "binding-request-backend" contains "CustomReady" key with value "<no value>"
        # Backend status in "backend-demo" is updated
        When The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo
                annotations:
                    servicebindingoperator.redhat.io/status.ready: 'binding:env:attribute'
                    servicebindingoperator.redhat.io/spec.host: 'binding:env:attribute'
            spec:
                host: example.common
            status:
                ready: true
            """
        Then Secret "binding-request-backend" contains "CustomReady" key with value "true"

    Scenario: Custom environment variable is injected into the application under the declared name ignoring global and service env prefix
        Given Imported Nodejs application "nodejs-rest-http-crud-a-d-c" is running
        * DB "db-demo-a-d-c" is running
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-a-d-c
            spec:
                envVarPrefix: REDHAT
                application:
                    name: nodejs-rest-http-crud-a-d-c
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-a-d-c
                    id: postgresDB
                    envVarPrefix: DEVTOOLS
                customEnvVar:
                    - name: SOME_KEY
                      value: 'SOME_VALUE:{{ .postgresDB.status.dbConnectionPort }}:{{ .postgresDB.status.dbName }}'
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-a-d-c" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-a-d-c" should be changed to "True"
        And Secret "binding-request-a-d-c" contains "SOME_KEY" key with value "SOME_VALUE:5432:db-demo-a-d-c"

    Scenario: Creating binding secret from the definitions managed in OLM operator descriptors
        Given Backend service CSV is installed
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ClusterServiceVersion
            metadata:
                name: some-backend-service.v0.2.0
            spec:
                displayName: Some Backend Service
                install:
                    strategy: deployment
                customresourcedefinitions:
                    owned:
                        - name: backservs.service.example.com
                          version: v1
                          kind: Backserv
                          statusDescriptors:
                            - description: Name of the Secret to hold the DB user and password
                              displayName: DB Password Credentials
                              path: secret
                              x-descriptors:
                              - urn:alm:descriptor:io.kubernetes:Secret
                              - binding:env:object:secret:username
                              - binding:env:object:secret:password
                            - description: Name of the ConfigMap to hold the DB config
                              displayName: DB Config Map
                              path: configmap
                              x-descriptors:
                              - urn:alm:descriptor:io.kubernetes:ConfigMap
                              - binding:env:object:configmap:db_host
                              - binding:env:object:configmap:db_port
            """
        * The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
            kind: CustomResourceDefinition
            metadata:
                name: backservs.service.example.com
            spec:
                group: service.example.com
                versions:
                    - name: v1
                      served: true
                      storage: true
                scope: Namespaced
                names:
                    plural: backservs
                    singular: backserv
                    kind: Backserv
                    shortNames:
                    - bs
            """
        * The Custom Resource is present
            """
            apiVersion: service.example.com/v1
            kind: Backserv
            metadata:
                name: demo-backserv-cr-2
            status:
                secret: csv-demo-secret
                configmap: csv-demo-cm
            """
        * The ConfigMap is present
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: csv-demo-cm
            data:
                db_host: 172.72.2.0
                db_port: "3306"
            """
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: csv-demo-secret
            type: Opaque
            stringData:
                username: admin
                password: secret123
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: sbr-csv-secret-cm-descriptors
            spec:
                services:
                -   group: service.example.com
                    version: v1
                    kind: Backserv
                    name: demo-backserv-cr-2
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "sbr-csv-secret-cm-descriptors" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "sbr-csv-secret-cm-descriptors" should be changed to "False"
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_CONFIGMAP_DB_HOST" key with value "172.72.2.0"
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_CONFIGMAP_DB_PORT" key with value "3306"
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_SECRET_PASSWORD" key with value "secret123"
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_SECRET_USERNAME" key with value "admin"


    # This test scenario is disabled until the issue is resolved: https://github.com/redhat-developer/service-binding-operator/issues/656
    @disabled
    Scenario: Create binding secret using specDescriptors definitions managed in OLM operator descriptors
        Given Backend service CSV is installed
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ClusterServiceVersion
            metadata:
                name: some-backend-service.v0.1.0
            spec:
                displayName: Some Backend Service
                install:
                    strategy: deployment
                customresourcedefinitions:
                    owned:
                        - name: backservs.service.example.com
                          version: v1
                          kind: Backserv
                          specDescriptors:
                            - description: SVC name
                              displayName: SVC name
                              path: svcName
                              x-descriptors:
                                - binding:env:attribute

            """
        * The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
            kind: CustomResourceDefinition
            metadata:
                name: backservs.service.example.com
            spec:
                group: service.example.com
                versions:
                    - name: v1
                      served: true
                      storage: true
                scope: Namespaced
                names:
                    plural: backservs
                    singular: backserv
                    kind: Backserv
                    shortNames:
                    - bs
            """
        * The Custom Resource is present
            """
            apiVersion: service.example.com/v1
            kind: Backserv
            metadata:
                name: demo-backserv-cr-1
            spec:
                svcName: demo-backserv-cr-1
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: sbr-csv-attribute
            spec:
                services:
                -   group: service.example.com
                    version: v1
                    kind: Backserv
                    name: demo-backserv-cr-1
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "sbr-csv-attribute" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "sbr-csv-attribute" should be changed to "False"
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_ENV_SVCNAME" key with value "demo-backserv-cr-1"

    Scenario: Bind an imported Node.js application to Etcd database
        Given Etcd operator running
        * Etcd cluster "etcd-cluster-example" is running
        * Nodejs application "node-todo-git" imported from "quay.io/pmacik/node-todo" image is running
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
              name: binding-request-etcd
            spec:
              application:
                group: apps
                version: v1
                resource: deployments
                name: node-todo-git
              services:
                - group: etcd.database.coreos.com
                  version: v1beta2
                  kind: EtcdCluster
                  name: etcd-cluster-example
              detectBindingResources: true
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-etcd" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-etcd" should be changed to "True"
        And Application endpoint "/api/todos" is available
