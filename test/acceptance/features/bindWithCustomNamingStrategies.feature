Feature: Bind an application to a service using custom naming strategies

    As a user of Service Binding Operator
    I want to bind applications to services it depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * CustomResourceDefinition backends.stable.example.com is available

    Scenario: Bind an application to a service with no naming strategy specified
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
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
                name: $scenario_id-binding
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
                    name: $scenario_id-backend
            """
        Then Service Binding is ready
        And The application env var "BACKEND_HOST_CROSS_NS_SERVICE" has value "cross.ns.service.stable.example.com"

    Scenario: Bind an application to a service with naming strategy none
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
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
                name: binding-request-naming-none
            spec:
                bindAsFiles: false
                namingStrategy: none
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend
            """
        Then Service Binding is ready
        And The application env var "host_cross_ns_service" has value "cross.ns.service.stable.example.com"

    Scenario: Bind an application to a service with custom naming strategy
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
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
                name: $scenario_id-backend
            spec:
                bindAsFiles: false
                namingStrategy: "PREFIX_{{ .service.kind | upper }}_{{ .name | upper }}_SUFFIX"
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend
            """
        Then Service Binding is ready
        And The application env var "PREFIX_BACKEND_HOST_CROSS_NS_SERVICE_SUFFIX" has value "cross.ns.service.stable.example.com"

    Scenario: Bind an application to a service with bind as file and no naming strategy
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
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
                name: $scenario_id-binding
            spec:
                bindAsFiles: true
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend
            """
        Then Service Binding is ready
        And The env var "host_cross_ns_service" is not available to the application
        And Content of file "/bindings/$scenario_id-binding/host_cross_ns_service" in application pod is
            """
            cross.ns.service.stable.example.com
            """

    Scenario: Bind an application to a service with bind files and custom naming strategy
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
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
                name: $scenario_id-binding
            spec:
                namingStrategy: "PREFIX_{{ .service.kind | upper }}_{{ .name | upper }}_SUFFIX"
                bindAsFiles: true
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend
            """
        Then Service Binding is ready
        And Content of file "/bindings/$scenario_id-binding/PREFIX_BACKEND_HOST_CROSS_NS_SERVICE_SUFFIX" in application pod is
            """
            cross.ns.service.stable.example.com
            """

    @error-naming
    Scenario: Bind an application to a service with naming strategy error
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
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
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                namingStrategy: "{{ .service.test.name | lower }}_incorrect"
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "$scenario_id-binding" should be changed to "False"
        Then jq ".status.conditions[] | select(.type=="CollectionReady").reason" of Service Binding "$scenario_id-binding" should be changed to "NamingStrategyError"
