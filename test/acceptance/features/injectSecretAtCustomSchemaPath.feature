Feature: Insert service binding to a custom location in application resource

    As a user of Service Binding Operator
    I want to insert service binding to custom location in application resource.
    The type of such location needs to be specified (corev1.Containers, corev1.Volumes, secretRef)

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * CustomResourceDefinition backends.stable.example.com is available
        * The Secret is present
            """
            apiVersion: v1
            kind: Secret
            metadata:
                name: backend-secret
            stringData:
                username: AzureDiamond
                password: hunter2
            """
        * The Custom Resource is present
            """
            apiVersion: stable.example.com/v1
            kind: Backend
            metadata:
                name: service-csp
                annotations:
                    service.binding: path={.status.data.dbCredentials},objectType=Secret,elementType=map
            status:
                data:
                    dbCredentials: backend-secret
            """
        * OLM Operator "custom_app" is running

    # https://github.com/redhat-developer/service-binding-operator/tree/master/examples/pod_spec_path
    Scenario: Specify container's path in Service Binding
        Given The Custom Resource is present
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
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-csp
            spec:
                bindAsFiles: false
                application:
                    name: demo-appconfig-csp
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                    bindingPath:
                        containersPath: spec.spec.containers
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-csp
            """
        Then Service Binding "binding-request-csp" is ready
        And Secret has been injected in to CR "demo-appconfig-csp" of kind "AppConfig" at path "{.spec.spec.containers[0].envFrom[0].secretRef.name}"

    Scenario: Specify secret's path in Service Binding
        Given The Custom Resource is present
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
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: binding-request-ssp
            spec:
                bindAsFiles: false
                application:
                    name: demo-appconfig-ssp
                    group: stable.example.com
                    version: v1
                    resource: appconfigs
                    bindingPath:
                        secretPath: spec.spec.secret
                services:
                  - group: stable.example.com
                    version: v1
                    kind: Backend
                    name: service-csp
            """
        Then Service Binding "binding-request-ssp" is ready
        And Secret has been injected in to CR "demo-appconfig-ssp" of kind "AppConfig" at path "{.spec.spec.secret}"
