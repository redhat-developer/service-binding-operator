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
ZAP_ENCODER_FLAG = --zap-level=debug --zap-encoder=console
ZAP_LEVEL_FLAG =
VERBOSE_FLAG =
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
	ZAP_LEVEL_FLAG = --zap-level 1
endif
ifeq ($(VERBOSE),3)
	Q_FLAG =
	QUIET_FLAG =
	S_FLAG =
	V_FLAG = -v
	VERBOSE_FLAG = --verbose
	X_FLAG = -x
	ZAP_LEVEL_FLAG = --zap-level 2
endif

ZAP_FLAGS = $(ZAP_ENCODER_FLAG) $(ZAP_LEVEL_FLAG)

SKIP_CLEANUP_ERROR ?= true

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

PROJECT_DIR ?= $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
SBR_MANIFESTS ?= ${PROJECT_DIR}service-binding-operator-manifests

CGO_ENABLED ?= 0
GO111MODULE ?= on
GOCACHE ?= "$(shell echo ${PWD})/out/gocache"
GOFLAGS ?= -mod=vendor
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)

# This variable is for artifacts to be archived by Prow jobs at OpenShift CI
# The actual value will be set by the OpenShift CI accordingly
ARTIFACT_DIR ?= "$(shell echo ${PWD})/out"

GOCOV_DIR ?= $(ARTIFACT_DIR)/test-coverage
GOCOV_FILE_TEMPL ?= $(GOCOV_DIR)/REPLACE_TEST.txt
GOCOV ?= "-covermode=atomic -coverprofile REPLACE_FILE"

GIT_COMMIT_ID ?= $(shell git rev-parse --short=8 HEAD)

OPERATOR_VERSION ?= 0.4.0
OPERATOR_REGISTRY ?= quay.io
OPERATOR_REPO_REF ?= $(OPERATOR_REGISTRY)/redhat-developer/servicebinding-operator
OPERATOR_TAG_SHORT ?= $(OPERATOR_VERSION)
OPERATOR_TAG_LONG ?= $(OPERATOR_VERSION)-$(GIT_COMMIT_ID)
OPERATOR_IMAGE_BUILDER ?= buildah
OPERATOR_SDK_EXTRA_ARGS ?= "--debug"
BASE_BUNDLE_VERSION ?= $(OPERATOR_VERSION)
BUNDLE_VERSION ?= $(BASE_BUNDLE_VERSION)-$(GIT_COMMIT_ID)
OPERATOR_BUNDLE_IMAGE_REF ?= $(OPERATOR_REPO_REF):bundle-$(BUNDLE_VERSION)
OPERATOR_INDEX_IMAGE_REF ?= $(OPERATOR_REPO_REF):index
OPERATOR_IMAGE_REF ?= $(OPERATOR_REPO_REF):$(GIT_COMMIT_ID)
CSV_PACKAGE_NAME ?= $(GO_PACKAGE_REPO_NAME)
CSV_CREATION_TIMESTAMP ?= $(shell TZ=GMT date '+%FT%TZ')

QUAY_USERNAME ?= redhat-developer+travis
REGISTRY_USERNAME ?= $(QUAY_USERNAME)
QUAY_TOKEN ?= ""
REGISTRY_PASSWORD ?= $(QUAY_TOKEN)


MANIFESTS_DIR ?= $(shell echo ${PWD})/tmp/manifests/$(BASE_BUNDLE_VERSION)
MANIFESTS_BUNDLE_DIR ?= $(MANIFESTS_DIR)/$(GIT_COMMIT_ID)
HACK_DIR ?= $(shell echo ${PWD})/hack
OUTPUT_DIR ?= $(shell echo ${PWD})/out
OLM_CATALOG_DIR ?= $(shell echo ${PWD})/deploy/olm-catalog
CRDS_DIR ?= $(shell echo ${PWD})/deploy/crds
LOGS_DIR ?= $(OUTPUT_DIR)/logs
GOLANGCI_LINT_BIN=$(OUTPUT_DIR)/golangci-lint
PYTHON_VENV_DIR=$(OUTPUT_DIR)/venv3

