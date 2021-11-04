Feature: Bind an application to a service

    As a user of Service Binding Operator
    I want to bind applications to services it depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    @smoke
    Scenario: Bind an application to backend service in the following order: Application, Service and Binding
        Given Generic test application is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: $scenario_id
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    Scenario:  Bind an application to backend service in the following order: Application, Binding and Service
        Given Generic test application is running
        And Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
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
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        When The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: $scenario_id
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    Scenario: Bind an application to backend service in the following order: Service, Binding and Application
        Given CustomResourceDefinition backends.stable.example.com is available
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: $scenario_id
            """
        And Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        When Generic test application is running
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    Scenario: Bind an application to backend service in the following order: Binding, Application and Service
        Given Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        * Generic test application is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        When The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: $scenario_id
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    Scenario: Bind an application to backend service in the following order: Binding, Service and Application
        Given Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
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
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: $scenario_id
            """
        When Generic test application is running
        Then Service Binding "$scenario_id" is ready
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
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: $scenario_id
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "$scenario_id" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "$scenario_id" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "$scenario_id" should be changed to "ApplicationNotFound"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "$scenario_id" should be changed to "False"


    @negative
    Scenario: Service Binding without application selector
        Given CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.common
                username: foo
            """
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
            """
        Then Error message is thrown
        And Service Binding "$scenario_id" is not persistent in the cluster

    @negative
    Scenario: Cannot create Service Binding with name and label selector
        Given CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.common
                username: foo
            """
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                    labelSelector:
                      matchLabels:
                        name: backend-operator
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
            """
        Then Error message "name and selector MUST NOT be defined in the application reference" is thrown

    @spec
    Scenario: Cannot create Service Binding with name and label selector in the spec API
        Given CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.common
                username: foo
            """
        When Invalid Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                workload:
                  name: $scenario_id
                  apiVersion: apps/v1
                  kind: Deployment
                  selector:
                    matchLabels:
                      name: backend-operator
                service:
                  apiVersion: stable.example.com/v1
                  kind: Backend
                  name: $scenario_id
            """
        Then Error message "name and selector MUST NOT be defined in the application reference" is thrown

    @negative
    Scenario: Cannot update Service Binding with name and label selector
        Given CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
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
                name: $scenario_id-valid
            spec:
                bindAsFiles: false
                application:
                    group: apps
                    version: v1
                    resource: deployments
                    labelSelector:
                      matchLabels:
                        name: backend-operator
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
            """
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-invalid
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                    labelSelector:
                      matchLabels:
                        name: backend-operator
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
            """
        Then Error message "name and selector MUST NOT be defined in the application reference" is thrown

    @spec
    Scenario: Cannot update Service Binding with name and label selector in the spec API
        Given CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.common
                username: foo
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-valid
            spec:
                workload:
                  apiVersion: apps/v1
                  kind: Deployment
                  selector:
                    matchLabels:
                      name: backend-operator
                service:
                  apiVersion: stable.example.com/v1
                  kind: Backend
                  name: $scenario_id
            """
        When Invalid Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: binding-request-with-name-label-selector-spec2
            spec:
                workload:
                  name: $scenario_id-invalid
                  apiVersion: apps/v1
                  kind: Deployment
                  selector:
                    matchLabels:
                      name: backend-operator
                service:
                  apiVersion: stable.example.com/v1
                  kind: Backend
                  name: $scenario_id
            """
        Then Error message "name and selector MUST NOT be defined in the application reference" is thrown

    @olm
    Scenario: Bind service to application using binding definition available in x-descriptors
        Given OLM Operator "backend-new-spec" is running
        * Generic test application is running
        * The Custom Resource is present
            """
            apiVersion: "beta.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
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
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                -   group: beta.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_HOST" has value "example.common"
        And The application env var "BACKEND_PORTS_FTP" has value "22"
        And The application env var "BACKEND_PORTS_TCP" has value "8080"

    Scenario: Custom environment variable is injected into the application under the declared name ignoring global and service env prefix
        Given Generic test application is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
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
                name: $scenario_id
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                    id: backendSVC
                mappings:
                    - name: HOST_ADDR
                      value: '{{ .backendSVC.data.host }}:{{ .backendSVC.data.port }}'
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "HOST_ADDR" has value "127.0.0.1:8080"

    @negative
    Scenario: Service Binding with empty services is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                services:
            """
        Then Error message is thrown
        And Service Binding "$scenario_id" is not persistent in the cluster

    @negative
    Scenario: Service Binding without gvk of services is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                services:
                -   name: $scenario_id
            """
        Then Error message is thrown

    @negative
    Scenario: Removing service from services field from existing serivce binding is not allowed
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/name: path={.metadata.name}
            """
        * Generic test application is running
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        * Service Binding "$scenario_id" is ready
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                services:
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Error message is thrown
        And Service Binding "$scenario_id" is not updated

    @negative
    Scenario: Service Binding without spec is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            """
        Then Error message is thrown
        And Service Binding "$scenario_id" is not persistent in the cluster

    @negative
    Scenario: Service Binding with empty spec is not allowed in the cluster
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
            """
        Then Error message is thrown
        And Service Binding "$scenario_id" is not persistent in the cluster

    @negative
    # Adding olm tag due to flakiness of this test on non-olm ci
    # This tests are also run on openshift and k8s with olm CI so no harm in skipping on non-olm CI run
    @olm
    Scenario: Emptying spec of existing service binding is not allowed
        Given CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/name: path={.metadata.name}
            """
        * Generic test application is running
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        * Service Binding "$scenario_id" is ready
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Error message is thrown
        And Service Binding "$scenario_id" is not updated

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
                name: $scenario_id
                annotations:
                    service.binding/name: path={.metadata.name}
            """
        * Generic test application is running
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        * Service Binding "$scenario_id" is ready
        When Invalid Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            """
        Then Error message is thrown
        And Service Binding "$scenario_id" is not updated

    Scenario: Bind an application to a service present in a different namespace
        Given Namespace is present
            """
            apiVersion: v1
            kind: Namespace
            metadata:
                name: $scenario_id
            """
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
                namespace: $scenario_id
                annotations:
                    service.binding/host_cross_ns_service: path={.spec.host_cross_ns_service}
            spec:
                host_cross_ns_service: cross.ns.service.stable.example.com
            """
        * Generic test application is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                    namespace: $scenario_id
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_HOST_CROSS_NS_SERVICE" has value "cross.ns.service.stable.example.com"

    Scenario: Inject all configmap keys into application
        Given The ConfigMap is present
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: $scenario_id
                annotations:
                    service.binding: path={.data},elementType=map
            data:
                word: "hello"
            """
        * Generic test application is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                -   group: ""
                    version: v1
                    kind: ConfigMap
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "CONFIGMAP_WORD" has value "hello"


    Scenario: Inject all secret keys into application
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
                annotations:
                    service.binding: path={.data},elementType=map
            data:
                word: "aGVsbG8="
            """
        * Generic test application is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                services:
                -   group: ""
                    version: v1
                    kind: Secret
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "SECRET_WORD" has value "aGVsbG8="

    @negative
    Scenario: Do not bind as env if there is no binding data is collected from the service
        Given Generic test application is running
        * The Service is present
            """
            apiVersion: v1
            kind: Service
            metadata:
                name: $scenario_id-svc
            spec:
                selector:
                    name: $scenario_id-svc
                ports:
                  - port: 8080
                    targetPort: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: false
                detectBindingResources: true
                services:
                  - group: ""
                    version: v1
                    kind: Service
                    name: $scenario_id-svc
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding CollectionReady.status is "True"
        * Service Binding InjectionReady.status is "False"
        * Service Binding InjectionReady.reason is "NoBindingData"
