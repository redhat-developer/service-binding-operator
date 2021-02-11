Feature: Reconcile when BackingService CR got deleted and recreated

    As a user of Service Binding Operator
    I want the SBR to be reconciled when there backend service was deleted and created again

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
    Scenario: Reconcile when BackingService CR got deleted and recreated
        Given OLM Operator "backend" is running
        And Generic test application "ssa-3" is running
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
                name: ssa-3-secret
            stringData:
                username: AzureDiamond
                password: hunter
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: ssa-3-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: ssa-3-secret
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: ssa-3
            spec:
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: ssa-3-service
                application:
                    name: ssa-3
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "ssa-3" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "ssa-3" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "ssa-3" should be changed to "True"
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter"
	    When BackingService is deleted
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: ssa-3-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: ssa-3-secret
            """
        And The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: ssa-3-secret
            stringData:
                username: AzureDiamond2
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: ssa-3-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: ssa-3-secret
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "ssa-3" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "ssa-3" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "ssa-3" should be changed to "True"
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond2"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

