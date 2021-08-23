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
                name: example-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.com
                username: foo
            """
        * Generic test application "generic-app-a-d-u" is running
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-a-d-u
            spec:
                bindAsFiles: false
                application:
                    name: generic-app-a-d-u
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
        * Service Binding "binding-request-a-d-u" is ready
        * The application env var "BACKEND_HOST" has value "example.com"
        * The application env var "BACKEND_USERNAME" has value "foo"
        When Service binding "binding-request-a-d-u" is deleted
        Then The env var "BACKEND_HOST" is not available to the application
        * The env var "BACKEND_USERNAME" is not available to the application
        * Service Binding secret is not present

    Scenario: Unbind a generic test application from the backing service when the backing service has been deleted
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: example-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.com
                username: foo
            """
        * Generic test application "generic-app-a-d-u" is running
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-a-d-u
            spec:
                bindAsFiles: false
                application:
                    name: generic-app-a-d-u
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
        * Service Binding "binding-request-a-d-u" is ready
        * The application env var "BACKEND_HOST" has value "example.com"
        * The application env var "BACKEND_USERNAME" has value "foo"
        * BackingService is deleted
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: example-backend
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/username: path={.spec.username}
            spec:
                host: example.com
                username: foo
            """
        When Service binding "binding-request-a-d-u" is deleted
        Then The env var "BACKEND_HOST" is not available to the application
        * The env var "BACKEND_USERNAME" is not available to the application
        * Service Binding secret is not present

    Scenario: Remove bindings projected as files from generic test application
        Given Generic test application "remove-bindings-as-files-app" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: remove-bindings-as-files-app-backend
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
                name: remove-bindings-as-files-app-sb
            spec:
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: remove-bindings-as-files-app-backend

                application:
                    name: remove-bindings-as-files-app
                    group: apps
                    version: v1
                    resource: deployments
            """
        * Service Binding "remove-bindings-as-files-app-sb" is ready
        * Content of file "/bindings/remove-bindings-as-files-app-sb/host" in application pod is
            """
            example.common
            """
        * Content of file "/bindings/remove-bindings-as-files-app-sb/port" in application pod is
            """
            8080
            """
        When Service Binding "remove-bindings-as-files-app-sb" is deleted
        Then The application got redeployed 2 times so far
        * File "/bindings/remove-bindings-as-files-app-sb/host" is unavailable in application pod
        * File "/bindings/remove-bindings-as-files-app-sb/port" is unavailable in application pod
        * Service Binding secret is not present

    Scenario: Remove not ready binding
        Given The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: example-backend
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
                name: unready-binding
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
        * jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "unready-binding" should be changed to "False"
        When Service binding "unready-binding" is deleted
        Then Service Binding "unready-binding" is not persistent in the cluster

    @smoke
    Scenario: SPEC Remove bindings from test application
        Given Generic test application "spec-remove-bindings-as-files-app" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: remove-bindings-as-files-app-backend
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: "8080"
            """
        * Service Binding is applied
            """
            apiVersion: service.binding/v1alpha2
            kind: ServiceBinding
            metadata:
                name: spec-remove-bindings-as-files-app-sb
            spec:
                type: mysql
                service:
                  apiVersion: stable.example.com/v1
                  kind: Backend
                  name: remove-bindings-as-files-app-backend

                application:
                    name: spec-remove-bindings-as-files-app
                    apiVersion: apps/v1
                    kind: Deployment
            """
        * Service Binding "spec-remove-bindings-as-files-app-sb" is ready
        * Content of file "/bindings/spec-remove-bindings-as-files-app-sb/host" in application pod is
            """
            example.common
            """
        * Content of file "/bindings/spec-remove-bindings-as-files-app-sb/port" in application pod is
            """
            8080
            """
        * Content of file "/bindings/spec-remove-bindings-as-files-app-sb/type" in application pod is
            """
            mysql
            """
        When Service Binding "spec-remove-bindings-as-files-app-sb" is deleted
        Then The application got redeployed 2 times so far
        * File "/bindings/spec-remove-bindings-as-files-app-sb/host" is unavailable in application pod
        * File "/bindings/spec-remove-bindings-as-files-app-sb/port" is unavailable in application pod
        * File "/bindings/spec-remove-bindings-as-files-app-sb/type" is unavailable in application pod
        * Service Binding secret is not present
