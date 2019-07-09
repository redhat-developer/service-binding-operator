.DEFAULT_GOAL := help

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
	X_FLAG = -x
endif

# Create output directory for artifacts and test results. ./out is supposed to
# be a safe place for all targets to write to while knowing that all content
# inside of ./out is wiped once "make clean" is run.
$(shell mkdir -p ./out);

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


#-----------------------------------------------------------------------------
# Global Variables
#-----------------------------------------------------------------------------

# By default the project should be build under GOPATH/src/github.com/<orgname>/<reponame>
GO_PACKAGE_ORG_NAME ?= $(shell basename $$(dirname $$PWD))
GO_PACKAGE_REPO_NAME ?= $(shell basename $$PWD)
GO_PACKAGE_PATH ?= github.com/${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}

GIT_COMMIT_ID = $(shell git rev-parse --short HEAD)

OPERATOR_STABLE_VERSION ?= 0.0.10
OPERATOR_DEV_VERSION ?= 0.0.11
OPERATOR_GROUP ?= ${GO_PACKAGE_ORG_NAME}
OPERATOR_PACKAGE ?= service-binding
OPERATOR_IMAGE ?= quay.io/${OPERATOR_GROUP}/${GO_PACKAGE_REPO_NAME}
OPERATOR_TAG_SHORT ?= $(OPERATOR_DEV_VERSION)
OPERATOR_TAG_LONG ?= $(OPERATOR_DEV_VERSION)-$(GIT_COMMIT_ID)
QUAY_TOKEN ?= ""

MANIFESTS_DIR ?= ./manifests
MANIFESTS_TMP ?= ./tmp/manifests

## -- Static code analysis (lint) targets --

GOLANGCI_LINT_BIN=./out/golangci-lint
.PHONY: lint
## Runs linters on Go code files and YAML files
lint: setup-venv lint-go-code lint-yaml courier

YAML_FILES := $(shell find . -path ./vendor -prune -o -type f -regex ".*y[a]ml" -print)
.PHONY: lint-yaml
## runs yamllint on all yaml files
lint-yaml: ${YAML_FILES}
	$(Q)./out/venv3/bin/pip install yamllint
	$(Q)./out/venv3/bin/yamllint -c .yamllint $(YAML_FILES)

.PHONY: lint-go-code
## Checks the code with golangci-lint
lint-go-code: $(GOLANGCI_LINT_BIN)
	# This is required for OpenShift CI enviroment
	# Ref: https://github.com/openshift/release/pull/3438#issuecomment-482053250
	$(Q)GOCACHE=$(shell pwd)/out/gocache ./out/golangci-lint ${V_FLAG} run --deadline=30m

$(GOLANGCI_LINT_BIN):
	$(Q)curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./out v1.17.1

.PHONY: courier
## Validate manifests using operator-courier
courier:
	$(Q)./out/venv3/bin/pip install operator-courier
	$(Q)./out/venv3/bin/operator-courier flatten ./manifests ./out/manifests
	$(Q)./out/venv3/bin/operator-courier verify ./out/manifests

.PHONY: setup-venv
## Setup virtual environment
setup-venv:
	$(Q)python3 -m venv ./out/venv3
	$(Q)./out/venv3/bin/pip install --upgrade setuptools
	$(Q)./out/venv3/bin/pip install --upgrade pip

## -- Test targets --

# Generate namespace name for test
out/test-namespace:
	@echo -n "test-namespace-$(shell uuidgen | tr '[:upper:]' '[:lower:]')" > ./out/test-namespace

.PHONY: get-test-namespace
get-test-namespace: out/test-namespace
	$(eval TEST_NAMESPACE := $(shell cat ./out/test-namespace))

# E2E test
.PHONY: e2e-setup
e2e-setup: e2e-cleanup
	$(Q)kubectl create namespace $(TEST_NAMESPACE)

.PHONY: e2e-cleanup
e2e-cleanup: get-test-namespace
	$(Q)-kubectl delete namespace $(TEST_NAMESPACE) --timeout=10s --wait

.PHONY: test-e2e
## Runs the e2e tests locally from test/e2e dir
test-e2e: e2e-setup
	$(info Running E2E test: $@)
	$(Q)GO111MODULE=on operator-sdk test local ./test/e2e --namespace $(TEST_NAMESPACE) --up-local --go-test-flags "-v -timeout=15m"

.PHONY: test-unit
## Runs the unit tests
test-unit:
	$(info Running unit test: $@)
	$(Q)GO111MODULE=on GOCACHE=$(shell pwd)/out/gocache go test $(shell GOCACHE=$(shell pwd)/out/gocache go list ./...|grep -v e2e) -v -mod vendor $(TEST_EXTRA_ARGS)

.PHONY: test-e2e-olm-ci
test-e2e-olm-ci:
	$(Q)sed -e "s,REPLACE_IMAGE,registry.svc.ci.openshift.org/${OPENSHIFT_BUILD_NAMESPACE}/stable:service-binding-operator-registry," ./test/e2e/catalog_source.yaml | kubectl apply -f -
	$(Q)kubectl apply -f ./test/e2e/subscription.yaml
	$(eval DEPLOYED_NAMESPACE := openshift-operators)
	$(Q)./hack/check-crds.sh
	$(Q)operator-sdk test local ./test/e2e --no-setup --go-test-flags "-v -timeout=15m"

