Feature: Bind values from a config map referred in backing service resource

    As a user I would like to inject into my app as env variables
    values persisted in a secret referred within service resource

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Inject into app a key from a secret referred within service resource
        Binding definition is declared on service CRD.
        Given OLM Operator "backend" is running
        And Generic test application "ssa-1" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/username: path={.status.data.dbCredentials},objectType=Secret,valueKey=username
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                      - bk
            """
        And The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: ssa-1-secret
            stringData:
                username: AzureDiamond
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: ssa-1-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: ssa-1-secret
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: ssa-1
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: ssa-1-service
                application:
                    name: ssa-1
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "ssa-1" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"

    Scenario: Inject into app all keys from a secret referred within service resource
        Binding definition is declared on service CRD.

        Given OLM Operator "backend" is running
        And Generic test application "ssa-2" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                      - bk
            """
        And The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: ssa-2-secret
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: ssa-2-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: ssa-2-secret
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: ssa-2
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: ssa-2-service
                application:
                    name: ssa-2
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "ssa-2" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    Scenario: Inject into app a key from a secret referred within service resource
        Binding definition is declared via OLM descriptor.

        Given Generic test application "ssd-1" is running
        And The Custom Resource Definition is present
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ClusterServiceVersion
            metadata:
              annotations:
                capabilities: Basic Install
              name: backend-operator.v0.1.0
            spec:
              customresourcedefinitions:
                owned:
                - description: Backend is the Schema for the backend API
                  kind: Backend
                  name: backends.stable.example.com
                  version: v1
                  specDescriptors:
                    - description: Host address
                      displayName: Host address
                      path: host
                      x-descriptors:
                        - service.binding:host
                  statusDescriptors:
                      - description: db credentials
                        displayName: db credentials
                        path: data.dbCredentials
                        x-descriptors:
                            - urn:alm:descriptor:io.kubernetes:Secret
                            - service.binding:username:sourceValue=username
              displayName: Backend Operator
              install:
                spec:
                  deployments:
                  - name: backend-operator
                    spec:
                      replicas: 1
                      selector:
                        matchLabels:
                          name: backend-operator
                      strategy: {}
                      template:
                        metadata:
                          labels:
                            name: backend-operator
                        spec:
                          containers:
                          - command:
                            - backend-operator
                            env:
                            - name: WATCH_NAMESPACE
                              valueFrom:
                                fieldRef:
                                  fieldPath: metadata.annotations['olm.targetNamespaces']
                            - name: POD_NAME
                              valueFrom:
                                fieldRef:
                                  fieldPath: metadata.name
                            - name: OPERATOR_NAME
                              value: backend-operator
                            image: REPLACE_IMAGE
                            imagePullPolicy: Always
                            name: backend-operator
                            resources: {}
                strategy: deployment
            """
        And The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: ssd-1-secret
            stringData:
                username: AzureDiamond
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: ssd-1-service
            spec:
                host: example.com
            status:
                data:
                    dbCredentials: ssd-1-secret
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: ssd-1
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: ssd-1-service
                application:
                    name: ssd-1
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "ssd-1" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_HOST" has value "example.com"

    Scenario: Inject into app all keys from a secret referred within service resource
        Binding definition is declared via OLM descriptor.

        Given Generic test application "ssd-2" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                      - bk
            """
        And The Custom Resource is present
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ClusterServiceVersion
            metadata:
              annotations:
                capabilities: Basic Install
              name: backend-operator.v0.1.0
            spec:
              customresourcedefinitions:
                owned:
                - description: Backend is the Schema for the backend API
                  kind: Backend
                  name: backends.stable.example.com
                  version: v1
                  specDescriptors:
                    - description: Host address
                      displayName: Host address
                      path: host
                      x-descriptors:
                        - service.binding:host
                  statusDescriptors:
                      - description: db credentials
                        displayName: db credentials
                        path: data.dbCredentials
                        x-descriptors:
                            - urn:alm:descriptor:io.kubernetes:Secret
                            - service.binding:elementType=map
              displayName: Backend Operator
              install:
                spec:
                  deployments:
                  - name: backend-operator
                    spec:
                      replicas: 1
                      selector:
                        matchLabels:
                          name: backend-operator
                      strategy: {}
                      template:
                        metadata:
                          labels:
                            name: backend-operator
                        spec:
                          containers:
                          - command:
                            - backend-operator
                            env:
                            - name: WATCH_NAMESPACE
                              valueFrom:
                                fieldRef:
                                  fieldPath: metadata.annotations['olm.targetNamespaces']
                            - name: POD_NAME
                              valueFrom:
                                fieldRef:
                                  fieldPath: metadata.name
                            - name: OPERATOR_NAME
                              value: backend-operator
                            image: REPLACE_IMAGE
                            imagePullPolicy: Always
                            name: backend-operator
                            resources: {}
                strategy: deployment
            """
        And The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: ssd-2-secret
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: ssd-2-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: ssd-2-secret
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: ssd-2
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: ssd-2-service
                application:
                    name: ssd-2
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "ssd-2" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"
