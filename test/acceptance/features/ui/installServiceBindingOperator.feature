@ui @disabled
Feature: Install Service Binding Operator

    Background: Logged into OpenShift Console
        Given user [ADMIN_USER] is logged into OpenShift Console with [ADMIN_PASSWORD]
        * "Administrator" view is opened

    Scenario: Install Community version of Service Binding Operator via OperaHub in OpenShift DevConsole
        Given "OperatorHub" page is opened
        * "Community" checkbox for "Provider Type" is selected
        * "Service Binding Operator" card is clicked and Community operators confirmed
        * "Install" button is clicked
        * "Installation Mode" is selected to be "All namespaces on the cluster (default)"
        * "Update Channel" is selected to be "beta"
        * "Approval Strategy" is selected to be "Automatic"
        * "Install" button is clicked
        * "Installing operator" is shown on page
        When "View Operator" button is clicked
        Then Operator page for "Service Binding Operator" is opened
        And Operator status is "Succeeded"

    Scenario: Install upstream version of Service Binding Operator from OperatorHub.io
        Given Catalog source is present
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: CatalogSource
            metadata:
                name: operatorhub-io
                namespace: openshift-marketplace
            spec:
                sourceType: grpc
                image: quay.io/operatorhubio/catalog:latest
                displayName: OperatorHub.io
            """
        * "OperatorHub" page is opened
        * "OperatorHub.io" checkbox for "Provider Type" is selected
        * "Service Binding Operator" card is clicked and Community operators confirmed
        * "Install" button is clicked
        * "Installation Mode" is selected to be "All namespaces on the cluster (default)"
        * "Update Channel" is selected to be "beta"
        * "Approval Strategy" is selected to be "Automatic"
        * "Install" button is clicked
        * "Installing operator" is shown on page
        When "View Operator" button is clicked
        Then Operator page for "Service Binding Operator" is opened
        And Operator status is "Succeeded"

    Scenario: Install productized version of Service Binding Operator released by Red Hat via OperaHub in OpenShift DevConsole
        * "OperatorHub" page is opened
        * "OperatorHub.io" checkbox for "Provider Type" is selected
        * "Service Binding Operator" card is clicked and Community operators confirmed
        * "Install" button is clicked
        * "Installation Mode" is selected to be "All namespaces on the cluster (default)"
        * "Update Channel" is selected to be "beta"
        * "Approval Strategy" is selected to be "Automatic"
        * "Install" button is clicked
        * "Installing operator" is shown on page
        When "View Operator" button is clicked
        Then Operator page for "Service Binding Operator" is opened
        And Operator status is "Succeeded"