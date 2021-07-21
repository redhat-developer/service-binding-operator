Feature: Successful Service Binding are Immutable

    As a user of Service Binding operator
    I should not be able to modify ready Service Binding

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
        """
        apiVersion: stable.example.com/v1
        kind: Backend
        metadata:
            name: service-immutable
            annotations:
                service.binding/host: path={.spec.host}
        spec:
            host: foo
        """

    Scenario: Cannot update a ready Service Binding
        Given Generic test application "app-immutable" is running
        And Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-immutable
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-immutable
                application:
                    name: app-immutable
                    group: apps
                    version: v1
                    resource: deployments
            """
        When Service Binding "binding-immutable" is ready
        Then Service Binding is unable to be applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-immutable
            spec:
                application:
                    name: app-immutable-2
                    group: apps
                    version: v1
                    resource: deployments
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-immutable
            """

    Scenario: Allow modifying a not-ready Service Binding
        Given Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-immutable-2
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-immutable
                application:
                    name: app1
                    group: apps
                    version: v1
                    resource: deployments
            """
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-immutable-2" should be changed to "False"
        When Generic test application "app2" is running
        And Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-immutable-2
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-immutable
                application:
                    name: app2
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "binding-immutable-2" is ready


    @spec
    Scenario: SPEC Cannot update a ready Service Binding
        Given Generic test application "spec-app-immutable" is running
        And Service Binding is applied
            """
            apiVersion: service.binding/v1alpha2
            kind: ServiceBinding
            metadata:
                name: spec-binding-immutable
            spec:
                type: foo
                service:
                  apiVersion: stable.example.com/v1
                  kind: Backend
                  name: service-immutable
                application:
                    name: spec-app-immutable
                    apiVersion: apps/v1
                    kind: Deployment
            """
        When Service Binding "spec-binding-immutable" is ready
        Then Service Binding is unable to be applied
            """
            apiVersion: service.binding/v1alpha2
            kind: ServiceBinding
            metadata:
                name: spec-binding-immutable
            spec:
                service:
                  apiVersion: stable.example.com/v1
                  kind: Backend
                  name: service-immutable
                application:
                    name: spec-app-immutable2
                    apiVersion: apps/v1
                    kind: Deployment
            """
