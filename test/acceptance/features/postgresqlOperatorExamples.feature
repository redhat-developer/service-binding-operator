@examples
Feature: Bind an application to postgresql database
    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed

    @smoke
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Application, DB and Service Binding
        Given Imported Nodejs application "nodejs-rest-http-crud-a-d-s" is running
        * DB "db-demo-a-d-s" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
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
        Then Service Binding "binding-request-a-d-s" is ready
        And application should be re-deployed
        And application should be connected to the DB "db-demo-a-d-s"
        And Secret contains "DATABASE_DBNAME" key with value "db-demo-a-d-s"
        And Secret contains "DATABASE_USER" key with value "postgres"
        And Secret contains "DATABASE_PASSWORD" key with value "password"
        And Secret contains "DATABASE_DB_PASSWORD" key with value "password"
        And Secret contains "DATABASE_DB_NAME" key with value "db-demo-a-d-s"
        And Secret contains "DATABASE_DB_PORT" key with value "5432"
        And Secret contains "DATABASE_DB_USER" key with value "postgres"
        And Secret contains "DATABASE_DB_HOST" key with dynamic IP addess as the value
        And Secret contains "DATABASE_DBCONNECTIONIP" key with dynamic IP addess as the value
        And Secret contains "DATABASE_DBCONNECTIONPORT" key with value "5432"

    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Application, Service Binding and DB
        Given Imported Nodejs application "nodejs-rest-http-crud-a-s-d" is running
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
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
        Then Service Binding "binding-request-a-s-d" is ready
        And application should be re-deployed
        And application should be connected to the DB "db-demo-a-s-d"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: DB, Service Binding and Application
        Given DB "db-demo-d-s-a" is running
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
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
        Then Service Binding "binding-request-d-s-a" is ready
        And application should be re-deployed
        And application should be connected to the DB "db-demo-d-s-a"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Service Binding, Application and DB
        Given Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
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
        Then Service Binding "binding-request-s-a-d" is ready
        And application should be re-deployed
        And application should be connected to the DB "db-demo-s-a-d"


    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an imported Node.js application to PostgreSQL database in the following order: Service Binding, DB and Application
        Given Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
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
        Then Service Binding "binding-request-s-d-a" is ready
        And application should be re-deployed
        And application should be connected to the DB "db-demo-s-d-a"

    @negative
    Scenario: Attempt to bind a non existing application to PostgreSQL database
        Given DB "db-demo-missing-app" is running
        * Imported Nodejs application "nodejs-missing-app" is not running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
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
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-request-missing-app" should be changed to "False"

    Scenario: Custom environment variable is injected into the application under the declared name ignoring global and service env prefix
        Given Imported Nodejs application "nodejs-rest-http-crud-a-d-c" is running
        * DB "db-demo-a-d-c" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-a-d-c
            spec:
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
                mappings:
                    - name: SOME_KEY
                      value: 'SOME_VALUE:{{ .postgresDB.status.dbConnectionPort }}:{{ .postgresDB.status.dbName }}'
            """
        Then Service Binding "binding-request-a-d-c" is ready
        And Secret contains "SOME_KEY" key with value "SOME_VALUE:5432:db-demo-a-d-c"

    # https://github.com/redhat-developer/service-binding-operator/tree/master/examples/pod_spec_path
    Scenario: Specify container's path in Service Binding
        Given DB "db-demo-csp" is running
        * The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
            kind: CustomResourceDefinition
            metadata:
                name: appconfigs.stable.example.com
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                scope: Namespaced
                names:
                    plural: appconfigs
                    singular: appconfig
                    kind: AppConfig
                    shortNames:
                    - ac
            """
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: demo-appconfig-csp
            spec:
                uri: "some uri"
                Command: "some command"
                image: my-image
                spec:
                    containers:
                    - name: hello-world
                      # Image from dockerhub, This is the import path for the Go binary to build and run.
                      image: yusufkaratoprak/kubernetes-gosample:latest
                      ports:
                      - containerPort: 8090
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-csp
            spec:
                application:
                    name: demo-appconfig-csp
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                    bindingPath:
                        containersPath: spec.spec.containers
                services:
                  - group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-csp
                    id: zzz
            """
        Then Service Binding "binding-request-csp" is ready
        And Secret has been injected in to CR "demo-appconfig-csp" of kind "AppConfig" at path "{.spec.spec.containers[0].envFrom[0].secretRef.name}"

    Scenario: Specify secret's path in Service Binding
        Given DB "db-demo-ssp" is running
        * The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
            kind: CustomResourceDefinition
            metadata:
                name: appconfigs.stable.example.com
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                scope: Namespaced
                names:
                    plural: appconfigs
                    singular: appconfig
                    kind: AppConfig
                    shortNames:
                    - ac
            """
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: demo-appconfig-ssp
            spec:
                spec:
                    secret: some-value
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-ssp
            spec:
                application:
                    name: demo-appconfig-ssp
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                    bindingPath:
                        secretPath: spec.spec.secret
                services:
                  - group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-ssp
                    id: zzz
            """
        Then Service Binding "binding-request-ssp" is ready
        And Secret has been injected in to CR "demo-appconfig-ssp" of kind "AppConfig" at path "{.spec.spec.secret}"
