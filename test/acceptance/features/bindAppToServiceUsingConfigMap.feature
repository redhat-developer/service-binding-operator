Feature: Bind values from a config map referred in backing service resource

    As a user I would like to inject into my app as env variables
    values persisted in a config map referred within service resource

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Inject into app a key from a config map referred within service resource
        Given Generic test application is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/certificate: path={.status.data.dbConfiguration},objectType=ConfigMap,sourceKey=certificate
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                    schema:
                        openAPIV3Schema:
                            type: object
                            properties:
                                apiVersion:
                                    type: string
                                kind:
                                    type: string
                                metadata:
                                    type: object
                                spec:
                                    type: object
                                    properties:
                                        image:
                                            type: string
                                        imageName:
                                            type: string
                                        dbName:
                                            type: string
                                status:
                                    type: object
                                    properties:
                                        data:
                                            type: object
                                            properties:
                                                dbConfiguration:
                                                    type: string
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                      - bk
            """
        And The ConfigMap is present
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: $scenario_id-configmap
            data:
                certificate: "certificate value"
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbConfiguration: $scenario_id-configmap
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-service
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id-binding" is ready
        And The application env var "BACKEND_CERTIFICATE" has value "certificate value"

    Scenario: Inject into app all keys from a config map referred within service resource
        Given Generic test application is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding: path={.status.data.dbConfiguration},objectType=ConfigMap,elementType=map
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                    schema:
                        openAPIV3Schema:
                            type: object
                            properties:
                                apiVersion:
                                    type: string
                                kind:
                                    type: string
                                metadata:
                                    type: object
                                spec:
                                    type: object
                                    properties:
                                        image:
                                            type: string
                                        imageName:
                                            type: string
                                        dbName:
                                            type: string
                                status:
                                    type: object
                                    properties:
                                        data:
                                            type: object
                                            properties:
                                                dbConfiguration:
                                                    type: string
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                      - bk
            """
        And The ConfigMap is present
            """
            apiVersion: v1
            kind: ConfigMap
            metadata:
                name: $scenario_id-configmap
            data:
                timeout: "30"
                certificate: certificate value
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbConfiguration: $scenario_id-configmap    # ConfigMap
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id-binding" is ready
        And The application env var "BACKEND_CERTIFICATE" has value "certificate value"
