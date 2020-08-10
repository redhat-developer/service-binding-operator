Feature: Associate binding secret to a custom schema path

    As a user of Service Binding Operator
    I want to bind secret in an arbitrary location.

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed

    Scenario: Bind secret to a custom schema path
        Given DB "db-demo-csp" is running
        * The CRD "appconfigs.stable.example.com" is present
            """
            apiVersion: apiextensions.k8s.io/v1beta1
            kind: CustomResourceDefinition
            metadata:
                name: appconfigs.stable.example.com
            spec:
                group: stable.example.com
                versions:
                  - name: v1
                    served: true
                    storage: true
                scope: Namespaced
                names:
                    plural: appconfigs
                    singular: appconfig
                    kind: AppConfig
                    shortNames:
                    - ac
            """
        * The application CR "demo-appconfig" is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: demo-appconfig
            spec:
                uri: "some uri"
                Command: "some command"
                image: my-image
                spec:
                    containers:
                    - name: hello-world
                    # Image from dockerhub, This is the import path for the Go binary to build and run.
                    image: yusufkaratoprak/kubernetes-gosample:latest
                    ports:
                    - containerPort: 8090
            """
        When Service Binding Request is applied to connect the database and the application
            """
            apiVersion: apps.openshift.io/v1alpha1
            kind: ServiceBindingRequest
            metadata:
                name: binding-request-csp
            spec:
                envVarPrefix: qiye111
                applicationSelector:
                    resourceRef: demo-appconfig
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                    bindingPath:
                        podSpecPath:
                            containers: spec.spec.containers
                            volumes: spec.spec.volumes
                backingServiceSelectors:
                  - group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    resourceRef: db-demo-csp
                    id: zzz
                    envVarPrefix: qiye
            """
        Then Secret "binding-request-csp" has been injected in to CR "demo-appconfig" of kind "AppConfig" at path "{.spec.spec.containers[0].envFrom[0].secretRef.name}"
        And jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding Request "binding-request-csp" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding Request "binding-request-csp" should be changed to "True"
