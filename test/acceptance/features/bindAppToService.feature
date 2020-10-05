Feature: Bind an application to a service

    As a user of Service Binding Operator
    I want to bind applications to services it depends on

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
        And Secret "binding-request-a-d-s" contains "DATABASE_USER" key with value "postgres"
        And Secret "binding-request-a-d-s" contains "DATABASE_PASSWORD" key with value "password"
        And Secret "binding-request-a-d-s" contains "DATABASE_DB_PASSWORD" key with value "password"
        And Secret "binding-request-a-d-s" contains "DATABASE_DB_NAME" key with value "db-demo-a-d-s"
        And Secret "binding-request-a-d-s" contains "DATABASE_DB_PORT" key with value "5432"
        And Secret "binding-request-a-d-s" contains "DATABASE_DB_USER" key with value "postgres"
        And Secret "binding-request-a-d-s" contains "DATABASE_DB_HOST" key with dynamic IP addess as the value
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
        Given OLM Operator "backend" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo-empty-app
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.common
                username: foo
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-empty-app
            spec:
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-empty-app
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-empty-app" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-empty-app" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "binding-request-empty-app" should be changed to "EmptyApplication"
        And Secret "binding-request-empty-app" contains "BACKEND_HOST" key with value "example.common"
        And Secret "binding-request-empty-app" contains "BACKEND_USERNAME" key with value "foo"


    Scenario: Backend Service status update gets propagated to the binding secret
        Given OLM Operator "backend" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/ready: path={.status.ready}
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
        * jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-backend" should be changed to "True"
        * jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-backend" should be changed to "False"
        * Secret "binding-request-backend" contains "CustomReady" key with value "<no value>"
        # Backend status in "backend-demo" is updated
        When The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/ready: path={.status.ready}
            spec:
                host: example.common
            status:
                ready: true
            """
        Then Secret "binding-request-backend" contains "CustomReady" key with value "true"


    Scenario: Backend Service new spec status update gets propagated to the binding secret
        Given OLM Operator "backend-new-spec" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo
            spec:
                host: example.common
                ports:
                    - protocol: tcp
                      port: 8080
                    - protocol: ftp
                      port: 22
            """
        * Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-backend-new-spec
            spec:
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-backend-new-spec" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-backend-new-spec" should be changed to "False"
        And Secret "binding-request-backend-new-spec" contains "BACKEND_HOST" key with value "example.common"
        And Secret "binding-request-backend-new-spec" contains "BACKEND_PORTS_FTP" key with value "22"
        And Secret "binding-request-backend-new-spec" contains "BACKEND_PORTS_TCP" key with value "8080"


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
                              - service.binding:username:sourceValue=username
                              - service.binding:password:sourceValue=password
                            - description: Name of the ConfigMap to hold the DB config
                              displayName: DB Config Map
                              path: configmap
                              x-descriptors:
                              - urn:alm:descriptor:io.kubernetes:ConfigMap
                              - service.binding:db_host:sourceValue=db_host
                              - service.binding:db_port:sourceValue=db_port
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
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_DB_HOST" key with value "172.72.2.0"
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_DB_PORT" key with value "3306"
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_PASSWORD" key with value "secret123"
        And Secret "sbr-csv-secret-cm-descriptors" contains "BACKSERV_USERNAME" key with value "admin"


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

    Scenario: Provide binding info through backing service CRD annotation
        Given The Custom Resource Definition is present
                """
                apiVersion: apiextensions.k8s.io/v1beta1
                kind: CustomResourceDefinition
                metadata:
                  name: backends.stable.example.com
                  annotations:
                      service.binding/host: path={.spec.host}
                      service.binding/ready: path={.status.ready}
                spec:
                  group: stable.example.com
                  versions:
                    - name: v1
                      served: true
                      storage: true
                  scope: Namespaced
                  names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                      - bk
                """
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo
            spec:
                host: example.common
            status:
                ready: true
            """
        * Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-backend-a
            spec:
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo
                    id: SBR
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-backend-a" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-backend-a" should be changed to "False"
        And Secret "binding-request-backend-a" contains "BACKEND_READY" key with value "true"
        And Secret "binding-request-backend-a" contains "BACKEND_HOST" key with value "example.common"

    Scenario: Each value in referred map from service resource gets injected into app as separate env variable
        Given OLM Operator "backend" is running
        And Generic test application "rsa-2-service" is running
        And The Custom Resource Definition is present
        """
        apiVersion: apiextensions.k8s.io/v1beta1
        kind: CustomResourceDefinition
        metadata:
            name: backends.stable.example.com
            annotations:
                service.binding/spec: path={.spec},elementType=map
        spec:
            group: stable.example.com
            versions:
              - name: v1
                served: true
                storage: true
            scope: Namespaced
            names:
                plural: backends
                singular: backend
                kind: Backend
                shortNames:
                  - bk
        """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: rsa-2-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                bootstrap:
                    - type: plain
                      url: myhost2.example.com
                      name: hostGroup1service.binding/
                    - type: tls
                      url: myhost1.example.com:9092,myhost2.example.com:9092
                      name: hostGroup2
                data:
                    dbConfiguration: database-config     # ConfigMap
                    dbCredentials: database-cred-Secret  # Secret
                    url: db.stage.ibm.com
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: rsa-2
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: rsa-2-service
                application:
                    name: rsa-2-service
                    group: apps
                    version: v1
                    resource: deployments

            """
        Then The application env var "BACKEND_SPEC_IMAGE" has value "docker.io/postgres"
        And The application env var "BACKEND_SPEC_IMAGENAME" has value "postgres"
        And The application env var "BACKEND_SPEC_DBNAME" has value "db-demo"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "rsa-2" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "rsa-2" should be changed to "True"

        @negative
        Scenario: Service Binding with empty services is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-empty-services
            spec:
                services:
            """
        Then Error message "invalid: spec.services: Invalid value: \"null\"" is thrown
        And Service Binding "binding-request-empty-services" is not persistent in the cluster

        @negative
        Scenario: Service Binding without gvk of services is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-without-gvk
            spec:
                services:
                -   name: backend-demo
            """
        Then Error message "spec.services.group: Required value" is thrown
        And Error message "spec.services.kind: Required value" is thrown
        And Error message "spec.services.version: Required value" is thrown
        And Service Binding "binding-request-without-gvk" is not persistent in the cluster

        @negative
        Scenario: Removing service from services field from existing serivce binding is not allowed
        Given Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-remove-service
            spec:
                services:
                -   group: service.example.com
                    version: v1
                    kind: Backserv
                    name: demo-backserv-cr-2
            """
        When Invalid Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-remove-service
            spec:
                services:
            """
        Then Error message "invalid: spec.services: Required value" is thrown
        And Service Binding "binding-request-remove-service" is not updated
