Feature: Bind values from a secret referred in backing service resource

    As a user I would like to inject into my app as env variables
    values persisted in a secret referred within service resource

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    Scenario: Inject into app a key from a secret referred within service resource
        Binding definition is declared on service CRD.
        Given Generic test application is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding/username: path={.status.data.dbCredentials},objectType=Secret,sourceKey=username
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
                                                dbCredentials:
                                                    type: string
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
                name: $scenario_id
            stringData:
                username: AzureDiamond
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: $scenario_id
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
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"

    Scenario: Inject into app all keys from a secret referred within service resource
        Given Generic test application is running
        And The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
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
                                                dbCredentials:
                                                    type: string
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
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: $scenario_id
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
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    @olm
    Scenario: Inject into app a key from a secret referred within service resource Binding definition is declared via OLM descriptor.

        Given Generic test application is running
        * OLM Operator "backends_foo" is running
        * The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.foo.example.com
            spec:
                group: foo.example.com
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
                                        data:
                                            type: object
                                            properties:
                                                dbCredentials:
                                                    type: string
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                    - bs
            """
        And The Custom Resource Definition is present
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ClusterServiceVersion
            metadata:
              annotations:
                capabilities: Basic Install
              name: backend-operator-foo.v0.1.0
            spec:
              customresourcedefinitions:
                owned:
                - description: Backend is the Schema for the backend API
                  kind: Backend
                  name: backends.foo.example.com
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
                name: $scenario_id
            stringData:
                username: AzureDiamond
            """
        And The Custom Resource is present
            """
            apiVersion: foo.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
            spec:
                host: example.com
            status:
                data:
                    dbCredentials: $scenario_id
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
                  - group: foo.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_HOST" has value "example.com"

    @olm
    Scenario: Inject into app all keys from a secret referred within service resource Binding definition is declared via OLM descriptor.

        Given Generic test application is running
        * OLM Operator "backends_bar" is running
        * The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.bar.example.com
            spec:
                group: bar.example.com
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
                                                dbCredentials:
                                                    type: string
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                    - bs
            """
        And The Custom Resource is present
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ClusterServiceVersion
            metadata:
              annotations:
                capabilities: Basic Install
              name: backend-operator-bar.v0.1.0
            spec:
              customresourcedefinitions:
                owned:
                - description: Backend is the Schema for the backend API
                  kind: Backend
                  name: backends.bar.example.com
                  version: v1
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
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        And The Custom Resource is present
            """
            apiVersion: bar.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
            spec:
                image: docker.io/postgres
                imageName: postgres
                dbName: db-demo
            status:
                data:
                    dbCredentials: $scenario_id
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
                  - group: bar.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    Scenario: Inject into app all keys from a secret existing in a same namespace with service and different from the service binding
        Given The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding: path={.status.credentials},objectType=Secret
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
                                status:
                                    type: object
                                    properties:
                                        credentials:
                                            type: string
                scope: Namespaced
                names:
                    plural: backends
                    singular: backend
                    kind: Backend
                    shortNames:
                      - bk
            """
        * Namespace is present
            """
            apiVersion: v1
            kind: Namespace
            metadata:
                name: $scenario_id
            """
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
                namespace: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
                namespace: $scenario_id
            status:
                credentials: $scenario_id
            """
        * Generic test application is running
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
                    name: $scenario_id
                    namespace: $scenario_id
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"

    Scenario: Inject data from secret referred in field belonging to list
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        * The Custom Resource Definition is present
            """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
                name: backends.stable.example.com
                annotations:
                    service.binding: path={.spec.containers[0].envFrom[0].secretRef.name},objectType=Secret
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
                                        containers:
                                            type: array
                                            items:
                                                type: object
                                                properties:
                                                    envFrom:
                                                        type: array
                                                        items:
                                                            type: object
                                                            properties:
                                                                secretRef:
                                                                    type: object
                                                                    properties:
                                                                        name:
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
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: $scenario_id
            spec:
                containers:
                - envFrom:
                  - secretRef:
                        name: $scenario_id
            """
        * Generic test application is running
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
                name: $scenario_id
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
        Then Service Binding "$scenario_id" is ready
        And The application env var "BACKEND_USERNAME" has value "AzureDiamond"
        And The application env var "BACKEND_PASSWORD" has value "hunter2"


    Scenario: Inject binding to an application from a Secret resource referred as service
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: foo
                password: bar
            """
        * Generic test application is running
        When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: ""
                version: v1
                kind: Secret
                name: $scenario_id
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
        Then Service Binding "$scenario_id" is ready
        And jq ".status.secret" of Service Binding "$scenario_id" should be changed to "$scenario_id"
        And Content of file "/bindings/$scenario_id/username" in application pod is
            """
            foo
            """
        And Content of file "/bindings/$scenario_id/password" in application pod is
            """
            bar
            """

    Scenario: Inject binding to an application from a Secret resource referred as service with mappings
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: foo
                password: bar
            """
        * Generic test application is running
        When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              bindAsFiles: true
              services:
              - group: ""
                version: v1
                kind: Secret
                name: $scenario_id
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
        Then Service Binding "$scenario_id" is ready
        And Service Binding "$scenario_id" has the binding secret name set in the status
        And Content of file "/bindings/$scenario_id/username" in application pod is
            """
            foo
            """
        And Content of file "/bindings/$scenario_id/password" in application pod is
            """
            bar
            """
        And Content of file "/bindings/$scenario_id/username_with_password" in application pod is
            """
            foo:bar
            """

    Scenario: Inject binding to an application from a Secret resource created later referred as service 
        Given Generic test application is running
        When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: ""
                version: v1
                kind: Secret
                name: $scenario_id
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
        * jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "$scenario_id" should be changed to "False"
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: foo
                password: bar
            """

        Then Service Binding "$scenario_id" is ready
        And Content of file "/bindings/$scenario_id/username" in application pod is
            """
            foo
            """
        And Content of file "/bindings/$scenario_id/password" in application pod is
            """
            bar
            """

    Scenario: Inject binding to an application from two Secret resources referred as services
        Given Generic test application is running
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-1
            stringData:
                username: foo
                password: bar
            """
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id-2
            stringData:
                username2: foo2
                password2: bar2
            """
        When Service Binding is applied
          """
          apiVersion: binding.operators.coreos.com/v1alpha1
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              services:
              - group: ""
                version: v1
                kind: Secret
                name: $scenario_id-1
              - group: ""
                version: v1
                kind: Secret
                name: $scenario_id-2
              application:
                name: $scenario_id
                group: apps
                version: v1
                resource: deployments
          """
        Then Service Binding "$scenario_id" is ready
        And Content of file "/bindings/$scenario_id/username" in application pod is
            """
            foo
            """
        And Content of file "/bindings/$scenario_id/password" in application pod is
            """
            bar
            """
        And Content of file "/bindings/$scenario_id/username2" in application pod is
            """
            foo2
            """
        And Content of file "/bindings/$scenario_id/password2" in application pod is
            """
            bar2

            """
    @spec
    Scenario: SPEC Inject binding to an application from a Secret resource referred as service
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: foo
                password: bar
                type: db
            """
        * Generic test application is running
        When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              service:
                apiVersion: v1
                kind: Secret
                name: $scenario_id
              workload:
                name: $scenario_id
                apiVersion: apps/v1
                kind: Deployment
          """
        Then Service Binding "$scenario_id" is ready
        And jq ".status.binding.name" of Service Binding "$scenario_id" should be changed to "$scenario_id"
        And Content of file "/bindings/$scenario_id/username" in application pod is
            """
            foo
            """
        And Content of file "/bindings/$scenario_id/password" in application pod is
            """
            bar
            """
        And Content of file "/bindings/$scenario_id/type" in application pod is
            """
            db
            """

    @spec
    @negative
    Scenario: SPEC Fail to bind if type binding item is not provided
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: foo
                password: bar
            """
        * Generic test application is running
        When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              service:
                apiVersion: v1
                kind: Secret
                name: $scenario_id
              workload:
                name: $scenario_id
                apiVersion: apps/v1
                kind: Deployment
          """
        Then jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "$scenario_id" should be changed to "False"
        And jq ".status.conditions[] | select(.type=="InjectionReady").reason" of Service Binding "$scenario_id" should be changed to "RequiredBindingNotFound"

    @spec
    Scenario: SPEC Inject binding to only specified containers inside application pod
        Given The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: $scenario_id
            stringData:
                username: foo
                password: bar
                type: db
            """
        * OLM Operator "custom_app" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: $scenario_id
            spec:
                template:
                    spec:
                        initContainers:
                            - image: init:latest
                              name: "foo"
                            - image: setup:latest
                        containers:
                            - image: some/image
                            - name: "foo"
                              image: foo:latest
                            - name: "bar"
                              image: bar:latest
                            - image: some/image
            """

        When Service Binding is applied
          """
          apiVersion: servicebinding.io/v1alpha3
          kind: ServiceBinding
          metadata:
              name: $scenario_id
          spec:
              service:
                apiVersion: v1
                kind: Secret
                name: $scenario_id
              workload:
                name: $scenario_id
                apiVersion: stable.example.com/v1
                kind: AppConfig
                containers:
                    - foo
                    - bar
                    - bla
          """
        Then Service Binding "$scenario_id" is ready
        * jq ".status.binding.name" of Service Binding "$scenario_id" should be changed to "$scenario_id"
        * jsonpath "{.spec.template.spec.containers[0].volumeMounts}" on "appconfigs/$scenario_id" should return no value
        * jsonpath "{.spec.template.spec.containers[1].volumeMounts}" on "appconfigs/$scenario_id" should return "[{"mountPath":"/bindings/$scenario_id","name":"$scenario_id"}]"
        * jsonpath "{.spec.template.spec.containers[2].volumeMounts}" on "appconfigs/$scenario_id" should return "[{"mountPath":"/bindings/$scenario_id","name":"$scenario_id"}]"
        * jsonpath "{.spec.template.spec.containers[3].volumeMounts}" on "appconfigs/$scenario_id" should return no value
        * jsonpath "{.spec.template.spec.initContainers[0].volumeMounts}" on "appconfigs/$scenario_id" should return "[{"mountPath":"/bindings/$scenario_id","name":"$scenario_id"}]"
