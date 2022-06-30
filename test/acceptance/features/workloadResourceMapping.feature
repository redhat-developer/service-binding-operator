@workload-resource-mapping
Feature: Bind services to workloads based on workload resource mapping

    As a user, I would like to be able to use workload resource bindings as defined by the specification.

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        And Service Binding Operator is running
        And CustomResourceDefinition backends.stable.example.com is available
        And OLM Operator "not_a_pod" is running
        And OLM Operator "custom_app" is running
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: notpodspecs.stable.example.com
            spec:
                versions:
                  - version: "*"
                    volumes: .spec.volumeData
                    containers:
                      - path: .spec.containerSpecs[*]
                        name: .id
                        env: .envData
                        volumeMounts: .volumeEntries
                      - path: .spec.initContainerSpecs[*]
                        name: .id
                        env: .envData
                        volumeMounts: .volumeEntries
            """

    Scenario: Mapping with coreos api service bindings
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: foo
                password: bar
                type: db
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: NotPodSpec
            metadata:
                name: $scenario_id-npc
            spec:
                containerSpecs:
                  - id: $scenario_id-id
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: true
                application:
                    name: $scenario_id-npc
                    group: stable.example.com
                    version: v1
                    resource: notpodspecs
                services:
                  - version: v1
                    group: ""
                    kind: Secret
                    name: $scenario_id-secret
            """
        Then Service Binding is ready
        And jq ".status.secret" of Service Binding should be changed to "$scenario_id-secret"
        And jsonpath "{.spec.containerSpecs[0].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"

    Scenario: Mapping should also bind according to bindingPath values
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: foo
                password: bar
                type: db
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig
            spec:
                template:
                    spec:
                        initContainers:
                          - name: foo
                        containers:
                          - name: bar
            """
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "*"
                    containers:
                      - path: .spec.template.spec.containers[*]
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: true
                application:
                    name: $scenario_id-appconfig
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                    bindingPath:
                        containersPath: spec.template.spec.initContainers
                services:
                  - version: v1
                    group: ""
                    kind: Secret
                    name: $scenario_id-secret
            """
        Then Service Binding is ready
        And jq ".status.secret" of Service Binding should be changed to "$scenario_id-secret"
        And jsonpath "{.spec.template.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.template.spec.initContainers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.template.spec.volumes}" on "appconfigs/$scenario_id-appconfig" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"

    Scenario: Mapping should allow bindAsFiles: false in podspec resources
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: foo
                password: bar
                type: db
            """
        And Generic test application is running
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: deployments.apps
            spec:
                versions:
                  - version: "*"
                    containers:
                      - path: .spec.template.spec.containers[*]
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                  - version: v1
                    group: ""
                    kind: Secret
                    name: $scenario_id-secret
            """
        Then Service Binding is ready
        And The application env var "SECRET_USERNAME" has value "foo"
        And The application env var "SECRET_PASSWORD" has value "bar"
        And The application env var "SECRET_TYPE" has value "db"

    Scenario: Mapping should allow bindAsFiles: false in non-podspec resources
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: foo
                password: bar
                type: db
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig
            spec:
                spec:
                    containers:
                    - name: bar
            """
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "*"
                    containers:
                      - path: .spec.spec.containers[*]
                        name: .name
                        env: .env
                        volumeMounts: .volumeMounts
                    volumes: .spec.template.spec.volumes
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id-appconfig
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                services:
                  - version: v1
                    group: ""
                    kind: Secret
                    name: $scenario_id-secret
            """
        Then Service Binding is ready
        And Secret has been injected in to CR "$scenario_id-appconfig" of kind "appconfigs.stable.example.com" at path "{.spec.spec.containers[0].envFrom[0].secretRef.name}"

    @spec
    Scenario: Deployment identity mappings should not change binding behavior
        Given Generic test application is running
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: deployments.apps
            spec:
                versions:
                  - version: "*"
                    containers:
                      - path: .spec.template.spec.containers[*]
                        name: .name
                        env: .env
                        volumeMounts: .volumeMounts
                      - path: .spec.template.spec.initContainers[*]
                        name: .name
                        env: .env
                        volumeMounts: .volumeMounts
                    volumes: .spec.template.spec.volumes
            """
        And The custom resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.common
                username: foo
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                type: mysql
                service:
                    apiVersion: stable.example.com/v1
                    kind: Backend
                    name: $scenario_id-backend
                workload:
                    name: $scenario_id
                    apiVersion: apps/v1
                    kind: Deployment
            """
        Then Service Binding is ready
        And The application env var "SERVICE_BINDING_ROOT" has value "/bindings"
        And Content of file "/bindings/$scenario_id-binding/host" in application pod is
            """
            example.common
            """
        And Content of file "/bindings/$scenario_id-binding/type" in application pod is
            """
            mysql
            """
        And The Workload Resource Mapping is deleted
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: deployments.apps
            """

    @spec
    Scenario: Specify container path using Workload Resource Mappings in custom resource
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
            kind: NotPodSpec
            metadata:
                name: $scenario_id-npc
            spec:
                containerSpecs:
                  - id: $scenario_id-id
                    image: scratch
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                    name: $scenario_id-secret
                    kind: Secret
                    apiVersion: v1
                workload:
                    apiVersion: stable.example.com/v1
                    kind: NotPodSpec
                    name: $scenario_id-npc
            """
        Then Service Binding is ready
        And jsonpath "{.spec.containerSpecs[0].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.volumeData}" on "notpodspecs/$scenario_id-npc" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"

    @spec
    Scenario: Projecting environment variables into non-podspec fields
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
            kind: NotPodSpec
            metadata:
                name: $scenario_id-npc
            spec:
                containerSpecs:
                  - id: $scenario_id-id
                    image: scratch
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                    name: $scenario_id-secret
                    kind: Secret
                    apiVersion: v1
                workload:
                    apiVersion: stable.example.com/v1
                    kind: NotPodSpec
                    name: $scenario_id-npc
                env:
                  - name: BINDING_TYPE
                    key:  type
                  - name: BINDING_USERNAME
                    key:  username
                  - name: BINDING_PASSWORD
                    key:  password
            """
        Then Service Binding is ready
        And jsonpath "{.spec.containerSpecs[0].envData}" on "notpodspecs/$scenario_id-npc" should return "[{"name":"BINDING_TYPE","valueFrom":{"secretKeyRef":{"key":"type","name":"$scenario_id-secret"}}},{"name":"BINDING_USERNAME","valueFrom":{"secretKeyRef":{"key":"username","name":"$scenario_id-secret"}}},{"name":"BINDING_PASSWORD","valueFrom":{"secretKeyRef":{"key":"password","name":"$scenario_id-secret"}}},{"name":"SERVICE_BINDING_ROOT","value":"/bindings"}]"

    @spec
    Scenario: Projecting values into a non-PodSpec resource for a specific API version
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
            kind: NotPodSpec
            metadata:
                name: $scenario_id-npc
            spec:
                containerSpecs:
                  - id: $scenario_id
                    image: scratch
                initContainerSpecs:
                  - id: $scenario_id
                    image: scratch
            """
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: notpodspecs.stable.example.com
            spec:
                versions:
                  - version: "v1"
                    volumes: .spec.volumeData
                    containers:
                      - path: .spec.containerSpecs[*]
                        name: .id
                        env: .envData
                        volumeMounts: .volumeEntries
                  - version: "*"
                    volumes: .spec.volumeData
                    containers:
                      - path: .spec.initContainerSpecs[*]
                        name: .id
                        env: .envData
                        volumeMounts: .volumeEntries
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                    name: $scenario_id-secret
                    kind: Secret
                    apiVersion: v1
                workload:
                    apiVersion: stable.example.com/v1
                    kind: NotPodSpec
                    name: $scenario_id-npc
            """
        Then Service Binding is ready
        And jsonpath "{.spec.containerSpecs[0].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.initContainerSpecs[0].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return no value
        And jsonpath "{.spec.volumeData}" on "notpodspecs/$scenario_id-npc" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"

    @spec
    Scenario: A workload resource mapping that doesn't specify defaults should use PodSpec-compatible defaults
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
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "*"
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig
            spec:
                template:
                    spec:
                        initContainers:
                          - name: foo
                        containers:
                          - name: bar
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                    name: $scenario_id-secret
                    kind: Secret
                    apiVersion: v1
                workload:
                    apiVersion: stable.example.com/v1
                    kind: AppConfig
                    name: $scenario_id-appconfig
            """
        Then Service Binding is ready
        And jsonpath "{.spec.template.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.template.spec.initContainers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.template.spec.volumes}" on "appconfigs/$scenario_id-appconfig" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"
        And The Workload Resource Mapping is deleted
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            """

    @spec
    Scenario: Binding filtering uses data from workload resource mappings
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: foo
                password: bar
                type: db
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: NotPodSpec
            metadata:
                name: $scenario_id-npc
            spec:
                initContainerSpecs:
                  - id: "foo"
                    image: init:latest
                  - id: "baz"
                    image: setup:latest
                containerSpecs:
                  - id: "baz"
                    image: some/image
                  - id: "foo"
                    image: foo:latest
                  - id: "bar"
                    image: bar:latest
                  - id: "quux"
                    image: some/image
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                    apiVersion: v1
                    kind: Secret
                    name: $scenario_id-secret
                workload:
                    name: $scenario_id-npc
                    apiVersion: stable.example.com/v1
                    kind: NotPodSpec
                    containers:
                        - foo
                        - bar
                        - bla
            """
        Then Service Binding is ready
        And jq ".status.binding.name" of Service Binding should be changed to "$scenario_id-secret"
        And jsonpath "{.spec.containerSpecs[0].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return no value
        And jsonpath "{.spec.containerSpecs[1].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.containerSpecs[2].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.containerSpecs[3].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return no value
        And jsonpath "{.spec.initContainerSpecs[0].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.initContainerSpecs[1].volumeEntries}" on "notpodspecs/$scenario_id-npc" should return no value

    @spec
    Scenario: Fall back to default volumes when not specified
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: foo
                password: bar
                type: db
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig
            spec:
                spec:
                    containers:
                        - name: bar
            """
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "*"
                    containers:
                      - path: .spec.spec.containers[*]
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                    apiVersion: v1
                    kind: Secret
                    name: $scenario_id-secret
                workload:
                    name: $scenario_id-appconfig
                    apiVersion: stable.example.com/v1
                    kind: AppConfig
            """
        Then Service Binding is ready
        And jsonpath "{.spec.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.template.spec.volumes}" on "appconfigs/$scenario_id-appconfig" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"
        And jsonpath "{.spec.template.spec.containers}" on "appconfigs/$scenario_id-appconfig" should return no value
        And jsonpath "{.spec.spec.volumes}" on "appconfigs/$scenario_id-appconfig" should return no value

    @spec
    Scenario: Fall back to default container paths when not specified
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                username: foo
                password: bar
                type: db
            """
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id-appconfig
            spec:
                template:
                    spec:
                        containers:
                            - name: bar
            """
        And The Workload Resource Mapping is present
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "*"
                    volumes: .spec.spec.volumes
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                    apiVersion: v1
                    kind: Secret
                    name: $scenario_id-secret
                workload:
                    name: $scenario_id-appconfig
                    apiVersion: stable.example.com/v1
                    kind: AppConfig
            """
        Then Service Binding is ready
        And jsonpath "{.spec.template.spec.volumes}" on "appconfigs/$scenario_id-appconfig" should return no value
        And jsonpath "{.spec.spec.containers}" on "appconfigs/$scenario_id-appconfig" should return no value
        And jsonpath "{.spec.template.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id-appconfig" should return "[{"mountPath":"/bindings/$scenario_id-binding","name":"$scenario_id-binding"}]"
        And jsonpath "{.spec.spec.volumes}" on "appconfigs/$scenario_id-appconfig" should return "[{"name":"$scenario_id-binding","secret":{"secretName":"$scenario_id-secret"}}]"

    @spec
    @negative
    Scenario: Multiple version entries with the same versions may not be specified
        When Invalid Workload Resource Mapping is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "v1"
                  - version: "v1"
            """
        Then Error message is thrown

    @spec
    @negative
    Scenario: Invalid fixed jsonpaths may not be set
        When Invalid Workload Resource Mapping is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - version: "v1"
                    annotations: .spec.template.spec.annotations[*]
            """
        Then Error message is thrown

    @spec
    @negative
    Scenario: Required fields must be set
        When Invalid Workload Resource Mapping is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ClusterWorkloadResourceMapping
            metadata:
                name: appconfigs.stable.example.com
            spec:
                versions:
                  - containers:
                      - name: .metadata.name
            """
        Then Error message is thrown
