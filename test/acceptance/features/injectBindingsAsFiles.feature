@olm
Feature: Bindings get injected as files in application

    As a user of Service Binding Operator
    I want to make service binding data accessible to application
    through files available in application pods

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * OLM Operator "backend" is running

    Scenario: Binding is injected as files at the location of SERVICE_BINDING_ROOT env var
        Given Generic test application "generic-app-a-d-u-1" is running with binding root as "/var/data"
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo-01
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-backend-vm-01
            spec:
                bindAsFiles: true
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-01
                    id: bk

                mappings:
                  - name: MYHOST
                    value: '{{ .bk.spec.host }}'

                application:
                    name: generic-app-a-d-u-1
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-backend-vm-01" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-backend-vm-01" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The env var "MYHOST" is not available to the application
        And The application env var "SERVICE_BINDING_ROOT" has value "/var/data"
        And Content of file "/var/data/binding-backend-vm-01/host" in application pod is
            """
            example.common
            """
        And Content of file "/var/data/binding-backend-vm-01/port" in application pod is
            """
            8080
            """
        And Content of file "/var/data/binding-backend-vm-01/MYHOST" in application pod is
            """
            example.common
            """

    Scenario: Binding is injected as file into application at default location
        Given Generic test application "generic-app-a-d-u-2" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo-02
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-backend-vm-02
            spec:
                bindAsFiles: true
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-02

                application:
                    name: generic-app-a-d-u-2
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-backend-vm-02" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-backend-vm-02" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The application env var "SERVICE_BINDING_ROOT" has value "/bindings"
        And Content of file "/bindings/binding-backend-vm-02/host" in application pod is
            """
            example.common
            """
        And Content of file "/bindings/binding-backend-vm-02/port" in application pod is
            """
            8080
            """

    Scenario: Binding is injected as file into application at the location specified through mountPath
        Given Generic test application "generic-app-a-d-u-3" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo-03
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-backend-vm-03
            spec:
                mountPath: "/foo/bar"
                bindAsFiles: true
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-03

                application:
                    name: generic-app-a-d-u-3
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-backend-vm-03" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-backend-vm-03" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The env var "SERVICE_BINDING_ROOT" is not available to the application
        And Content of file "/foo/bar/host" in application pod is
            """
            example.common
            """
        And Content of file "/foo/bar/port" in application pod is
            """
            8080
            """

    Scenario: Binding is injected as files at the location of SERVICE_BINDING_ROOT env var even with mountPath
        Given Generic test application "generic-app-a-d-u-4" is running with binding root as "/var/data"
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo-04
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-backend-vm-04
            spec:
                mountPath: "/foo/bar"
                bindAsFiles: true
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-04
                    id: bk

                mappings:
                  - name: MYHOST
                    value: '{{ .bk.spec.host }}'

                application:
                    name: generic-app-a-d-u-4
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-backend-vm-04" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-backend-vm-04" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The application env var "SERVICE_BINDING_ROOT" has value "/var/data"
        And Content of file "/var/data/binding-backend-vm-04/host" in application pod is
            """
            example.common
            """
        And Content of file "/var/data/binding-backend-vm-04/port" in application pod is
            """
            8080
            """
        And Content of file "/var/data/binding-backend-vm-04/MYHOST" in application pod is
            """
            example.common
            """

    Scenario: Binding is injected as file into application at the location specified through mountPath with empty prefix
        Given Generic test application "generic-app-a-d-u-5" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo-05
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-backend-vm-05
            spec:
                mountPath: "/foo/bar"
                bindAsFiles: true
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-05

                application:
                    name: generic-app-a-d-u-5
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-backend-vm-05" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-backend-vm-05" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The env var "SERVICE_BINDING_ROOT" is not available to the application
        And Content of file "/foo/bar/host" in application pod is
            """
            example.common
            """
        And Content of file "/foo/bar/port" in application pod is
            """
            8080
            """

    Scenario: After updating mountpath bindings are reinjected to the new location
        Given Generic test application "generic-app-a-d-u-6" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo-06
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-backend-vm-06
            spec:
                mountPath: "/foo/bar"
                bindAsFiles: true
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-06

                application:
                    name: generic-app-a-d-u-6
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-backend-vm-06" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-backend-vm-06" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The env var "SERVICE_BINDING_ROOT" is not available to the application
        And Content of file "/foo/bar/host" in application pod is
            """
            example.common
            """
        And Content of file "/foo/bar/port" in application pod is
            """
            8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-backend-vm-06
            spec:
                mountPath: "/foo/newbar"
                bindAsFiles: true
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo-06

                application:
                    name: generic-app-a-d-u-6
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-backend-vm-06" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-backend-vm-06" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The env var "SERVICE_BINDING_ROOT" is not available to the application
        And Content of file "/foo/newbar/host" in application pod is
            """
            example.common
            """
        And Content of file "/foo/newbar/port" in application pod is
            """
            8080
            """
