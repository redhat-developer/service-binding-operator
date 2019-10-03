# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /bin/bash

#-----------------------------------------------------------------------------
# VERBOSE target
#-----------------------------------------------------------------------------

# When you run make VERBOSE=1 (the default), executed commands will be printed
# before executed. If you run make VERBOSE=2 verbose flags are turned on and
# quiet flags are turned off for various commands. Use V_FLAG in places where
# you can toggle on/off verbosity using -v. Use Q_FLAG in places where you can
# toggle on/off quiet mode using -q. Use S_FLAG where you want to toggle on/off
# silence mode using -s...
VERBOSE ?= 1
Q = @
Q_FLAG = -q
QUIET_FLAG = --quiet
V_FLAG =
VERBOSE_FLAG =
S_FLAG = -s
X_FLAG =
ifeq ($(VERBOSE),1)
	Q =
endif
ifeq ($(VERBOSE),2)
	Q =
	Q_FLAG =
	QUIET_FLAG =
	S_FLAG =
	V_FLAG = -v
	VERBOSE_FLAG = --verbose
	X_FLAG = -x
endif

EC=$(SHELL) -c '. ../../hack/examples-commons.sh && $$1' EC
ifneq (,$(findstring n,$(MAKEFLAGS)))
EC=: EC
endif
export HACK_YAMLS=../../hack/yamls

.DEFAULT_GOAL := help

## -- Utility targets --

## Print help message for all Makefile targets
## Run `make` or `make help` to see the help
.PHONY: help
help: ## Credit: https://gist.github.com/prwhite/8168133#gistcomment-2749866

	@printf "Usage:\n  make <target>";

	@awk '{ \
			if ($$0 ~ /^.PHONY: [a-zA-Z\-\_0-9]+$$/) { \
				helpCommand = substr($$0, index($$0, ":") + 2); \
				if (helpMessage) { \
					printf "\033[36m%-20s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^[a-zA-Z\-\_0-9.]+:/) { \
				helpCommand = substr($$0, 0, index($$0, ":")); \
				if (helpMessage) { \
					printf "\033[36m%-20s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^##/) { \
				if (helpMessage) { \
					helpMessage = helpMessage"\n                     "substr($$0, 3); \
				} else { \
					helpMessage = substr($$0, 3); \
				} \
			} else { \
				if (helpMessage) { \
					print "\n                     "helpMessage"\n" \
				} \
				helpMessage = ""; \
			} \
		}' \
		$(MAKEFILE_LIST)

## -- Cluster Admin targets --

.PHONY: install-service-binding-operator
## Install the Service Binding Operator
install-service-binding-operator:
	${Q}${EC} install_service_binding_operator_source
	${Q}${EC} install_service_binding_operator_subscription

.PHONY: uninstall-service-binding-operator
## Uninstall the Service Binding Operator
uninstall-service-binding-operator:
	${Q}-${EC} uninstall_service_binding_operator_subscription
	${Q}-${EC} uninstall_service_binding_operator_source

.PHONY: install-backing-db-operator
## Install the Backing Service DB Operator
install-backing-db-operator:
	${Q}${EC} install_postgresql_operator_source
	${Q}${EC} install_postgresql_operator_subscription

.PHONY: uninstall-backing-db-operator
## Uninstall the Backing Service DB Operator
uninstall-backing-db-operator:
	${Q}-${EC} uninstall_postgresql_operator_subscription
	${Q}-${EC} uninstall_postgresql_operator_source

.PHONY: install-serverless-operator
## Install the Serverless Operator
install-serverless-operator:
	${Q}${EC} install_serverless_operator_subscription

.PHONY: uninstall-serverless-operator
## Uninstall the Serverless Operator
uninstall-serverless-operator:
	${Q}-${EC} uninstall_serverless_operator_subscription

.PHONY: install-service-mesh-operator
## Install the Service Mesh Operator
install-service-mesh-operator:
	${Q}${EC} install_service_mesh_operator_subscription

.PHONY: uninstall-service-mesh-operator
## Uninstall the Service Mesh Operator
uninstall-service-mesh-operator:
	${Q}-${EC} uninstall_service_mesh_operator_subscription

.PHONY: install-all-operators
## Install all the operators
install-all-operators: install-backing-db-operator install-service-binding-operator install-serverless-operator install-service-mesh-operator

.PHONY: uninstall-all-operators
## Uninstall all the operators
uninstall-all-operators: uninstall-backing-db-operator uninstall-service-binding-operator uninstall-serverless-operator uninstall-service-mesh-operator

.PHONY: install-knative-serving
## Install Knative Serving
install-knative-serving:
	${Q}${EC} install_knative_serving

.PHONY: uninstall-knative-serving
## Uninstall Knative Serving
uninstall-knative-serving:
	${Q}-${EC} uninstall_knative_serving

.PHONY: install-quarkus-native-s2i-builder
## Install ubi-quarkus-native-s2i builder
install-quarkus-native-s2i-builder:
	${Q}${EC} install_ubi_quarkus_native_s2i_builder_image

.PHONY: install-all
## Install all the operators, Knative Serving and the Quarkus native s2i builder image
install-all: install-all-operators install-knative-serving install-quarkus-native-s2i-builder

## -- Application Developer targets --

.PHONY: create-project
## Create the OpenShift project/namespace
create-project:
	${Q}-${EC} create_project

.PHONY: delete-project
## Delete the OpenShift project/namespace
delete-project:
	${Q}${EC} delete_project

.PHONY: create-backing-db-instance
## Create the Backing Service Database
create-backing-db-instance:
	${Q}${EC} install_postgresql_db_instance

.PHONY: set-labels-on-knative-app
## Set binding labels on the knative application
set-labels-on-knative-app:
	${Q}oc label services.serving.knative.dev knative-app connects-to=postgres environment=demo --overwrite

.PHONY: create-service-binding-request
## Create the Service Binding Request
create-service-binding-request:
	${Q}oc apply -f service-binding-request.yaml
