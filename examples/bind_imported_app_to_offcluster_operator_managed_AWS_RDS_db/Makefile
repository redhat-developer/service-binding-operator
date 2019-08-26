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

.PHONY: install-service-binding-operator-source
## Install the Service Binding Operator
install-service-binding-operator-source:
	@oc apply -f operator-source.service-binding-operator.yaml

.PHONY: install-aws-rds-operator-source
## Install the AWS RDS operator
install-aws-rds-operator-source:
	@oc apply -f operator-source.aws-rds-operator.yaml

.PHONY: install-operator-secrets
## Install the AWS RDS operator secrets
install-operator-secrets: install-aws-rds-operator-secrets

.PHONY: install-aws-rds-operator-secrets
## Install the AWS RDS operator secrets
install-aws-rds-operator-secrets:
	$(eval AWS_ACCESS_KEY_ID := $(shell cat ~/.aws/credentials | grep aws_access_key_id | cut -d " " -f3))
	$(eval AWS_SECRET_ACCESS_KEY := $(shell cat ~/.aws/credentials | grep aws_secret_access_key | cut -d " " -f3))
	@sed -e 's,REPLACE_AWS_ACCESS_KEY_ID,$(shell echo -n "$(AWS_ACCESS_KEY_ID)" | base64),g' secret.aws-rds-operator.yaml | \
	 sed -e 's,REPLACE_AWS_SECRET_ACCESS_KEY,$(shell echo -n "$(AWS_SECRET_ACCESS_KEY)" | base64),g' | oc apply -f -

## -- Application Developer targets --

.PHONY: create-project
## Create the OpenShift project/namespace
create-project:
	@oc new-project service-binding-demo

.PHONY: deploy-shell-app
## Deploy the shell application
deploy-shell-app:
	@cat image-stream.shell-app.yaml build-config.shell-app.yaml deployment-config.shell-app.yaml | oc apply -f -

.PHONY: deploy-mydb
## Deploy the database that supports the applications
deploy-mydb:
	@cat secret.mydb.yaml rds.mydb.yaml | oc apply -f -

.PHONY: create-service-binding-request-shell-app
## Create the Service Binding Request for the shell app
create-service-binding-request-shell-app:
	@oc apply -f service-binding-request.shell-app.yaml

.PHONY: set-labels-on-nodejs-app
## Set the binding labels for the NodeJS app
set-labels-on-nodejs-app:
	@oc patch dc nodejs-app -p '{"metadata": {"labels": {"connects-to": "postgres", "environment": "nodejs"}}}'

.PHONY: create-service-binding-request-nodejs-app
## Create the Service Binding Request for the NodeJS app
create-service-binding-request-nodejs-app:
	@oc apply -f service-binding-request.nodejs-app.yaml

.PHONY: delete-project
## Delete the OpenShift project/namespace
delete-project:
	@-oc delete project service-binding-demo
