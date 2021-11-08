Feature: Unbind an application from a service

    As a user of Service Binding Operator
    I want to unbind an application from a service

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * CustomResourceDefinition backends.stable.example.com is available

    Scenario: Unbind a generic test application from the backing service
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.com
                username: foo
            """
        * Generic test application is running
        * Service Binding is applied
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
                    id: backend
            """
        * Service Binding "$scenario_id-binding" is ready
        * The application env var "BACKEND_HOST" has value "example.com"
        * The application env var "BACKEND_USERNAME" has value "foo"
        When Service binding "$scenario_id-binding" is deleted
        Then The env var "BACKEND_HOST" is not available to the application
        * The env var "BACKEND_USERNAME" is not available to the application
        * Service Binding secret is not present

    Scenario: Unbind a generic test application from the backing service when the backing service has been deleted
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.com
                username: foo
            """
        * Generic test application is running
        * Service Binding is applied
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
                    name: example-backend
                    id: backend
            """
        * Service Binding "$scenario_id-binding" is ready
        * The application env var "BACKEND_HOST" has value "example.com"
        * The application env var "BACKEND_USERNAME" has value "foo"
        * BackingService is deleted
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.com
                username: foo
            """
        When Service binding "$scenario_id-binding" is deleted
        Then The env var "BACKEND_HOST" is not available to the application
        * The env var "BACKEND_USERNAME" is not available to the application
        * Service Binding secret is not present

    Scenario: Remove bindings projected as files from generic test application
        Given Generic test application is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: "8080"
            """
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend

                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        * Service Binding "$scenario_id-binding" is ready
        * Content of file "/bindings/$scenario_id-binding/host" in application pod is
            """
            example.common
            """
        * Content of file "/bindings/$scenario_id-binding/port" in application pod is
            """
            8080
            """
        When Service Binding "$scenario_id-binding" is deleted
        Then The application got redeployed 2 times so far
        * File "/bindings/$scenario_id-binding/host" is unavailable in application pod
        * File "/bindings/$scenario_id-binding/port" is unavailable in application pod
        * Service Binding secret is not present

    Scenario: Remove not ready binding
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.com
                username: foo
            """
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                application:
                    name: not-found-app
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: example-backend
            """
        * jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "$scenario_id-binding" should be changed to "False"
        When Service binding "$scenario_id-binding" is deleted
        Then Service Binding "$scenario_id-binding" is not persistent in the cluster

    @smoke
    @spec
    Scenario: SPEC Remove bindings from test application
        Given Generic test application is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: "8080"
            """
        * Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
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
        * Service Binding "scenario_id-binding" is ready
        * Content of file "/bindings/$scenario_id-binding/host" in application pod is
            """
            example.common
            """
        * Content of file "/bindings/$scenario_id-binding/port" in application pod is
            """
            8080
            """
        * Content of file "/bindings/$scenario_id-binding/type" in application pod is
            """
            mysql
            """
        When Service Binding "$scenario_id-binding" is deleted
        Then The application got redeployed 2 times so far
        * File "/bindings/$scenario_id-binding/host" is unavailable in application pod
        * File "/bindings/$scenario_id-binding/port" is unavailable in application pod
        * File "/bindings/$scenario_id-binding/type" is unavailable in application pod
        * Service Binding secret is not present
