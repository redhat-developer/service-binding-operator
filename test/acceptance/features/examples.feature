@examples
Feature: Verify examples provided in Service Binding Operator github repository

    As a user of Service Binding Operator
    I want to verify if all the examples given are vouching for the mentioned features

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running

    # https://github.com/redhat-developer/service-binding-operator/blob/master/examples/route_k8s_resource/README.md
    Scenario: Bind an application to an openshift route
        Given The openshift route is present
            """
            kind: Route
            apiVersion: route.openshift.io/v1
            metadata:
                name: my-route
                annotations:
                    openshift.io/host.generated: 'true'
                    service.binding/host: path={.spec.host} #annotate here.
            spec:
                host: example-sbo.apps.ci-ln-smyggvb-d5d6b.origin-ci-int-aws.dev.rhcloud.com
                path: /
                to:
                    kind: Service
                    name: my-service
                    weight: 100
                port:
                    targetPort: 80
            wildcardPolicy: None
            """
        * "hello-app" is deployed from image "openshift/hello-openshift"
        When Service Binding is applied
            """
            apiVersion: operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: sbr-to-bind-hello-app-to-route
            spec:
                application:
                    group: apps
                    resource: deployments
                    name: hello-app
                    version: v1
                services:
                  - group: route.openshift.io
                    version: v1
                    kind: Route
                    name: my-route
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "sbr-to-bind-hello-app-to-route" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "sbr-to-bind-hello-app-to-route" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="Ready").status" of Service Binding "sbr-to-bind-hello-app-to-route" should be changed to "True"
        And Secret "sbr-to-bind-hello-app-to-route" contains "ROUTE_HOST" key with value "example-sbo.apps.ci-ln-smyggvb-d5d6b.origin-ci-int-aws.dev.rhcloud.com"