CONTAINER_RUNTIME ?= docker

# -- Variables for uploading code coverage reports to Codecov.io --
# This default path is set by the OpenShift CI
CODECOV_TOKEN_PATH ?= "/usr/local/redhat-developer-service-binding-operator-codecov-token/token"
REPO_OWNER := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].org')
REPO_NAME := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].repo')
BASE_COMMIT := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].base_sha')
PR_COMMIT := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].pulls[0].sha')
PULL_NUMBER := $(shell echo $$CLONEREFS_OPTIONS | jq '.refs[0].pulls[0].number')

# -- Variables for acceptance tests
TEST_ACCEPTANCE_START_SBO ?= local
TEST_ACCEPTANCE_OUTPUT_DIR ?= $(OUTPUT_DIR)/acceptance-tests
TEST_ACCEPTANCE_REPORT_DIR ?= $(OUTPUT_DIR)/acceptance-tests-report
TEST_ACCEPTANCE_ARTIFACTS ?= /tmp/artifacts

TEST_ACCEPTANCE_TAGS ?=

ifdef TEST_ACCEPTANCE_TAGS
TEST_ACCEPTANCE_TAGS_ARG := --tags="~@disabled" --tags="$(TEST_ACCEPTANCE_TAGS)"
else
TEST_ACCEPTANCE_TAGS_ARG := --tags="~@disabled"
endif

## -- Static code analysis (lint) targets --

.PHONY: lint
## Runs linters on Go code files and YAML files - DISABLED TEMPORARILY
lint: setup-venv lint-go-code lint-yaml lint-python-code lint-feature-files

YAML_FILES := $(shell find . -path ./vendor -prune -o -type f -regex ".*\.y[a]ml" -print)
.PHONY: lint-yaml
## runs yamllint on all yaml files
lint-yaml: ${YAML_FILES}
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install yamllint==1.23.0
	$(Q)$(PYTHON_VENV_DIR)/bin/yamllint -c .yamllint $(YAML_FILES)

.PHONY: lint-go-code
## Checks the code with golangci-lint
lint-go-code: $(GOLANGCI_LINT_BIN)
	# This is required for OpenShift CI enviroment
	# Ref: https://github.com/openshift/release/pull/3438#issuecomment-482053250
	$(Q)GOFLAGS="$(GOFLAGS)" GOCACHE="$(GOCACHE)" $(OUTPUT_DIR)/golangci-lint ${V_FLAG} run --deadline=30m

$(GOLANGCI_LINT_BIN):
	$(Q)curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./out v1.18.0

## -- Check the python code
.PHONY: lint-python-code
lint-python-code: setup-venv
	$(Q)PYTHON_VENV_DIR=$(PYTHON_VENV_DIR) ./hack/check-python/lint-python-code.sh

## -- Check the acceptance tests feature files
.PHONY: lint-feature-files
lint-feature-files:
	$(Q)./hack/check-feature-files.sh

.PHONY: setup-venv
## Setup virtual environment
setup-venv:
	$(Q)python3 -m venv $(PYTHON_VENV_DIR)
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install --upgrade setuptools
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install --upgrade pip

## -- Test targets --

# Generate namespace name for test
out/test-namespace:
	@echo -n "test-namespace-$(shell uuidgen | tr '[:upper:]' '[:lower:]' | head -c 8)" > $(OUTPUT_DIR)/test-namespace

.PHONY: get-test-namespace
get-test-namespace: out/test-namespace
	$(eval TEST_NAMESPACE := $(shell cat $(OUTPUT_DIR)/test-namespace))

# Testing setup
.PHONY: deploy-test-3rd-party-crds
deploy-test-3rd-party-crds: get-test-namespace
	$(Q)kubectl --namespace $(TEST_NAMESPACE) apply -f ./test/third-party-crds/

.PHONY: create-test-namespace
create-test-namespace:
	$(Q)kubectl create namespace $(TEST_NAMESPACE)

