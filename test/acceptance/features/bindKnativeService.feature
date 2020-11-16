Feature: Bind knative service to a service

    As a user of Service Binding Operator
    I want to bind knative service to services it depends on

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed

    Scenario: Bind an imported quarkus app which is deployed as knative service to PostgreSQL database
        Given Openshift Serverless Operator is running
        * Knative serving is running
        * DB "db-demo-knative" is running
        * Quarkus application "knative-app" is imported as Knative service
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
              name: binding-request-knative
            spec:
              application:
                group: serving.knative.dev
                version: v1beta1
                resource: services
                name: knative-app
              services:
              - group: postgresql.baiju.dev
                version: v1alpha1
                kind: Database
                name: db-demo-knative
                id: knav
              customEnvVar:
                - name: JDBC_URL
                  value: jdbc:postgresql://{{ .knav.status.dbConnectionIP }}:{{ .knav.status.dbConnectionPort }}/{{ .knav.status.dbName }}
                - name: DB_USER
                  value: "{{ .knav.status.dbCredentials.user }}"
                - name: DB_PASSWORD
                  value: "{{ .knav.status.dbCredentials.password }}"
            """
        Then Service Binding "binding-request-knative" is ready
        And deployment must contain intermediate secret "binding-request-knative"
        And application should be connected to the DB "db-demo-knative"
