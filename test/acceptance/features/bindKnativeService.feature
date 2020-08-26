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
    When Service Binding Request is applied to connect the database and the application
      """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
              name: binding-request-knative
            spec:
              applicationSelector:
                group: serving.knative.dev
                version: v1beta1
                resource: services
                resourceRef: knative-app
              backingServiceSelector:
                group: postgresql.baiju.dev
                version: v1alpha1
                kind: Database
                resourceRef: db-demo-knative
                id: knav
              customEnvVar:
                - name: JDBC_URL
                  value: jdbc:postgresql://{{ .knav.status.dbConnectionIP }}:{{ .knav.status.dbConnectionPort }}/{{ .knav.status.dbName }}
                - name: DB_USER
                  value: "{{ .knav.status.dbCredentials.user }}"
                - name: DB_PASSWORD
                  value: "{{ .knav.status.dbCredentials.password }}"
      """
    Then deployment must contain intermediate secret "binding-request-knative"
    Then application should be connected to the DB "db-demo-knative"
    And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-knative" should be changed to "True"
    And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-knative" should be changed to "True"