## -- Build Go binary and OCI image targets --

.PHONY: build 
## Build: compile the operator for Linux/AMD64.
build: out/operator

out/operator:
	$(Q)CGO_ENABLED=0 GO111MODULE=on GOARCH=amd64 GOOS=linux go build ${V_FLAG} -o ./out/operator cmd/manager/main.go

## Build-Image: using operator-sdk to build a new image
build-image:
	$(Q)GO111MODULE=on operator-sdk build "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)"

## Vendor: 'go mod vendor' resets the vendor folder to what is defined in go.mod.
vendor: go.mod go.sum
	$(Q)GOCACHE=$(shell pwd)/out/gocache GO111MODULE=on go mod vendor ${V_FLAG}

## Generate CSV: using oeprator-sdk generate cluster-service-version for current operator version
generate-csv:
	operator-sdk olm-catalog gen-csv --csv-version=$(OPERATOR_VERSION) --verbose

generate-olm:
	operator-courier --verbose flatten $(MANIFESTS_DIR) $(MANIFESTS_TMP)
	cp -vf deploy/crds/*_crd.yaml $(MANIFESTS_TMP)

## -- Publish image and manifests targets --

## Prepare-CSV: using a temporary location copy all operator CRDs and metadata to generate a CSV.
prepare-csv: build-image
	$(eval ICON_BASE64_DATA := $(shell cat ./assets/icon/red-hat-logo.png | base64))
	@rm -rf $(MANIFESTS_TMP) || true
	@mkdir -p ${MANIFESTS_TMP}
	cp -vf $(MANIFESTS_DIR)/stable/* $(MANIFESTS_TMP)
	cp -vf $(MANIFESTS_DIR)/dev/* $(MANIFESTS_TMP)
	cp -vf $(MANIFESTS_DIR)/*.package.yaml $(MANIFESTS_TMP)
	cp -vf deploy/crds/*_crd.yaml $(MANIFESTS_TMP)
	sed -i -e 's,REPLACE_IMAGE,$(OPERATOR_IMAGE),g' $(MANIFESTS_TMP)/*.yaml
	sed -i -e 's,REPLACE_PACKAGE,$(OPERATOR_PACKAGE),g' $(MANIFESTS_TMP)/*.yaml
	sed -i -e 's,REPLACE_VERSION,$(OPERATOR_TAG_LONG),g' $(MANIFESTS_TMP)/*.yaml
	sed -i -e 's,REPLACE_ICON_BASE64_DATA,$(ICON_BASE64_DATA),' $(MANIFESTS_TMP)/*.yaml
	operator-courier --verbose verify $(MANIFESTS_TMP)

.PHONY: push-operator
## Push-Operator: Uplaod operator to Quay.io application repository
push-operator: prepare-csv
	$(Q)operator-courier push $(MANIFESTS_TMP) $(OPERATOR_GROUP) $(OPERATOR_PACKAGE) $(OPERATOR_TAG_LONG) "$(QUAY_TOKEN)"

## Push-Image: push docker image to upstream, including latest tag.
push-image: build-image
	docker tag "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)" "$(OPERATOR_IMAGE):latest"
	docker push "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)"
	docker push "$(OPERATOR_IMAGE):latest"

## -- Local deployment targets --

.PHONY: local
## Run operator locally
local: deploy-clean deploy-rbac deploy-crds deploy-cr
	$(Q)operator-sdk up local

.PHONY: deploy-rbac
## Setup service account and deploy RBAC
deploy-rbac:
	$(Q)kubectl create -f deploy/service_account.yaml
	$(Q)kubectl create -f deploy/role.yaml
	$(Q)kubectl create -f deploy/role_binding.yaml

.PHONY: deploy-crds
## Deploy CRD
deploy-crds:
	$(Q)kubectl create -f deploy/crds/apps_v1alpha1_servicebindingrequest_crd.yaml

.PHONY: deploy-cr
## Deploy CRs
deploy-cr:
	$(Q)kubectl apply -f deploy/crds/apps_v1alpha1_servicebindingrequest_cr.yaml

.PHONY: deploy-clean
## Removing CRDs and CRs
deploy-clean:
	$(Q)-kubectl delete -f deploy/crds/apps_v1alpha1_servicebindingrequest_cr.yaml
	$(Q)-kubectl delete -f deploy/crds/apps_v1alpha1_servicebindingrequest_crd.yaml
	$(Q)-kubectl delete -f deploy/operator.yaml
	$(Q)-kubectl delete -f deploy/role_binding.yaml
	$(Q)-kubectl delete -f deploy/role.yaml
	$(Q)-kubectl delete -f deploy/service_account.yaml


## -- Cleanup targets --

.PHONY: clean
## Removes temp directories
clean:
	$(Q)-rm -rf ${V_FLAG} ./out ./tmp
