Feature: Bind multiple applications to a single service

    As a user of Service Binding Operator
    I want to bind multiple applications to a single service that depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * OLM Operator "custom_app" is running

    Scenario: Successfully bind two applications to a single service
        Given Test applications "gen-app-a-s-f-1" and "gen-app-a-s-f-2" is running
        * The common label "app-custom=test" is set for both apps
        * CustomResourceDefinition backends.stable.example.com is available
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
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
                    dbCredentials: $scenario_id-secret
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

    Scenario: Bind applications via label selector after the service binding has been created
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig-1
                labels:
                    app-custom: $scenario_id
            spec:
                spec:
                    containers:
                        - name: foo
            """
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "*"
                    containers:
                      - path: .spec.spec.containers[*]
                    volumes: .spec.spec.volumes
            """
        And Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                services:
                  - group: ""
                    version: v1
                    kind: Secret
                    name: $scenario_id-secret
                application:
                    labelSelector:
                      matchLabels:
                        app-custom: $scenario_id
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
            """
        And Service Binding is ready
        And jsonpath "{.spec.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig-1" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.spec.volumes}" on "appconfigs/$scenario_id-appconfig-1" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"
        When 2 minutes have passed
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig-2
                labels:
                    app-custom: $scenario_id
            spec:
                spec:
                    containers:
                        - name: bar
            """
        Then jsonpath "{.spec.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig-2" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.spec.volumes}" on "appconfigs/$scenario_id-appconfig-2" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"

    Scenario: Bind applications via label selector as environment variables after the service binding has been created
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig-1
                labels:
                    app-custom: $scenario_id
            spec:
                spec:
                    containers:
                        - name: foo
            """
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "*"
                    containers:
                      - path: .spec.spec.containers[*]
                    volumes: .spec.spec.volumes
            """
        And Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                services:
                  - group: ""
                    version: v1
                    kind: Secret
                    name: $scenario_id-secret
                application:
                    labelSelector:
                      matchLabels:
                        app-custom: $scenario_id
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
            """
        And Service Binding is ready
        And Secret has been injected in to CR "$scenario_id-appconfig-1" of kind "appconfigs.stable.example.com" at path "{.spec.spec.containers[0].envFrom[0].secretRef.name}"
        When 2 minutes have passed
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig-2
                labels:
                    app-custom: $scenario_id
            spec:
                spec:
                    containers:
                        - name: bar
            """
        Then Secret has been injected in to CR "$scenario_id-appconfig-2" of kind "appconfigs.stable.example.com" at path "{.spec.spec.containers[0].envFrom[0].secretRef.name}"

    @spec
    Scenario: Bind applications via label selector after the spec service binding has been created
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: AzureDiamond
                password: hunter2
                type:     secret
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig-1
                labels:
                    app-custom: $scenario_id
            spec:
                spec:
                    containers:
                        - name: foo
            """
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "*"
                    containers:
                      - path: .spec.spec.containers[*]
            """
        And Service Binding is applied
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                    apiVersion: v1
                    kind: Secret
                    name: $scenario_id-secret
                workload:
                    selector:
                        matchLabels:
                            app-custom: $scenario_id
                    apiVersion: stable.example.com/v1
                    kind: AppConfig
            """
        And Service Binding is ready
        And jsonpath "{.spec.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig-1" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.template.spec.volumes}" on "appconfigs/$scenario_id-appconfig-1" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"
        When 2 minutes have passed
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig-2
                labels:
                    app-custom: $scenario_id
            spec:
                spec:
                    containers:
                        - name: bar
            """
        Then jsonpath "{.spec.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig-2" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.template.spec.volumes}" on "appconfigs/$scenario_id-appconfig-2" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"
