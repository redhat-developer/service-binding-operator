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
                name: backend-no-naming
                annotations:
                    service.binding/host_cross_ns_service: path={.spec.host_cross_ns_service}
            spec:
                host_cross_ns_service: cross.ns.service.stable.example.com
            """
        * Generic test application "myapp-no-naming" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-no-naming
            spec:
                application:
                    name: myapp-no-naming
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-no-naming
            """
        Then Service Binding "binding-request-no-naming" is ready
        And The application env var "BACKEND_HOST_CROSS_NS_SERVICE" has value "cross.ns.service.stable.example.com"

    Scenario: Bind an application to a service with naming strategy none
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: backend-naming-none
                annotations:
                    service.binding/host_cross_ns_service: path={.spec.host_cross_ns_service}
            spec:
                host_cross_ns_service: cross.ns.service.stable.example.com
            """
        * Generic test application "myapp-naming-none" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-naming-none
            spec:
                namingStrategy: none
                application:
                    name: myapp-naming-none
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-naming-none
            """
        Then Service Binding "binding-request-naming-none" is ready
        And The application env var "host_cross_ns_service" has value "cross.ns.service.stable.example.com"

    Scenario: Bind an application to a service with custom naming strategy
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: backend-custom-naming
                annotations:
                    service.binding/host_cross_ns_service: path={.spec.host_cross_ns_service}
            spec:
                host_cross_ns_service: cross.ns.service.stable.example.com
            """
        * Generic test application "myapp-custom-naming" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-custom-naming
            spec:
                namingStrategy: "PREFIX_{{ .service.kind | upper }}_{{ .name | upper }}_SUFFIX"
                application:
                    name: myapp-custom-naming
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-custom-naming
            """
        Then Service Binding "binding-request-custom-naming" is ready
        And The application env var "PREFIX_BACKEND_HOST_CROSS_NS_SERVICE_SUFFIX" has value "cross.ns.service.stable.example.com"

    Scenario: Bind an application to a service with bind as file and no naming strategy
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: backend-bind-files
                annotations:
                    service.binding/host_cross_ns_service: path={.spec.host_cross_ns_service}
            spec:
                host_cross_ns_service: cross.ns.service.stable.example.com
            """
        * Generic test application "myapp-bind-files" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-bind-files
            spec:
                bindAsFiles: true
                application:
                    name: myapp-bind-files
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-bind-files
            """
        Then Service Binding "binding-request-bind-files" is ready
        And The env var "host_cross_ns_service" is not available to the application
        And Content of file "/bindings/binding-request-bind-files/host_cross_ns_service" in application pod is
            """
            cross.ns.service.stable.example.com
            """

    Scenario: Bind an application to a service with bind files and custom naming strategy
        Given The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: backend-custom-file-naming
                annotations:
                    service.binding/host_cross_ns_service: path={.spec.host_cross_ns_service}
            spec:
                host_cross_ns_service: cross.ns.service.stable.example.com
            """
        * Generic test application "myapp-custom-file-naming" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-custom-file-naming
            spec:
                namingStrategy: "PREFIX_{{ .service.kind | upper }}_{{ .name | upper }}_SUFFIX"
                mountPath: "/foo/bar"
                bindAsFiles: true
                application:
                    name: myapp-custom-file-naming
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-custom-file-naming
            """
        Then Service Binding "binding-request-custom-file-naming" is ready
        And Content of file "/foo/bar/PREFIX_BACKEND_HOST_CROSS_NS_SERVICE_SUFFIX" in application pod is
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
                name: backend-naming-error
                annotations:
                    service.binding/host_cross_ns_service: path={.spec.host_cross_ns_service}
            spec:
                host_cross_ns_service: cross.ns.service.stable.example.com
            """
        * Generic test application "myapp-naming-error" is running
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-naming-error
            spec:
                namingStrategy: "{{ .service.test.name | lower }}_incorrect"
                application:
                    name: myapp-naming-error
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-naming-error
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-naming-error" should be changed to "False"
        Then jq ".status.conditions[] | select(.type=="CollectionReady").reason" of Service Binding "binding-request-naming-error" should be changed to "NamingStrategyError"
