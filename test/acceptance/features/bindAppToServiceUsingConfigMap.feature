Feature: Bind values from a config map referred in backing service resource

    As a user I would like to inject into my app as env variables
    values persisted in a config map referred within service resource

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Inject into app a key from a config map referred within service resource
        Binding definition is declared on service CRD.

        Given OLM Operator "backend" is running
        And Generic test application "cmsa-1" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
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
                name: cmsa-1-configmap
            data:
                certificate: "certificate value"
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: cmsa-1-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbConfiguration: cmsa-1-configmap
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: cmsa-1
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: cmsa-1-service
                application:
                    name: cmsa-1
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "cmsa-1" is ready
        And The application env var "BACKEND_CERTIFICATE" has value "certificate value"

    Scenario: Inject into app all keys from a config map referred within service resource
        Binding definition is declared on service CRD.

        Given OLM Operator "backend" is running
        And Generic test application "cmsa-2" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
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
                name: cmsa-2-configmap
            data:
                timeout: "30"
                certificate: certificate value
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: cmsa-2-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbConfiguration: cmsa-2-configmap    # ConfigMap
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: cmsa-2
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: cmsa-2-service
                application:
                    name: cmsa-2
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "cmsa-2" is ready
        And The application env var "BACKEND_CERTIFICATE" has value "certificate value"