.PHONY: test-setup
test-setup: test-cleanup create-test-namespace deploy-test-3rd-party-crds

.PHONY: test-cleanup
test-cleanup: get-test-namespace
	$(Q)-TEST_NAMESPACE=$(TEST_NAMESPACE) $(HACK_DIR)/test-cleanup.sh

.PHONY: test-unit
## Runs the unit tests without code coverage
test-unit:
	$(info Running unit test: $@)
	$(Q)GO111MODULE=$(GO111MODULE) GOCACHE="$(GOCACHE)" \
		go test $(shell GOFLAGS="$(GOFLAGS)" GOCACHE="$(GOCACHE)" go list ./...) -v $(GOFLAGS) $(TEST_EXTRA_ARGS)

.PHONY: test-unit-with-coverage
## Runs the unit tests with code coverage
test-unit-with-coverage:
	$(info Running unit test: $@)
	$(eval GOCOV_FILE := $(shell echo $(GOCOV_FILE_TEMPL) | sed -e 's,REPLACE_TEST,$(@),'))
	$(eval GOCOV_FLAGS := $(shell echo $(GOCOV) | sed -e 's,REPLACE_FILE,$(GOCOV_FILE),'))
	$(Q)mkdir -p $(GOCOV_DIR)
	$(Q)rm -vf '$(GOCOV_DIR)/*.txt'
	$(Q)GO111MODULE=$(GO111MODULE) GOCACHE="$(GOCACHE)" \
		go test $(shell GOFLAGS="$(GOFLAGS)" GOCACHE="$(GOCACHE)" go list ./...) $(GOCOV_FLAGS) -v $(GOFLAGS) $(TEST_EXTRA_ARGS)
	$(Q)GOFLAGS="$(GOFLAGS)" GOCACHE="$(GOCACHE)" go tool cover -func=$(GOCOV_FILE)

.PHONY: test-acceptance-setup
## Setup the environment for the acceptance tests
test-acceptance-setup: setup-venv
ifeq ($(TEST_ACCEPTANCE_START_SBO), local)
test-acceptance-setup: stop-local build test-cleanup create-test-namespace deploy-test-3rd-party-crds deploy-rbac deploy-crds
	$(Q)echo "Starting local SBO instance"
	$(eval TEST_ACCEPTANCE_SBO_STARTED := $(shell OPERATOR_NAMESPACE="$(TEST_NAMESPACE)" ZAP_FLAGS="$(ZAP_FLAGS)" OUTPUT="$(TEST_ACCEPTANCE_OUTPUT_DIR)" RUN_IN_BACKGROUND=true ./hack/deploy-sbo-local.sh))
else ifeq ($(TEST_ACCEPTANCE_START_SBO), remote)
test-acceptance-setup: test-cleanup create-test-namespace
	$(Q)echo "Using remote SBO instance running in '$(SBO_NAMESPACE)' namespace"
else ifeq ($(TEST_ACCEPTANCE_START_SBO), operator-hub)
test-acceptance-setup:
	$(eval TEST_ACCEPTANCE_SBO_STARTED := $(shell ./hack/deploy-sbo-operator-hub.sh))
endif
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install -q -r test/acceptance/features/requirements.txt

.PHONY: test-acceptance
## Runs acceptance tests
test-acceptance: test-acceptance-setup
	$(Q)echo "Running acceptance tests"
	$(Q)TEST_ACCEPTANCE_START_SBO=$(TEST_ACCEPTANCE_START_SBO) \
		TEST_ACCEPTANCE_SBO_STARTED=$(TEST_ACCEPTANCE_SBO_STARTED) \
		TEST_NAMESPACE=$(TEST_NAMESPACE) \
		SBO_NAMESPACE=$(SBO_NAMESPACE) \
		$(PYTHON_VENV_DIR)/bin/behave --junit --junit-directory $(TEST_ACCEPTANCE_OUTPUT_DIR) $(V_FLAG) --no-capture --no-capture-stderr $(TEST_ACCEPTANCE_TAGS_ARG) $(EXTRA_BEHAVE_ARGS) test/acceptance/features
