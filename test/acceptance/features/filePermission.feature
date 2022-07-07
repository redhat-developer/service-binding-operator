@file-permissions
Feature: File permissions of the bound files

    As a user I want ensure the files has the least perssmions

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Bound files has the least permissions
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-secret
            stringData:
                testfile: baz
            """
        * Generic test application is running
        When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id-binding
          spec:
              bindAsFiles: true
              services:
              - group: ""
                version: v1
                kind: Secret
                name: $scenario_id-secret
                id: sec
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
              mappings:
                - name: username_with_password
                  value: '{{ .username }}:{{ .password }}'
          """
        Then Service Binding is ready
        And File "testfile" has the least permissions
