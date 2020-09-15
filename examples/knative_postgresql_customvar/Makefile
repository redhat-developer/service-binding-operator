include ../commons.mk

.DEFAULT_GOAL := help

## -- Cluster Admin Targets --

.PHONY: install-all-operators
## Install all the operators
install-all-operators: install-backing-db-operator install-service-binding-operator install-serverless-operator install-service-mesh-operator

.PHONY: uninstall-all-operators
## Uninstall all the operators
uninstall-all-operators: uninstall-backing-db-operator uninstall-service-binding-operator uninstall-serverless-operator uninstall-service-mesh-operator

.PHONY: install-all
## Install all the operators, Knative Serving and the Quarkus native s2i builder image
install-all: install-all-operators install-knative-serving install-quarkus-native-s2i-builder

## -- Application Developer targets --

.PHONY: set-labels-on-knative-app
## Set binding labels on the knative application
set-labels-on-knative-app:
	${Q}oc label services.serving.knative.dev knative-app connects-to=postgres environment=demo --overwrite

.PHONY: create-service-binding
## Create the Service Binding
create-service-binding:
	${Q}oc apply -f service-binding.yaml

.PHONY: install-serverless-ui
## Create the knative namespace, install knative serving
install-serverless-ui:
	${Q}oc apply -f serving.yaml

