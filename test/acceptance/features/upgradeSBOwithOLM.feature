@upgrade-with-olm @olm
Feature: Upgrade SBO with OLM

  As the user of SBO I want to make sure SBO can be upgraded with OLM from the previous version.

  Background:
    Given Namespace [TEST_NAMESPACE] is used
    * Service Binding Operator is not running

  Scenario: Upgrade SBO with OLM from previous version
    # Install previous version
    Given [TEST_OPERATOR_CSV] CSV from [TEST_OPERATOR_PACKAGE] package is installed from [TEST_OPERATOR_CHANNEL] channel of the [TEST_OPERATOR_INDEX_IMAGE] index image
    * Operator subscription is approved
    * Service Binding Operator is running
    # Upgrade to latest version
    * Operator subscription is approved
    * Service Binding Operator is running
    * Waited for 10 seconds
    # Smoke test SBO
    * CustomResourceDefinition backends.stable.example.com is available
    * Generic test application is running
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
                group: apps
                version: v1
                kind: Deployment
                name: $scenario_id
        """
    * Service Binding is ready
    * Service Binding is applied
        """
        apiVersion: servicebinding.io/v1beta1
        kind: ServiceBinding
        metadata:
            name: $scenario_id-binding-spec
        spec:
            type: backend
            service:
                apiVersion: stable.example.com/v1
                kind: Backend
                name: $scenario_id-backend
            workload:
                apiVersion: apps/v1
                kind: Deployment
                name: $scenario_id
        """
    When Service Binding is ready
    * The env var "host" is not available to the application
    * The env var "port" is not available to the application
    * The application env var "SERVICE_BINDING_ROOT" has value "/bindings"
    * Content of file "/bindings/$scenario_id-binding-spec/host" in application pod is
        """
        example.common
        """
    * Content of file "/bindings/$scenario_id-binding-spec/port" in application pod is
        """
        8080
        """
    * Content of file "/bindings/$scenario_id-binding/host" in application pod is
        """
        example.common
        """
    * Content of file "/bindings/$scenario_id-binding/port" in application pod is
        """
        8080
        """
