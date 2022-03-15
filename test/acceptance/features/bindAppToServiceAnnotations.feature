@annotations
Feature: Bind an application to a service using annotations

    As a user of Service Binding Operator
    I want to bind application to services that expose bindable information
    via annotations placed on service's CRD or CR.

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Provide binding info through backing service CRD annotation and ensure app env vars reflect it
        Given Generic test application is running
        Given The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/host: path={.spec.host}
                    service.binding/ready: path={.status.ready}
                    service.binding/environment: path={.spec.userLabels.environment}
                    service.binding/dataType: path={.status.data.type}
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
                                        userLabels:
                                            type: object
                                            properties:
                                                environment:
                                                    type: string
                                status:
                                    type: object
                                    properties:
                                        ready:
                                            type: boolean
                                        data:
                                            type: object
                                            properties:
                                                type:
                                                    type: string
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
                name: $scenario_id-backend
            spec:
                host: example.common
                userLabels:
                    environment: staging
            status:
                ready: true
                data:
                    type: base64
            """
        * Service Binding is applied
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
                    id: SBR
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """

        Then Service Binding is ready
        And The application env var "BACKEND_READY" has value "true"
        And The application env var "BACKEND_HOST" has value "example.common"
        And The application env var "BACKEND_ENVIRONMENT" has value "staging"
        And The application env var "BACKEND_DATATYPE" has value "base64"

    Scenario: Each value in referred map from service resource gets injected into app as separate env variable
        Given Generic test application is running
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
                name: $scenario_id-backend
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
        Then Service Binding is ready
        And The application env var "BACKEND_SPEC_IMAGE" has value "docker.io/postgres"
        And The application env var "BACKEND_SPEC_IMAGENAME" has value "postgres"
        And The application env var "BACKEND_SPEC_DBNAME" has value "db-demo"

    Scenario: Each value in referred slice of strings from service resource gets injected into app as separate env variable
        Given Generic test application is running
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
                name: $scenario_id-backend
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
                name: $scenario_id
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
        Then Service Binding is ready
        And The application env var "BACKEND_TAGS_0" has value "knowledge"
        And The application env var "BACKEND_TAGS_1" has value "is"
        And The application env var "BACKEND_TAGS_2" has value "power"

    Scenario: Values extracted from each map by a given key in referred slice of maps from service resource gets injected into app as separate env variable
        Given Generic test application is running
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
                name: $scenario_id-backend
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
        Then Service Binding is ready
        And The application env var "BACKEND_URL_0" has value "primary.example.com"
        And The application env var "BACKEND_URL_1" has value "secondary.example.com"
        And The application env var "BACKEND_URL_2" has value "black-hole.example.com"

    Scenario: Each value in referred slice of maps from service resource gets injected into app as separate env variable
        Given Generic test application is running
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
                name: $scenario_id-backend
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
        Then Service Binding is ready
        And The application env var "BACKEND_WEBARROWS_PRIMARY" has value "primary.example.com"
        And The application env var "BACKEND_WEBARROWS_SECONDARY" has value "secondary.example.com"
        And The application env var "BACKEND_WEBARROWS_404" has value "black-hole.example.com"

    Scenario: Bind referring service using group version resource
        Given Generic test application is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
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
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
                services:
                -   group: stable.example.com
                    version: v1
                    resource: backends
                    name: $scenario_id-backend
            """
        Then Service Binding is ready
        And The application env var "BACKEND_HOST_INTERNAL_DB" has value "internal.db.stable.example.com"

    Scenario: Bind referring application using group version kind
        Given Generic test application is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
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
                name: $scenario_id-binding
            spec:
                bindAsFiles: false
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    kind: Deployment
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend
            """
        Then Service Binding is ready
        And The application env var "BACKEND_HOST_INTERNAL_DB" has value "internal.db.stable.example.com"

    @external-feedback
    Scenario: Application cannot be bound to service containing annotation with an invalid sourceValue value
        Given Generic test application is running
        And CustomResourceDefinition backends.stable.example.com is available
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/webarrows: path={.spec.connections},elementType=sliceOfMaps,sourceKey=type,sourceValue=asdf
            spec:
                connections:
                  - type: primary
                    url: primary.example.com
            status:
                somestatus: good
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
        Then Service Binding CollectionReady.status is "False"
        And Service Binding CollectionReady.reason is "ValueNotFound"
        And Service Binding CollectionReady.message is "Value for key webarrows_primary not found"

    Scenario: Application cannot be bound to service containing invalid elementType annotation
        Given Generic test application is running
        And CustomResourceDefinition backends.stable.example.com is available
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/credentials: path={.spec.connections.dbCredentials},elementType=asdf
            spec:
                connections:
                  - type: primary
                    url: primary.example.com
            status:
                somestatus: good
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
        Then Service Binding CollectionReady.status is "False"
        And Service Binding CollectionReady.reason is "InvalidAnnotation"
        And Service Binding Ready.message is "Annotation service.binding/credentials: path={.spec.connections.dbCredentials},elementType=asdf not implemented!"

    Scenario: Application cannot be bound to service containing invalid objectType annotation
        Given Generic test application is running
        And CustomResourceDefinition backends.stable.example.com is available
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/credentials: path={.spec.connections.dbCredentials},objectType=asdf,sourceKey=username
            spec:
                connections:
                  - type: primary
                    url: primary.example.com
            status:
                somestatus: good
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
        Then Service Binding CollectionReady.status is "False"
        And Service Binding CollectionReady.reason is "InvalidAnnotation"
        And Service Binding Ready.message is "Annotation service.binding/credentials: path={.spec.connections.dbCredentials},objectType=asdf,sourceKey=username not implemented!"

    Scenario: Application cannot be bound to service containing invalid path annotation
        Given Generic test application is running
        And CustomResourceDefinition backends.stable.example.com is available
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    service.binding/credentials: path=asdf
            spec:
                connections:
                  - type: primary
                    url: primary.example.com
            status:
                somestatus: good
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
        Then Service Binding CollectionReady.status is "False"
        And Service Binding CollectionReady.reason is "InvalidAnnotation"
        And Service Binding CollectionReady.message is "Failed to create binding definition from "service.binding/credentials: path=asdf": could not create binding model for annotation key service.binding/credentials and value path=asdf: path has invalid syntax: "asdf""
        And Service Binding Ready.message is "could not create binding model for annotation key service.binding/credentials and value path=asdf: path has invalid syntax: "asdf""
