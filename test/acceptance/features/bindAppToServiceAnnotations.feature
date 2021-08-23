@annotations
Feature: Bind an application to a service using annotations

    As a user of Service Binding Operator
    I want to bind application to services that expose bindable information
    via annotations placed on service's CRD or CR.

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Provide binding info through backing service CRD annotation and ensure app env vars reflect it
        Given Generic test application "rsa-2-service" is running
        Given The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/ready: path={.status.ready}
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
                                        host:
                                            type: string
                                status:
                                    type: object
                                    properties:
                                        ready:
                                            type: boolean
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                        - bk
            """
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: backend-demo
            spec:
                host: example.common
            status:
                ready: true
            """
        * Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-backend-a
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: backend-demo
                    id: SBR
                application:
                    name: rsa-2-service
                    group: apps
                    version: v1
                    resource: deployments
            """

        Then Service Binding "binding-request-backend-a" is ready
        And The application env var "BACKEND_READY" has value "true"
        And The application env var "BACKEND_HOST" has value "example.common"

    Scenario: Each value in referred map from service resource gets injected into app as separate env variable
        Given Generic test application "rsa-2-service" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/spec: path={.spec},elementType=map
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
                                        bootstrap:
                                            type: array
                                            items:
                                                type: object
                                                properties:
                                                    type:
                                                        type: string
                                                    url:
                                                        type: string
                                                    name:
                                                        type: string
                                        data:
                                            type: object
                                            properties:
                                                dbConfiguration:
                                                    type: string
                                                dbCredentials:
                                                    type: string
                                                url:
                                                    type: string
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
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: rsa-2-service
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                bootstrap:
                  - type: plain
                    url: myhost2.example.com
                    name: hostGroup1service.binding/
                  - type: tls
                    url: myhost1.example.com:9092,myhost2.example.com:9092
                    name: hostGroup2
                data:
                    dbConfiguration: database-config     # ConfigMap
                    dbCredentials: database-cred-Secret  # Secret
                    url: db.stage.ibm.com
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: rsa-2
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: rsa-2-service
                application:
                    name: rsa-2-service
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "rsa-2" is ready
        And The application env var "BACKEND_SPEC_IMAGE" has value "docker.io/postgres"
        And The application env var "BACKEND_SPEC_IMAGENAME" has value "postgres"
        And The application env var "BACKEND_SPEC_DBNAME" has value "db-demo"

    Scenario: Each value in referred slice of strings from service resource gets injected into app as separate env variable
        Given Generic test application "slos-app" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/tags: path={.spec.tags},elementType=sliceOfStrings
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
                                        tags:
                                            type: array
                                            items:
                                                type: string
                                status:
                                    type: object
                                    properties:
                                        somestatus:
                                            type: string
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
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: slos-service
            spec:
                tags:
                  - knowledge
                  - is
                  - power
            status:
                somestatus: good
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: slos-binding
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: slos-service
                application:
                    name: slos-app
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "slos-binding" is ready
        And The application env var "BACKEND_TAGS_0" has value "knowledge"
        And The application env var "BACKEND_TAGS_1" has value "is"
        And The application env var "BACKEND_TAGS_2" has value "power"

    Scenario: Values extracted from each map by a given key in referred slice of maps from service resource gets injected into app as separate env variable
        Given Generic test application "slom-to-slos-app" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/url: path={.spec.connections},elementType=sliceOfStrings,sourceValue=url
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
                                        connections:
                                            type: array
                                            items:
                                                type: object
                                                properties:
                                                    type:
                                                        type: string
                                                    url:
                                                        type: string
                                status:
                                    type: object
                                    properties:
                                        somestatus:
                                            type: string
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
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: slom-to-slos-service
            spec:
                connections:
                  - type: primary
                    url: primary.example.com
                  - type: secondary
                    url: secondary.example.com
                  - type: '404'
                    url: black-hole.example.com
            status:
                somestatus: good
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: slom-to-slos-binding
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: slom-to-slos-service
                application:
                    name: slom-to-slos-app
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "slom-to-slos-binding" is ready
        And The application env var "BACKEND_URL_0" has value "primary.example.com"
        And The application env var "BACKEND_URL_1" has value "secondary.example.com"
        And The application env var "BACKEND_URL_2" has value "black-hole.example.com"

    Scenario: Each value in referred slice of maps from service resource gets injected into app as separate env variable
        Given Generic test application "slom-app" is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/webarrows: path={.spec.connections},elementType=sliceOfMaps,sourceKey=type,sourceValue=url
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
                                        connections:
                                            type: array
                                            items:
                                                type: object
                                                properties:
                                                    type:
                                                        type: string
                                                    url:
                                                        type: string
                                status:
                                    type: object
                                    properties:
                                        somestatus:
                                            type: string
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
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: slom-service
            spec:
                connections:
                  - type: primary
                    url: primary.example.com
                  - type: secondary
                    url: secondary.example.com
                  - type: '404'
                    url: black-hole.example.com
            status:
                somestatus: good
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: slom-binding
            spec:
                bindAsFiles: false
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: slom-service
                application:
                    name: slom-app
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "slom-binding" is ready
        And The application env var "BACKEND_WEBARROWS_PRIMARY" has value "primary.example.com"
        And The application env var "BACKEND_WEBARROWS_SECONDARY" has value "secondary.example.com"
        And The application env var "BACKEND_WEBARROWS_404" has value "black-hole.example.com"

    Scenario: Bind referring service using group version resource
        Given Generic test application "binding-service-via-gvr" is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: binding-service-via-gvr-service
                annotations:
                    service.binding/host_internal_db: path={.spec.host_internal_db}
            spec:
                host_internal_db: internal.db.stable.example.com
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-service-via-gvr
            spec:
                bindAsFiles: false
                application:
                    name: binding-service-via-gvr
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    resource: backends
                    name: binding-service-via-gvr-service
            """
        Then Service Binding "binding-service-via-gvr" is ready
        And The application env var "BACKEND_HOST_INTERNAL_DB" has value "internal.db.stable.example.com"

    Scenario: Bind referring application using group version kind
        Given Generic test application "binding-app-via-gvk" is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: binding-app-via-gvk-service
                annotations:
                    service.binding/host_internal_db: path={.spec.host_internal_db}
            spec:
                host_internal_db: internal.db.stable.example.com
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-app-via-gvk
            spec:
                bindAsFiles: false
                application:
                    name: binding-app-via-gvk
                    group: apps
                    version: v1
                    kind: Deployment
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: binding-app-via-gvk-service
            """
        Then Service Binding "binding-app-via-gvk" is ready
        And The application env var "BACKEND_HOST_INTERNAL_DB" has value "internal.db.stable.example.com"
