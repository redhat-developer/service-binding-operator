@spec
Feature: Support spec v1beta1 resources

    As a user, I would like to use v1beta1 resources as defined by the binding specification.

    Background:
        Given OLM Operator "provisioned_backend" is running
        And OLM Operator "not_a_pod" is running
        And Namespace [TEST_NAMESPACE] is used

    Scenario: Bind application to provisioned service using v1beta1 ServiceBindings
        Given Generic test application is running
        And The Secret is present
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
            apiVersion: stable.example.com/v1
            kind: ProvisionedBackend
            metadata:
                name: $scenario_id-backend
            spec:
                foo: bar
            status:
                binding:
                    name: $scenario_id-secret
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1beta1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                service:
                  apiVersion: stable.example.com/v1
                  kind: ProvisionedBackend
                  name: $scenario_id-backend
                workload:
                  name: $scenario_id
                  apiVersion: apps/v1
                  kind: Deployment
            """
        Then Service Binding is ready
        And jq ".status.binding.name" of Service Binding "$scenario_id-binding" should be changed to "$scenario_id-secret"
        And Content of file "/bindings/$scenario_id-binding/username" in application pod is
            """
            foo
            """
        And Content of file "/bindings/$scenario_id-binding/password" in application pod is
            """
            bar
            """
        And Content of file "/bindings/$scenario_id-binding/type" in application pod is
            """
            db
            """

    Scenario: Project bindings to custom resource using v1beta1 resources
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
            apiVersion: servicebinding.io/v1beta1
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
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1beta1
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

    Scenario: Project a v1beta1 service binding using a v1beta1 workload mapping
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
            apiVersion: servicebinding.io/v1beta1
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
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1beta1
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

    Scenario: Project a v1beta1 service binding using a v1beta1 workload mapping
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
            apiVersion: servicebinding.io/v1beta1
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
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1beta1
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

    @negative
    Scenario: Reject invalid v1alpha3 workload resource mapping
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

    @negative
    Scenario: Reject invalid v1alpha3 service binding
        Given CustomResourceDefinition backends.stable.example.com is available
        And The Custom Resource is present
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
        When Invalid Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
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
                  name: $scenario_id-backend
            """
        Then Error message "name and selector MUST NOT be defined in the application reference" is thrown
