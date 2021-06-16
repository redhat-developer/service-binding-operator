Feature: Making Service Biniding Operator RBAC pluggable so that other controllers/admins can add additional rules for the service binding operator.
    Scenario: Service Binding Operator installation should contain aggregating cluster role and are bound with Namespace.
        Given Namespace [TEST_NAMESPACE] is used
        When Service Binding Operator is running
        Then cluster role "service-binding-operator-controller-role" is available in the cluster
        And operator service account is bound to "service-binding-operator-controller-role" in "service-binding-operator-controller-rolebinding"
