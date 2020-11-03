Feature: Insert service binding to a custom location in application resource

    As a user of Service Binding Operator
    I want to insert service binding to custom location in application resource.
    The type of such location needs to be specified (corev1.Containers, corev1.Volumes, secretRef)

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * PostgreSQL DB operator is installed
        * The Custom Resource Definition is present
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

    Scenario: Specify container's path in Service Binding
        Given DB "db-demo-csp" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: demo-appconfig-csp
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
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-csp
            spec:
                envVarPrefix: qiye111
                application:
                    name: demo-appconfig-csp
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                    bindingPath:
                        containersPath: spec.spec.containers
                services:
                  - group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-csp
                    id: zzz
                    envVarPrefix: qiye
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-csp" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-csp" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-request-csp" should be changed to "True"
        And Secret "binding-request-csp" has been injected in to CR "demo-appconfig-csp" of kind "AppConfig" at path "{.spec.spec.containers[0].envFrom[0].secretRef.name}"

    Scenario: Specify secret's path in Service Binding
        Given DB "db-demo-ssp" is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: AppConfig
            metadata:
                name: demo-appconfig-ssp
            spec:
                spec:
                    secret: some-value
            """
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-ssp
            spec:
                envVarPrefix: qiye111
                application:
                    name: demo-appconfig-ssp
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                    bindingPath:
                        secretPath: spec.spec.secret
                services:
                  - group: postgresql.baiju.dev
                    version: v1alpha1
                    kind: Database
                    name: db-demo-ssp
                    id: zzz
                    envVarPrefix: qiye
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "binding-request-ssp" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "binding-request-ssp" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "binding-request-ssp" should be changed to "True"
        And Secret "binding-request-ssp" has been injected in to CR "demo-appconfig-ssp" of kind "AppConfig" at path "{.spec.spec.secret}"
