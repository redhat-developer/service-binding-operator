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
	@oc apply -f service-binding-operator-source.yaml
	$(eval CSV_VERSION := $(shell oc get packagemanifest service-binding-operator -o jsonpath='{.status.channels[?(@.name == "stable")].currentCSV}'))
	@sed -e 's,REPLACE_CSV_VERSION,$(CSV_VERSION),g' service-binding-operator-subscription.yaml | oc apply -f -

.PHONY: install-backing-db-operator
## Install the Backing Service DB Operator
install-backing-db-operator:
	@oc apply -f backing-db-operator-source.yaml
	$(eval CSV_VERSION := $(shell oc get packagemanifest db-operators -o jsonpath='{.status.channels[?(@.name == "stable")].currentCSV}'))
	@sed -e 's,REPLACE_CSV_VERSION,$(CSV_VERSION),g' backing-db-operator-subscription.yaml | oc apply -f -

.PHONY: install-serverless-operator
## Install the Serverless Operator
install-serverless-operator:
	$(eval CSV_VERSION := $(shell oc get packagemanifest serverless-operator -o jsonpath='{.status.channels[?(@.name == "techpreview")].currentCSV}'))
	@sed -e 's,REPLACE_CSV_VERSION,$(CSV_VERSION),g' serverless-operator-subscription.yaml | oc apply -f -

.PHONY: install-service-mesh-operator
## Install the Service Mesh Operator
install-service-mesh-operator:
	$(eval CSV_VERSION := $(shell oc get packagemanifest servicemeshoperator -o jsonpath='{.status.channels[?(@.name == "1.0")].currentCSV}'))
	@sed -e 's,REPLACE_CSV_VERSION,$(CSV_VERSION),g' service-mesh-operator-subscription.yaml | oc apply -f -

.PHONY: install-all-operators
## Install all the operators
install-all-operators: install-backing-db-operator install-service-binding-operator install-serverless-operator install-service-mesh-operator

.PHONY: install-all
## Install all the operators and the quarkus native s2i builder image
install-all: install-all-operators install-quarkus-native-s2i-builder

.PHONY: install-quarkus-native-s2i-builder
## Install ubi-quarkus-native-s2i builder
install-quarkus-native-s2i-builder:
	@oc import-image quay.io/quarkus/ubi-quarkus-native-s2i:19.1.1 -n openshift --confirm
	@oc patch is ubi-quarkus-native-s2i -n openshift -p '{"spec": {"tags": [{"name" : "19.1.1", "annotations": {"tags": "builder"}}]}}'

## -- Application Developer targets --

.PHONY: create-project
## Create the OpenShift project/namespace
create-project:
	@oc new-project service-binding-demo

.PHONY: create-backing-db-instance
## Create the Backing Service Database
create-backing-db-instance:
	@oc apply -f create-db-instance.yaml

.PHONY: create-service-binding-request
## Create the Service Binding Request
create-service-binding-request:
	@oc apply -f service-binding-request.yaml

#PHONY: set-labels-on-knative-app
## Set the binding labels for the app
#set-labels-on-knative-app:
#	@oc project service-binding-demo
#	@oc patch dc knative-app -p '{"metadata": {"labels": {"connects-to": "postgres", "environment": "demo"}}}'

.PHONY: delete-project
## Delete the OpenShift project/namespace
delete-project:
	@-oc delete project service-binding-demo
