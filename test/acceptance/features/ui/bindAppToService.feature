@ui @disabled
Feature: Bind an application to a service

    As a user of Service Binding Operator
    I want to bind applications to services it depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed

    Scenario: Bind Node.js application to DB in Topology view of OpenShift Console
        Given user [ADMIN_USER] is logged into OpenShift Console with [ADMIN_PASSWORD]
        * "Developer" view is opened
        * "Topology" page is opened
        * DB "db-demo-ui" is running
        * Imported Nodejs application "nodejs-app-ui" is not running
        * DB "db-demo-ui" is shown
        * Application "nodejs-app-ui" is shown
        * Arrow is dragged and dropped from application icon to DB icon to "Create a binding connector"
        Then Arrow is rendered from application icon to DB icon to indicate binding
        And application should be re-deployed
        And application should be connected to the DB "db-demo-ui"
