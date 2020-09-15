.DEFAULT_GOAL := help

include ../commons.mk

## -- Application Developer targets --

.PHONY: create-service-binding
## Create the Service Binding
create-service-binding:
	${Q}oc apply -f service-binding.yaml
