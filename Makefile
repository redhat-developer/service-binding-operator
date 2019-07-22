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

CGO_ENABLED ?= 0
GO111MODULE ?= on
GOCACHE ?= "$(shell echo ${PWD})/out/gocache"

GOCOV_FILE_TEMPL ?= ./out/coverage.REPLACE_TEST.txt
GOCOV ?= "-covermode=count -coverprofile REPLACE_FILE"

GIT_COMMIT_ID = $(shell git rev-parse --short HEAD)

OPERATOR_VERSION ?= 0.0.10
OPERATOR_GROUP ?= ${GO_PACKAGE_ORG_NAME}
OPERATOR_IMAGE ?= quay.io/${OPERATOR_GROUP}/${GO_PACKAGE_REPO_NAME}
OPERATOR_TAG_SHORT ?= $(OPERATOR_VERSION)
OPERATOR_TAG_LONG ?= $(OPERATOR_VERSION)-$(GIT_COMMIT_ID)

QUAY_TOKEN ?= ""

MANIFESTS_DIR ?= ./manifests
MANIFESTS_TMP ?= ./tmp/manifests

GOLANGCI_LINT_BIN=./out/golangci-lint

## -- Static code analysis (lint) targets --

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
	$(Q)GOCACHE=$(GOCACHE) ./out/golangci-lint ${V_FLAG} run --deadline=30m

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
	$(Q)kubectl --namespace $(TEST_NAMESPACE) apply -f ./test/third-party-crds/postgresql_v1alpha1_database_crd.yaml

.PHONY: e2e-cleanup
e2e-cleanup: get-test-namespace
	$(Q)-kubectl delete namespace $(TEST_NAMESPACE) --timeout=45s --wait

.PHONY: test-e2e
## Runs the e2e tests locally from test/e2e dir
test-e2e: e2e-setup
	$(info Running E2E test: $@)
	$(eval GOCOV_FILE := $(shell echo $(GOCOV_FILE_TEMPL) | sed -e 's,REPLACE_TEST,$(@),'))
	$(eval GOCOV_FLAGS := $(shell echo $(GOCOV) | sed -e 's,REPLACE_FILE,$(GOCOV_FILE),'))
	$(Q)GO111MODULE=$(GO111MODULE) GOCACHE=$(GOCACHE) SERVICE_BINDING_OPERATOR_DISABLE_ELECTION=true \
		operator-sdk --verbose test local ./test/e2e \
			--debug \
			--namespace $(TEST_NAMESPACE) \
			--up-local \
			--go-test-flags "-timeout=15m $(GOCOV_FLAGS)"
	$(Q)go tool cover -func=$(GOCOV_FILE)

.PHONY: test-unit
## Runs the unit tests
test-unit:
	$(info Running unit test: $@)
	$(eval GOCOV_FILE := $(shell echo $(GOCOV_FILE_TEMPL) | sed -e 's,REPLACE_TEST,$(@),'))
	$(eval GOCOV_FLAGS := $(shell echo $(GOCOV) | sed -e 's,REPLACE_FILE,$(GOCOV_FILE),'))
	$(Q)GO111MODULE=$(GO111MODULE) GOCACHE=$(GOCACHE) \
		go test $(shell GOCACHE="$(GOCACHE)" go list ./...|grep -v e2e) $(GOCOV_FLAGS) -v -mod vendor $(TEST_EXTRA_ARGS)
	$(Q)go tool cover -func=$(GOCOV_FILE)

.PHONY: test
## Test: Runs unit and integration (e2e) tests
test: test-unit test-e2e

.PHONY: test-e2e-olm-ci
## OLM-E2E: Adds the operator as a subscription, and run e2e tests without any setup.
test-e2e-olm-ci:
	$(Q)sed -e "s,REPLACE_IMAGE,registry.svc.ci.openshift.org/${OPENSHIFT_BUILD_NAMESPACE}/stable:service-binding-operator-registry," ./test/operator-hub/catalog_source.yaml | kubectl apply -f -
	$(Q)kubectl apply -f ./test/operator-hub/subscription.yaml
	$(eval DEPLOYED_NAMESPACE := openshift-operators)
	$(eval GOCOV_FILE := $(shell echo $(GOCOV_FILE_TEMPL) | sed -e 's,REPLACE_TEST,$(@),'))
	$(eval GOCOV_FLAGS := $(shell echo $(GOCOV) | sed -e 's,REPLACE_FILE,$(GOCOV_FILE),'))
	$(Q)./hack/check-crds.sh
	$(Q)operator-sdk --verbose test local ./test/e2e --no-setup --go-test-flags "-timeout=15m $(GOCOV_FLAGS)"
	$(Q)go tool cover -func=$(GOCOV_FILE)

