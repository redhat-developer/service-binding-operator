Feature: Bind an application to a service

    As a user of Service Binding Operator
    I want to bind applications to services it depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    @smoke
    Scenario: Bind an application to backend service in the following order: Application, Service and Binding
        Given Generic test application "gen-app-a-s-b" is running
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
                name: service-a-s-b
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
                name: service-binding-a-s-b
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-a-s-b
                application:
                    name: gen-app-a-s-b
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "service-binding-a-s-b" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    Scenario:  Bind an application to backend service in the following order: Application, Binding and Service
        Given Generic test application "gen-app-a-b-s" is running
        And Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: service-binding-a-b-s
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-a-b-s
                application:
                    name: gen-app-a-b-s
                    group: apps
                    version: v1
                    resource: deployments
            """
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
        When The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: service-a-b-s
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: backend-secret
            """
        Then Service Binding "service-binding-a-b-s" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an application to backend service in the following order: Service, Binding and Application
        Given CustomResourceDefinition backends.stable.example.com is available
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
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: service-s-b-a
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: backend-secret
            """
        And Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: service-binding-s-b-a
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-s-b-a
                application:
                    name: gen-app-s-b-a
                    group: apps
                    version: v1
                    resource: deployments
            """
        When Generic test application "gen-app-s-b-a" is running
        Then Service Binding "service-binding-s-b-a" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an application to backend service in the following order: Binding, Application and Service
        Given Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: service-binding-b-a-s
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-b-a-s
                application:
                    name: gen-app-b-a-s
                    group: apps
                    version: v1
                    resource: deployments
            """
        * Generic test application "gen-app-b-a-s" is running
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
        When The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: service-b-a-s
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: backend-secret
            """
        Then Service Binding "service-binding-b-a-s" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    # Currently disabled as not supported by SBO
    @disabled
    Scenario: Bind an application to backend service in the following order: Binding, Service and Application
        Given Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: service-binding-b-s-a
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-b-s-a
                application:
                    name: gen-app-b-s-a
                    group: apps
                    version: v1
                    resource: deployments
            """
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
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: service-b-s-a
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: backend-secret
            """
        When Generic test application "gen-app-b-s-a" is running
        Then Service Binding "service-binding-b-s-a" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    @negative
    Scenario: Attempt to bind a non existing application to a backend service
        Given CustomResourceDefinition backends.stable.example.com is available
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
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: service-missing-app
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
                name: service-binding-missing-app
            spec:
                bindAsFiles: false
                application:
                    name: gen-missing-app
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-missing-app
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "service-binding-missing-app" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "service-binding-missing-app" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "service-binding-missing-app" should be changed to "ApplicationNotFound"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "service-binding-missing-app" should be changed to "False"


    @negative
    Scenario: Service Binding without application selector
        Given CustomResourceDefinition backends.stable.example.com is available
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
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-empty-app
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-empty-app
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-empty-app" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-empty-app" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "binding-request-empty-app" should be changed to "EmptyApplication"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-request-empty-app" should be changed to "True"
        And Secret contains "BACKEND_HOST" key with value "example.common"
        And Secret contains "BACKEND_USERNAME" key with value "foo"

    @olm
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
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-backend-new-spec
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-backend-new-spec" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-backend-new-spec" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-request-backend-new-spec" should be changed to "True"
        And Secret contains "BACKEND_HOST" key with value "example.common"
        And Secret contains "BACKEND_PORTS_FTP" key with value "22"
        And Secret contains "BACKEND_PORTS_TCP" key with value "8080"

    Scenario: Custom environment variable is injected into the application under the declared name ignoring global and service env prefix
        Given Generic test application "gen-app-c-e" is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: service-c-e
                annotations:
                    service.binding/port: path={.data.port}
                    service.binding/host: path={.data.host}
            data:
                port: "8080"
                host: "127.0.0.1"
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: service-binding-c-e
            spec:
                bindAsFiles: false
                application:
                    name: gen-app-c-e
                    group: apps
                    version: v1
                    resource: deployments
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-c-e
                    id: backendSVC
                mappings:
                    - name: HOST_ADDR
                      value: '{{ .backendSVC.data.host }}:{{ .backendSVC.data.port }}'
            """
        Then Service Binding "service-binding-c-e" is ready
        And The application env var "HOST_ADDR" has value "127.0.0.1:8080"

    @olm
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
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: sbr-csv-secret-cm-descriptors
            spec:
                bindAsFiles: false
                services:
                -   group: service.example.com
                    version: v1
                    kind: Backserv
                    name: demo-backserv-cr-2
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "sbr-csv-secret-cm-descriptors" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "sbr-csv-secret-cm-descriptors" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "sbr-csv-secret-cm-descriptors" should be changed to "True"
        And Secret contains "BACKSERV_DB_HOST" key with value "172.72.2.0"
        And Secret contains "BACKSERV_DB_PORT" key with value "3306"
        And Secret contains "BACKSERV_PASSWORD" key with value "secret123"
        And Secret contains "BACKSERV_USERNAME" key with value "admin"


    # This test scenario is disabled until the issue is resolved: https://github.com/redhat-developer/service-binding-operator/issues/656
    @disabled
    @olm
    Scenario: Create binding secret using specDescriptors definitions managed in OLM operator descriptors
        Given Backend service CSV is installed
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
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
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: sbr-csv-attribute
            spec:
                bindAsFiles: false
                services:
                -   group: service.example.com
                    version: v1
                    kind: Backserv
                    name: demo-backserv-cr-1
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "sbr-csv-attribute" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "sbr-csv-attribute" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "sbr-csv-attribute" should be changed to "True"
        And Secret contains "BACKSERV_ENV_SVCNAME" key with value "demo-backserv-cr-1"

    @examples
    Scenario: Bind an imported Node.js application to Etcd database
        Given Etcd operator running
        * Etcd cluster "etcd-cluster-example" is running
        * Nodejs application "node-todo-git" imported from "quay.io/pmacik/node-todo" image is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
              name: binding-request-etcd
            spec:
              bindAsFiles: false
              namingStrategy: "ETCDCLUSTER_{{ .name | upper }}"
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
        Then Service Binding "binding-request-etcd" is ready
        And Application endpoint "/api/todos" is available

    @negative
    Scenario: Service Binding with empty services is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-empty-services
            spec:
                services:
            """
        Then Error message is thrown
        And Service Binding "binding-request-empty-services" is not persistent in the cluster

    @negative
    Scenario: Service Binding without gvk of services is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-without-gvk
            spec:
                services:
                -   name: backend-demo
            """
        Then Error message is thrown

    @negative
    Scenario: Removing service from services field from existing serivce binding is not allowed
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: demo-backserv-cr-2
            """
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-remove-service
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: demo-backserv-cr-2
            """
        * jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-remove-service" should be changed to "True"
        * jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-remove-service" should be changed to "False"
        * jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "binding-request-remove-service" should be changed to "EmptyApplication"
        * jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-request-remove-service" should be changed to "True"
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-remove-service
            spec:
                services:
            """
        Then Error message is thrown
        And Service Binding "binding-request-remove-service" is not updated

    @negative
    Scenario: Service Binding without spec is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-without-spec
            """
        Then Error message is thrown
        And Service Binding "binding-request-without-spec" is not persistent in the cluster

    @negative
    Scenario: Service Binding with empty spec is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-empty-spec
            spec:
            """
        Then Error message is thrown
        And Service Binding "binding-request-empty-spec" is not persistent in the cluster

    @negative
    # Adding olm tag due to flakiness of this test on non-olm ci
    # This tests are also run on openshift and k8s with olm CI so no harm in skipping on non-olm CI run
    @olm
    Scenario: Emptying spec of existing service binding is not allowed
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: demo-backserv-cr-2
            """
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-emptying-spec
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: demo-backserv-cr-2
            """
        * jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-emptying-spec" should be changed to "True"
        * jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-emptying-spec" should be changed to "False"
        * jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "binding-request-emptying-spec" should be changed to "EmptyApplication"
        * jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-request-emptying-spec" should be changed to "True"
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-emptying-spec
            spec:
            """
        Then Error message is thrown
        And Service Binding "binding-request-emptying-spec" is not updated

    @negative
    # Adding olm tag due to flakiness of this test on non-olm ci
    # This tests are also run on openshift and k8s with olm CI so no harm in skipping non-olm CI run
    @olm
    Scenario: Removing spec of existing service binding is not allowed
        Given CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: demo-backserv-cr-2
            """
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-remove-spec
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: demo-backserv-cr-2
            """
        * jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-remove-spec" should be changed to "True"
        * jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-remove-spec" should be changed to "False"
        * jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "binding-request-remove-spec" should be changed to "EmptyApplication"
        * jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-request-remove-spec" should be changed to "True"
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-remove-spec
            """
        Then Error message is thrown
        And Service Binding "binding-request-remove-spec" is not updated

    Scenario: Bind an application to a service present in a different namespace
        Given Namespace is present
            """
            apiVersion: v1
            kind: Namespace
            metadata:
                name: backend-services
            """
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: backend-cross-ns-service
                namespace: backend-services
                annotations:
                    service.binding/host_cross_ns_service: path={.spec.host_cross_ns_service}
            spec:
                host_cross_ns_service: cross.ns.service.stable.example.com
            """
        * Generic test application "myapp-in-sbr-ns" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-cross-ns-service
            spec:
                bindAsFiles: false
                application:
                    name: myapp-in-sbr-ns
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-cross-ns-service
                    namespace: backend-services
            """
        Then Service Binding "binding-request-cross-ns-service" is ready
        And The application env var "BACKEND_HOST_CROSS_NS_SERVICE" has value "cross.ns.service.stable.example.com"

    Scenario: Inject all configmap keys into application
        Given The ConfigMap is present
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: example
                annotations:
                    service.binding: path={.data},elementType=map
            data:
                word: "hello"
            """
        * Generic test application "myapp-cm" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-configmap
            spec:
                services:
                -   group: ""
                    version: v1
                    kind: ConfigMap
                    name: example
                application:
                    name: myapp-cm
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "binding-request-configmap" is ready
        And The application env var "CONFIGMAP_DATA_WORD" has value "hello"


    Scenario: Inject all secret keys into application
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: example
                annotations:
                    service.binding: path={.data},elementType=map
            data:
                word: "aGVsbG8="
            """
        * Generic test application "myapp-secret" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-secret
            spec:
                services:
                -   group: ""
                    version: v1
                    kind: Secret
                    name: example
                application:
                    name: myapp-secret
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "binding-request-secret" is ready
        And The application env var "SECRET_DATA_WORD" has value "aGVsbG8="