ifeq ($(TEST_ACCEPTANCE_START_SBO), local)
	$(Q)kill $(TEST_ACCEPTANCE_SBO_STARTED)
endif

.PHONY: test-acceptance-smoke
## Runs a sub-set of acceptance tests tagged with @smoke tag
test-acceptance-smoke:
	$(Q)TEST_ACCEPTANCE_TAGS=@smoke $(MAKE) test-acceptance

.PHONY: test-acceptance-generate-report
## Generate acceptance tests report
test-acceptance-generate-report:
	$(Q)CONTAINER_RUNTIME=$(CONTAINER_RUNTIME) $(HACK_DIR)/allure-report.sh generate

.PHONY: test-acceptance-serve-report
## Serves acceptance tests report at http://localhost:8088
test-acceptance-serve-report:
	$(Q)CONTAINER_RUNTIME=$(CONTAINER_RUNTIME) $(HACK_DIR)/allure-report.sh serve

.PHONY: test-acceptance-artifacts
## Collect artifacts from acceptance tests to be archived in CI
test-acceptance-artifacts:
	$(Q)echo "Gathering acceptance tests artifacts"
	$(Q)mkdir -p $(TEST_ACCEPTANCE_ARTIFACTS) \
		&& cp -rvf $(TEST_ACCEPTANCE_OUTPUT_DIR) $(TEST_ACCEPTANCE_ARTIFACTS)/

.PHONY: test
## Test: Runs unit and acceptance tests
test: test-unit test-acceptance

## -- Build Go binary and OCI image targets --

.PHONY: build
## Build: compile the operator binary
build: out/operator

out/operator:
	$(Q)GOCACHE="$(GOCACHE)" GOARCH=$(ARCH) GOOS=$(OS) go build $(GOFLAGS) ${V_FLAG} -o $(OUTPUT_DIR)/operator cmd/manager/main.go

## Generate-K8S: after modifying _types, generate Kubernetes scaffolding.
generate-k8s:
	$(Q)GOFLAGS="$(GOFLAGS)" operator-sdk generate k8s

$(OUTPUT_DIR)/openapi-gen:
	$(Q)GOCACHE="$(GOCACHE)" go build $(GOFLAGS) -o $(OUTPUT_DIR)/openapi-gen k8s.io/kube-openapi/cmd/openapi-gen

## Generate-OpenAPI: after modifying _types, generate OpenAPI scaffolding.
generate-openapi: $(OUTPUT_DIR)/openapi-gen
	# Build the latest openapi-gen from source
	$(Q)GOFLAGS="$(GOFLAGS)" GOCACHE="$(GOCACHE)" $(OUTPUT_DIR)/openapi-gen --logtostderr=true -o "" -i $(GO_PACKAGE_PATH)/pkg/apis/operators/v1alpha1 -O zz_generated.openapi -p ./pkg/apis/operators/v1alpha1 -h ./hack/boilerplate.go.txt -r "-"

## Vendor: 'go mod vendor' resets the vendor folder to what is defined in go.mod.
vendor: go.mod go.sum
	$(Q)GOCACHE="$(GOCACHE)" go mod vendor ${V_FLAG}

## -- Local deployment targets --

.PHONY: local
## Local: Run operator locally
local: stop-local build get-test-namespace deploy-clean deploy-rbac deploy-crds
	$(Q)OPERATOR_NAMESPACE="$(TEST_NAMESPACE)" ZAP_FLAGS="$(ZAP_FLAGS)" OUTPUT="$(TEST_ACCEPTANCE_OUTPUT_DIR)" ./hack/deploy-sbo-local.sh

.PHONY: stop-local
## Stop Local: Stop locally running operator
stop-local:
	$(Q)-./hack/stop-sbo-local.sh

.PHONY: deploy-rbac
## Deploy-RBAC: Setup service account and deploy RBAC
deploy-rbac:
	$(Q)kubectl apply -f deploy/service_account.yaml -n $(TEST_NAMESPACE)
	$(Q)kubectl apply -f deploy/role.yaml -n $(TEST_NAMESPACE)
	$(Q)kubectl apply -f deploy/role_binding.yaml -n $(TEST_NAMESPACE)