## -- Build Go binary and OCI image targets --

.PHONY: build
## Build: compile the operator for Linux/AMD64.
build: out/operator

out/operator:
	$(Q)GOARCH=amd64 GOOS=linux go build ${V_FLAG} -o ./out/operator cmd/manager/main.go

## Build-Image: using operator-sdk to build a new image
build-image:
	$(Q)operator-sdk build --image-builder=buildah "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)"

## Generate-K8S: after modifying _types, generate Kubernetes scaffolding.
generate-k8s:
	$(Q)GOCACHE=$(GOCACHE) operator-sdk generate k8s

## Generate-OpenAPI: after modifying _types, generate OpenAPI scaffolding.
generate-openapi:
	$(Q)GOCACHE=$(GOCACHE) operator-sdk generate openapi

## Vendor: 'go mod vendor' resets the vendor folder to what is defined in go.mod.
vendor: go.mod go.sum
	$(Q)GOCACHE=$(GOCACHE) go mod vendor ${V_FLAG}

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
	operator-courier --verbose flatten $(MANIFESTS_DIR) $(MANIFESTS_TMP)
	cp -vf deploy/crds/*_crd.yaml $(MANIFESTS_TMP)
	sed -i -e 's,REPLACE_IMAGE,"$(OPERATOR_IMAGE):latest",g' $(MANIFESTS_TMP)/*.yaml
	sed -i -e 's,REPLACE_ICON_BASE64_DATA,$(ICON_BASE64_DATA),' $(MANIFESTS_TMP)/*.yaml
	operator-courier --verbose verify $(MANIFESTS_TMP)

.PHONY: push-operator
## Push-Operator: Uplaod operator to Quay.io application repository
push-operator: prepare-csv
	operator-courier push $(MANIFESTS_TMP) $(OPERATOR_GROUP) $(GO_PACKAGE_REPO_NAME) $(OPERATOR_VERSION) "$(QUAY_TOKEN)"

## Push-Image: push container image to upstream, including latest tag.
push-image: build-image
	podman tag "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)" "$(OPERATOR_IMAGE):latest"
	podman push "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)"
	podman push "$(OPERATOR_IMAGE):latest"

## -- Local deployment targets --

.PHONY: local
## Local: Run operator locally
local: deploy-clean deploy-rbac deploy-crds deploy-cr
	$(Q)operator-sdk up local

.PHONY: deploy-rbac
## Deploy-RBAC: Setup service account and deploy RBAC
deploy-rbac:
	$(Q)kubectl create -f deploy/service_account.yaml
	$(Q)kubectl create -f deploy/role.yaml
	$(Q)kubectl create -f deploy/role_binding.yaml

.PHONY: deploy-crds
## Deploy-CRD: Deploy CRD
deploy-crds:
	$(Q)kubectl create -f deploy/crds/apps_v1alpha1_servicebindingrequest_crd.yaml

.PHONY: deploy-cr
## Deploy-CR: Deploy CRs
deploy-cr:
	$(Q)kubectl apply -f deploy/crds/apps_v1alpha1_servicebindingrequest_cr.yaml

.PHONY: deploy-clean
## Deploy-Clean: Removing CRDs and CRs
deploy-clean:
	$(Q)-kubectl delete -f deploy/crds/apps_v1alpha1_servicebindingrequest_cr.yaml
	$(Q)-kubectl delete -f deploy/crds/apps_v1alpha1_servicebindingrequest_crd.yaml
	$(Q)-kubectl delete -f deploy/operator.yaml
	$(Q)-kubectl delete -f deploy/role_binding.yaml
	$(Q)-kubectl delete -f deploy/role.yaml
	$(Q)-kubectl delete -f deploy/service_account.yaml

.PHONY: deploy
## Deploy:
deploy: deploy-rbac deploy-crds


## -- Cleanup targets --

.PHONY: clean
## Removes temp directories
clean:
	$(Q)-rm -rf ${V_FLAG} ./out
