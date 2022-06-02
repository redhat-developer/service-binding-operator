@optional-annotations
@external-feedback
Feature: Bind an application to a service using optional annotations

    As a user of Service Binding Operator
    I want to bind application to services that expose optional bindable information
    via annotations placed on service's CRD or CR.

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * OLM Operator "custom_app" is running
        * CustomResourceDefinition backends.stable.example.com is available

    Scenario: Bind to the application with annotations and with optional source field present
        Given Generic test application is running
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/uri: "path={.spec.uri},optional=true"
                    service.binding/image: "path={.spec.image}"
            spec:
                uri: "youknow.where"
                image: "busybox"
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: true
                services:
                  - group: stable.example.com
                    version: v1
                    kind: AppConfig
                    name: $scenario_id
                application:
                    group: apps
                    version: v1
                    resource: deployments
                    name: $scenario_id
            """
        Then Service Binding is ready
        * Content of file "/bindings/$scenario_id/uri" in application pod is
            """
            youknow.where
            """
        * Content of file "/bindings/$scenario_id/image" in application pod is
            """
            busybox
            """

    Scenario: Bind to the application with annotations and with missing optional source field to be excluded
        Given Generic test application is running
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/uri: "path={.spec.uri},optional=true"
                    service.binding/image: "path={.spec.image}"
            spec:
                image: "busybox"
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: true
                services:
                  - group: stable.example.com
                    version: v1
                    kind: AppConfig
                    name: $scenario_id
                application:
                    group: apps
                    version: v1
                    resource: deployments
                    name: $scenario_id
            """
        Then Service Binding is ready
        * File "/bindings/$scenario_id/uri" is unavailable in application pod
        * Content of file "/bindings/$scenario_id/image" in application pod is
            """
            busybox
            """

    @negative
    Scenario: Bind to the application with annotations and with missing non-optional source field to cause binding failure
        Given Generic test application is running
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/uri: "path={.spec.uri}"
                    service.binding/image: "path={.spec.image}"
            spec:
                image: "busybox"
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: true
                services:
                  - group: stable.example.com
                    version: v1
                    kind: AppConfig
                    name: $scenario_id
                application:
                    group: apps
                    version: v1
                    resource: deployments
                    name: $scenario_id
            """
        Then Service Binding CollectionReady.status is "False"
        And Service Binding CollectionReady.reason is "ErrorReadingBinding"

    @negative
    Scenario: Bind to the application with annotations and with missing non-optional source field to cause binding failure with explicit false
        Given Generic test application is running
        And The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id
                annotations:
                    service.binding/uri: "path={.spec.uri}"
                    service.binding/image: "path={.spec.image}"
            spec:
                image: "busybox"
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: true
                services:
                  - group: stable.example.com
                    version: v1
                    kind: AppConfig
                    name: $scenario_id
                application:
                    group: apps
                    version: v1
                    resource: deployments
                    name: $scenario_id
            """
        Then Service Binding CollectionReady.status is "False"
        And Service Binding CollectionReady.reason is "ErrorReadingBinding"
