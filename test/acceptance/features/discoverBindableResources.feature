Feature: Discover bindable resources in a cluster

    As a user of Service Binding Operator
    I would like to discover resources that could participate in a service binding

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        And Service Binding Operator is running

    @external-feedback
    Scenario: Discover the list of bindable kinds by reading status part of cluster-scoped BindableKinds resource
        Given OLM Operator "provisioned_backend_with_annotations" is running
        And OLM Operator "backend_with_annotations" is running
        Then bindablekinds/bindable-kinds is available in the cluster
        And User acceptance-tests-dev can read resource bindablekinds/bindable-kinds
        And Kind ProvisionedBackend with apiVersion stable.example.com/v1 is listed in bindable kinds
        And Kind BindableBackend with apiVersion stable.example.com/v1 is listed in bindable kinds
