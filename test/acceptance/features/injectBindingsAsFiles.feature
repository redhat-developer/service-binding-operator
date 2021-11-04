Feature: Bindings get injected as files in application

    As a user of Service Binding Operator
    I want to make service binding data accessible to application
    through files available in application pods

    Background:
        Given Namespace [TEST_NAMESPACE] is used
        * Service Binding Operator is running
        * CustomResourceDefinition backends.stable.example.com is available

    Scenario: Binding is injected as files at the location of SERVICE_BINDING_ROOT env var
        Given Generic test application is running with binding root as "/var/data"
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: "8080"
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                bindAsFiles: true
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend
                    id: bk

                mappings:
                  - name: MYHOST
                    value: '{{ .bk.spec.host }}'

                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "$scenario_id-binding" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "$scenario_id-binding" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The env var "MYHOST" is not available to the application
        And The application env var "SERVICE_BINDING_ROOT" has value "/var/data"
        And Content of file "/var/data/$scenario_id-binding/host" in application pod is
            """
            example.common
            """
        And Content of file "/var/data/$scenario_id-binding/port" in application pod is
            """
            8080
            """
        And Content of file "/var/data/$scenario_id-binding/MYHOST" in application pod is
            """
            example.common
            """

    Scenario: Binding is injected as file into application at default location
        Given Generic test application is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: "8080"
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                services:
                -   group: stable.example.com
                    version: v1
                    kind: Backend
                    name: $scenario_id-backend

                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then jq ".status.conditions[] | select(.type=="CollectionReady").status" of Service Binding "$scenario_id-binding" should be changed to "True"
        And jq ".status.conditions[] | select(.type=="InjectionReady").status" of Service Binding "$scenario_id-binding" should be changed to "True"
        And The env var "host" is not available to the application
        And The env var "port" is not available to the application
        And The application env var "SERVICE_BINDING_ROOT" has value "/bindings"
        And Content of file "/bindings/$scenario_id-binding/host" in application pod is
            """
            example.common
            """
        And Content of file "/bindings/$scenario_id-binding/port" in application pod is
            """
            8080
            """
        And The container declared in application resource contains env "SERVICE_BINDING_ROOT" set only once

    @negative
    Scenario: Do not bind as files if there is no binding data is collected from the service
        Given Generic test application is running
        * The Service is present
            """
            apiVersion: v1
            kind: Service
            metadata:
                name: $scenario_id-svc
            spec:
                selector:
                    name: $scenario_id-svc
                ports:
                  - port: 8080
                    targetPort: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                bindAsFiles: true
                detectBindingResources: true
                services:
                  - group: ""
                    version: v1
                    kind: Service
                    name: $scenario_id-svc
                application:
                    name: $scenario_id
                    group: apps
                    version: v1
                    resource: deployments
            """
        Then Service Binding CollectionReady.status is "True"
        * Service Binding InjectionReady.status is "False"
        * Service Binding InjectionReady.reason is "NoBindingData"

    @spec
    Scenario: SPEC Inject bindings gathered through annotations into application at default location
        Given Generic test application is running
        * The Custom Resource is present
            """
            apiVersion: "stable.example.com/v1"
            kind: Backend
            metadata:
                name: $scenario_id-backend
                annotations:
                    "service.binding/host": "path={.spec.host}"
                    "service.binding/port": "path={.spec.port}"
            spec:
                host: example.common
                port: "8080"
            """
        When Service Binding is applied
            """
            apiVersion: servicebinding.io/v1alpha3
            kind: ServiceBinding
            metadata:
                name: $scenario_id-binding
            spec:
                type: mysql
                service:
                  apiVersion: stable.example.com/v1
                  kind: Backend
                  name: $scenario_id-backend

                workload:
                    name: $scenario_id
                    apiVersion: apps/v1
                    kind: Deployment
            """
        Then Service Binding "$scenario_id-binding" is ready
        And The application env var "SERVICE_BINDING_ROOT" has value "/bindings"
        And Content of file "/bindings/$scenario_id-binding/host" in application pod is
            """
            example.common
            """
        And Content of file "/bindings/$scenario_id-binding/port" in application pod is
            """
            8080
            """
        And Content of file "/bindings/$scenario_id-binding/type" in application pod is
            """
            mysql
            """

    Scenario: SERVICE_BINDING_ROOT is not defined twice in the deployment after binding two services
        Given Generic test application is running
        * The env var "SERVICE_BINDING_ROOT" is not available to the application
        * The Service is present
            """
            apiVersion: v1
            kind: Service
            metadata:
                name: $scenario_id-2
                annotations:
                    service.binding/service2: path={.metadata.name}
            spec:
                selector:
                    name: $scenario_id
                ports:
                  - port: 80
                    targetPort: 8080
            """
        * The Service is present
            """
            apiVersion: v1
            kind: Service
            metadata:
                name: $scenario_id-3
                annotations:
                    service.binding/service3: path={.metadata.name}
            spec:
                selector:
                    name: $scenario_id
                ports:
                  - port: 80
                    targetPort: 8080
            """
        When Service Binding is applied
            """
            apiVersion: binding.operators.coreos.com/v1alpha1
            kind: ServiceBinding
            metadata:
                name: $scenario_id
            spec:
                application:
                    group: apps
                    version: v1
                    resource: deployments
                    name: $scenario_id
                bindAsFiles: true
                detectBindingResources: true
                services:
                  - group: ""
                    version: v1
                    kind: Service
                    name: $scenario_id-2
                  - group: ""
                    version: v1
                    kind: Service
                    name: $scenario_id-3
            """
        Then Service Binding is ready
        * Service Binding has the binding secret name set in the status
        * The container declared in application resource contains env "SERVICE_BINDING_ROOT" set only once
