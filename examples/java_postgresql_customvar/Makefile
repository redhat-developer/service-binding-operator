.DEFAULT_GOAL := help

include ../commons.mk

## -- Application Developer targets --

.PHONY: set-labels-on-java-app
## Set the binding labels for the  app
set-labels-on-java-app:
	${Q}oc project service-binding-demo
	${Q}oc patch dc java-app -p '{"metadata": {"labels": {"connects-to": "postgres", "environment": "demo"}}}'

.PHONY: create-service-binding-request
## Create the Service Binding Request
create-service-binding-request:
	${Q}oc apply -f service-binding-request.yaml