.PHONY: deploy-crds
## Deploy-CRD: Deploy CRD
deploy-crds:
	$(Q)kubectl apply -f deploy/crds/operators.coreos.com_servicebindings_crd.yaml

.PHONY: deploy-clean
## Deploy-Clean: Removing CRDs and CRs
deploy-clean:
	$(Q)-$(HACK_DIR)/deploy-clean.sh

.PHONY: deploy
## Deploy:
deploy: deploy-rbac deploy-crds


## -- Cleanup targets --

.PHONY: clean
## Removes temp directories
clean:
	$(Q)-rm -rf ${V_FLAG} $(OUTPUT_DIR)


## -- Targets for uploading code coverage reports to Codecov.io --

.PHONY: upload-codecov-report
## Uploads the test coverage reports to codecov.io.
## DO NOT USE LOCALLY: must only be called by OpenShift CI when processing new PR and when a PR is merged!
upload-codecov-report:
ifneq ($(PR_COMMIT), null)
	@echo "uploading test coverage report for pull-request #$(PULL_NUMBER)..."
	@/bin/bash <(curl -s https://codecov.io/bash) \
		-t $(shell tr -d ' \n' <$CODECOV_TOKEN_PATH) \
		-f $(GOCOV_DIR)/*.txt \
		-C $(PR_COMMIT) \
		-r $(REPO_OWNER)/$(REPO_NAME) \
		-P $(PULL_NUMBER) \
		-Z > codecov-upload.log
else
	@echo "uploading test coverage report after PR was merged..."
	@/bin/bash <(curl -s https://codecov.io/bash) \
		-t $(shell tr -d ' \n' <$CODECOV_TOKEN_PATH) \
		-f $(GOCOV_DIR)/*.txt \
		-C $(BASE_COMMIT) \
		-r $(REPO_OWNER)/$(REPO_NAME) \
		-Z > codecov-upload.log
endif


## -- Bundle validation, push and release targets

.PHONY: registry-login
registry-login:
	@$(CONTAINER_RUNTIME) login -u "$(REGISTRY_USERNAME)" --password-stdin $(OPERATOR_REGISTRY) <<<"$(REGISTRY_PASSWORD)"

.PHONY: build-operator-image
## Make a dev release on every merge to master
build-operator-image:
	$(Q)$(CONTAINER_RUNTIME) build -f Dockerfile.rhel -t $(OPERATOR_IMAGE_REF) .

.PHONY: push-operator-image
## push operator image to quay
push-operator-image: build-operator-image registry-login
	$(CONTAINER_RUNTIME) push "$(OPERATOR_IMAGE_REF)"

.PHONY: prepare-operator-manifests
## Prepare operator manifests
prepare-operator-manifests: push-operator-image
	@rm -rf $(MANIFESTS_DIR) || true
	@mkdir -p $(MANIFESTS_BUNDLE_DIR)
	operator-sdk generate csv --csv-version $(BUNDLE_VERSION) $(VERBOSE_FLAG) --from-version=0.0.23
	cp -vrf $(OLM_CATALOG_DIR)/$(GO_PACKAGE_REPO_NAME)/$(BUNDLE_VERSION)/* $(MANIFESTS_BUNDLE_DIR)
	cp -vrf $(OLM_CATALOG_DIR)/$(GO_PACKAGE_REPO_NAME)/*package.yaml $(MANIFESTS_DIR)
	cp -vrf $(CRDS_DIR)/*_crd.yaml $(MANIFESTS_BUNDLE_DIR)
	$(eval OPERATOR_IMAGE_SHA := $(shell $(CONTAINER_RUNTIME) inspect --format='{{index .RepoDigests 0}}' $(OPERATOR_IMAGE_REF)))
	sed -i -e 's,REPLACE_IMAGE,$(OPERATOR_IMAGE_SHA),g' $(MANIFESTS_BUNDLE_DIR)/*.clusterserviceversion.yaml
	sed -i -e 's,CSV_CREATION_TIMESTAMP,$(CSV_CREATION_TIMESTAMP),g' $(MANIFESTS_BUNDLE_DIR)/*.clusterserviceversion.yaml
	awk -i inplace '!/^[[:space:]]+replaces:[[:space:]]+[[:graph:]]+/ { print $0 }' $(MANIFESTS_BUNDLE_DIR)/*.clusterserviceversion.yaml
	sed -i -e 's,BUNDLE_VERSION,$(BUNDLE_VERSION),g' $(MANIFESTS_DIR)/*.yaml
	sed -i -e 's,PACKAGE_NAME,$(CSV_PACKAGE_NAME),g' $(MANIFESTS_DIR)/*.yaml
	sed -i -e 's,ICON_BASE64_DATA,$(shell base64 --wrap=0 ./assets/icon/red-hat-logo.svg),g' $(MANIFESTS_BUNDLE_DIR)/*.clusterserviceversion.yaml
	sed -i -e 's,ICON_MEDIA_TYPE,image/svg+xml,g' $(MANIFESTS_BUNDLE_DIR)/*.clusterserviceversion.yaml

.PHONY: verify-operator-manifests
## verify operator manifests
verify-operator-manifests: prepare-operator-manifests
	$(Q)python3 -m venv $(PYTHON_VENV_DIR)
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install --upgrade setuptools
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install --upgrade pip
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install operator-courier==2.1.2
	$(Q)$(PYTHON_VENV_DIR)/bin/operator-courier --version
	$(Q)$(PYTHON_VENV_DIR)/bin/operator-courier $(VERBOSE_FLAG) verify $(MANIFESTS_DIR)
	rm -rf deploy/olm-catalog/$(GO_PACKAGE_REPO_NAME)/$(BUNDLE_VERSION)

.PHONY: build-bundle-image
build-bundle-image: verify-operator-manifests
	$(Q)operator-sdk bundle create --directory $(MANIFESTS_BUNDLE_DIR) -b $(CONTAINER_RUNTIME) --package $(CSV_PACKAGE_NAME) --channels beta,alpha --default-channel beta $(VERBOSE_FLAG) $(OPERATOR_BUNDLE_IMAGE_REF)

.PHONY: push-index-image
push-bundle-image: build-bundle-image registry-login
	$(Q)$(CONTAINER_RUNTIME) push $(OPERATOR_BUNDLE_IMAGE_REF)
	$(Q)operator-sdk bundle validate -b $(CONTAINER_RUNTIME) $(OPERATOR_BUNDLE_IMAGE_REF)

.PHONY: push-bundle-image
build-index-image: push-bundle-image
	$(Q)opm index add -u $(CONTAINER_RUNTIME) -p $(CONTAINER_RUNTIME) --bundles $(OPERATOR_BUNDLE_IMAGE_REF) --tag $(OPERATOR_INDEX_IMAGE_REF)

.PHONY: push-index-image
## push index image
push-index-image: build-index-image registry-login
	$(Q)$(CONTAINER_RUNTIME) push $(OPERATOR_INDEX_IMAGE_REF)

.PHONY: release-operator
## Build and release operator, bundle and index images
release-operator: push-operator-image push-bundle-image push-index-image

.PHONY: dev-release
## validating the operator by installing new quay releases
dev-release:
	BUNDLE_VERSION=$(BUNDLE_VERSION) ./hack/dev-release.sh

.PHONY: validate-release
## validate the operator by installing the releases
validate-release: setup-venv
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install yq==2.10.0
	BUNDLE_VERSION=$(BASE_BUNDLE_VERSION) CHANNEL="beta" ./hack/validate-release.sh

.PHONY: prepare-operatorhub-pr
## prepare files for OperatorHub PR
## use this target when the operator needs to be released as upstream operator
prepare-operatorhub-pr:
	./hack/prepare-operatorhub-pr.sh $(OPERATOR_VERSION) $(OPERATOR_BUNDLE_IMAGE_REF)