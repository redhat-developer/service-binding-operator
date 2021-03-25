PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
include hack/common.mk

# Current Operator version
VERSION ?= 0.8.0

GIT_COMMIT_ID ?= $(shell git rev-parse --short=8 HEAD)

OPERATOR_REGISTRY ?= quay.io
OPERATOR_REPO_REF ?= $(OPERATOR_REGISTRY)/redhat-developer/servicebinding-operator
OPERATOR_IMAGE_REF ?= $(OPERATOR_REPO_REF):$(GIT_COMMIT_ID)
OPERATOR_IMAGE_SHA_REF ?= $(shell $(CONTAINER_RUNTIME) inspect --format='{{index .RepoDigests 0}}' $(OPERATOR_IMAGE_REF) | cut -f 2 -d '@')
OPERATOR_BUNDLE_IMAGE_REF ?= $(OPERATOR_REPO_REF):bundle-$(VERSION)-$(GIT_COMMIT_ID)
OPERATOR_INDEX_IMAGE_REF ?= $(OPERATOR_REPO_REF):index

OPERATOR_CHANNELS ?= beta
DEFAULT_OPERATOR_CHANNEL ?= beta

CSV_PACKAGE_NAME ?= service-binding-operator

BUNDLE_METADATA_OPTS ?= --channels=$(OPERATOR_CHANNELS) --default-channel=$(DEFAULT_OPERATOR_CHANNEL)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

CGO_ENABLED ?= 0
GO111MODULE ?= on
GOCACHE ?= "$(shell echo ${PWD})/out/gocache"
GOFLAGS ?= -mod=vendor

ARTIFACT_DIR ?= $(shell echo ${PWD})/out
HACK_DIR ?= $(shell echo ${PWD})/hack
OUTPUT_DIR ?= $(shell echo ${PWD})/out
GOLANGCI_LINT_BIN=$(OUTPUT_DIR)/golangci-lint
PYTHON_VENV_DIR=$(OUTPUT_DIR)/venv3

CONTAINER_RUNTIME ?= docker

QUAY_USERNAME ?= redhat-developer+travis
REGISTRY_USERNAME ?= $(QUAY_USERNAME)
QUAY_TOKEN ?= ""
REGISTRY_PASSWORD ?= $(QUAY_TOKEN)

# -- Variables for acceptance tests
TEST_ACCEPTANCE_START_SBO ?= local
TEST_ACCEPTANCE_OUTPUT_DIR ?= $(OUTPUT_DIR)/acceptance-tests
TEST_ACCEPTANCE_REPORT_DIR ?= $(OUTPUT_DIR)/acceptance-tests-report
TEST_ACCEPTANCE_ARTIFACTS ?= $(ARTIFACT_DIR)
TEST_NAMESPACE = $(shell $(HACK_DIR)/get-test-namespace $(OUTPUT_DIR))

TEST_ACCEPTANCE_TAGS ?=

ifdef TEST_ACCEPTANCE_TAGS
TEST_ACCEPTANCE_TAGS_ARG ?= --tags="~@disabled" --tags="~@examples" --tags="$(TEST_ACCEPTANCE_TAGS)"
else
TEST_ACCEPTANCE_TAGS_ARG ?= --tags="~@disabled" --tags="~@examples"
endif

GO ?= CGO_ENABLED=$(CGO_ENABLED) GOCACHE=$(GOCACHE) GOFLAGS="$(GOFLAGS)" GO111MODULE=$(GO111MODULE) go


.DEFAULT_GOAL := help

.PHONY: lint
## Runs linters
lint: setup-venv lint-go-code lint-yaml lint-python-code lint-feature-files lint-conflicts

YAML_FILES := $(shell find . -path ./vendor -prune -o -path ./config -prune -o -type f -regex ".*\.y[a]ml" -print)
.PHONY: lint-yaml
## Runs yamllint on all yaml files
lint-yaml: ${YAML_FILES}
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install yamllint==1.23.0
	$(Q)$(PYTHON_VENV_DIR)/bin/yamllint -c .yamllint $(YAML_FILES)

.PHONY: lint-go-code
## Checks GO code
lint-go-code: $(GOLANGCI_LINT_BIN) fmt vet
	# This is required for OpenShift CI enviroment
	# Ref: https://github.com/openshift/release/pull/3438#issuecomment-482053250
	$(Q)GOFLAGS="$(GOFLAGS)" GOCACHE="$(GOCACHE)" $(OUTPUT_DIR)/golangci-lint ${V_FLAG} run --deadline=30m

$(GOLANGCI_LINT_BIN):
	$(Q)curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./out v1.18.0

.PHONY: lint-python-code
## Check the python code
lint-python-code: setup-venv
	$(Q)PYTHON_VENV_DIR=$(PYTHON_VENV_DIR) ./hack/check-python/lint-python-code.sh

## Check the acceptance tests feature files
.PHONY: lint-feature-files
lint-feature-files:
	$(Q)./hack/check-feature-files.sh

## Check for the presence of conflict notes in source file
.PHONY: lint-conflicts
lint-conflicts:
	$(Q)./hack/check-conflicts.sh

.PHONY: test
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
## Run unit and integration tests
test: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.0/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -covermode=atomic -coverprofile cover.out

.PHONY: build
## Build operator binary
build:
	$(GO) build -o bin/manager main.go

.PHONY: run
## Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet install
	$(GO) run ./main.go

.PHONY: install
## Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
## Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: deploy-cert-manager
deploy-cert-manager:
	kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.1.0/cert-manager.yaml
	kubectl rollout status -n cert-manager deploy/cert-manager-webhook -w --timeout=120s

.PHONY: deploy
## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize image deploy-cert-manager
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(OPERATOR_IMAGE_REF)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
## UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

.PHONY: manifests
## Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	$(GO) fmt ./...

# Run go vet against code
vet:
	$(GO) vet ./...

.PHONY: bundle
# Generate bundle manifests and metadata, then validate generated files.
bundle: manifests kustomize push-image
#	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(OPERATOR_REPO_REF)@$(OPERATOR_IMAGE_SHA_REF)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle --select-optional name=operatorhub

.PHONY: setup-venv
# Setup virtual environment
setup-venv:
	$(Q)python3 -m venv $(PYTHON_VENV_DIR)
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install --upgrade setuptools
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install --upgrade pip

# Testing setup
.PHONY: deploy-test-3rd-party-crds
deploy-test-3rd-party-crds:
	$(Q)kubectl --namespace $(TEST_NAMESPACE) apply -f ./test/third-party-crds/

.PHONY: create-test-namespace
create-test-namespace:
	$(Q)kubectl get namespace $(TEST_NAMESPACE) || kubectl create namespace $(TEST_NAMESPACE)

.PHONY: test-setup
test-setup: test-cleanup create-test-namespace deploy-test-3rd-party-crds

.PHONY: test-cleanup
test-cleanup:
	$(Q)-TEST_NAMESPACE=$(TEST_NAMESPACE) $(HACK_DIR)/test-cleanup.sh

.PHONY: deploy-rbac
deploy-rbac:
	@true

.PHONY: deploy-crds
deploy-crds: install
	@true

.PHONY: stop-local
## Stop Local: Stop locally running operator
stop-local:
	$(Q)-$(HACK_DIR)/remove-sbr-finalizers.sh
	$(Q)-$(HACK_DIR)/stop-sbo-local.sh

.PHONY: test-acceptance-setup
# Setup the environment for the acceptance tests
test-acceptance-setup: setup-venv
ifeq ($(TEST_ACCEPTANCE_START_SBO), local)
test-acceptance-setup: stop-local build test-cleanup create-test-namespace deploy-test-3rd-party-crds deploy-rbac deploy-crds
	$(Q)echo "Starting local SBO instance"
	$(eval TEST_ACCEPTANCE_SBO_STARTED := $(shell ZAP_FLAGS="$(ZAP_FLAGS)" OUTPUT="$(TEST_ACCEPTANCE_OUTPUT_DIR)" RUN_IN_BACKGROUND=true ./hack/deploy-sbo-local.sh))
else ifeq ($(TEST_ACCEPTANCE_START_SBO), remote)
test-acceptance-setup: test-cleanup create-test-namespace
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
		$(PYTHON_VENV_DIR)/bin/behave --junit --junit-directory $(TEST_ACCEPTANCE_OUTPUT_DIR) $(V_FLAG) --no-capture --no-capture-stderr $(TEST_ACCEPTANCE_TAGS_ARG) $(EXTRA_BEHAVE_ARGS) test/acceptance/features
ifeq ($(TEST_ACCEPTANCE_START_SBO), local)
	$(Q)kill $(TEST_ACCEPTANCE_SBO_STARTED)
endif

.PHONY: registry-login
registry-login:
	@$(CONTAINER_RUNTIME) login -u "$(REGISTRY_USERNAME)" --password-stdin $(OPERATOR_REGISTRY) <<<"$(REGISTRY_PASSWORD)"

.PHONY: image
## Build the image
image:
	$(Q)$(CONTAINER_RUNTIME) build -f Dockerfile -t $(OPERATOR_IMAGE_REF) .

.PHONY: push-image
# push operator image to registry
push-image: image registry-login
	$(CONTAINER_RUNTIME) push "$(OPERATOR_IMAGE_REF)"

.PHONY: bundle-image
# Build the bundle image
bundle-image: bundle
	$(CONTAINER_RUNTIME) build -f bundle.Dockerfile -t $(OPERATOR_BUNDLE_IMAGE_REF) .

.PHONY: push-bundle-image
push-bundle-image: bundle-image registry-login
	$(Q)$(CONTAINER_RUNTIME) push $(OPERATOR_BUNDLE_IMAGE_REF)
	$(Q)operator-sdk bundle validate --select-optional name=operatorhub -b $(CONTAINER_RUNTIME) $(OPERATOR_BUNDLE_IMAGE_REF)

.PHONY: index-image
index-image: push-bundle-image
	$(Q)opm index add -u $(CONTAINER_RUNTIME) -p $(CONTAINER_RUNTIME) --bundles $(OPERATOR_BUNDLE_IMAGE_REF) --tag $(OPERATOR_INDEX_IMAGE_REF)

.PHONY: push-index-image
# push index image
push-index-image: index-image registry-login
	$(Q)$(CONTAINER_RUNTIME) push $(OPERATOR_INDEX_IMAGE_REF)

.PHONY: release-operator
## Build and release operator, bundle and index images to registry
release-operator: push-image push-bundle-image push-index-image

.PHONY: prepare-operatorhub-pr
## prepare files for OperatorHub PR
## use this target when the operator needs to be released as upstream operator
prepare-operatorhub-pr:
	./hack/prepare-operatorhub-pr.sh $(VERSION) $(OPERATOR_BUNDLE_IMAGE_REF)

.PHONY: deploy-from-index-image
## deploy the operator from a given index image
deploy-from-index-image:
	$(info "Installing SBO using a Catalog Source from '$(OPERATOR_INDEX_IMAGE_REF)' index image")
	$(Q)OPERATOR_INDEX_IMAGE=$(OPERATOR_INDEX_IMAGE_REF) \
		OPERATOR_CHANNEL=$(DEFAULT_OPERATOR_CHANNEL) \
		OPERATOR_PACKAGE=$(CSV_PACKAGE_NAME) \
		SKIP_REGISTRY_LOGIN=true \
		./install.sh

.PHONY: test-acceptance-with-bundle
## Run acceptance tests with the operator installed from a given index image and channel
test-acceptance-with-bundle: deploy-from-index-image
	$(Q)TEST_ACCEPTANCE_START_SBO=remote $(MAKE) test-acceptance

.PHONY: test-acceptance-artifacts
# Collect artifacts from acceptance tests to be archived in CI
test-acceptance-artifacts:
	$(Q)echo "Gathering acceptance tests artifacts"
	$(Q)mkdir -p $(TEST_ACCEPTANCE_ARTIFACTS) \
		&& cp -rvf $(TEST_ACCEPTANCE_OUTPUT_DIR) $(TEST_ACCEPTANCE_ARTIFACTS)/

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

.PHONY: release-manifests
## prepare a manifest file for releasing operator on vanilla k8s cluster
release-manifests: REF=$(shell $(KUSTOMIZE) cfg grep "kind=ClusterServiceVersion" $(OUTPUT_DIR)/operatorhub-pr-files | $(YQ) e '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image' -)
release-manifests: prepare-operatorhub-pr kustomize yq
	git worktree add $(OUTPUT_DIR)/foo $(GIT_COMMIT_ID)
	cd $(OUTPUT_DIR)/foo/config/manager && $(KUSTOMIZE) edit set image controller=$(REF)
	$(KUSTOMIZE) build $(OUTPUT_DIR)/foo/config/default > $(OUTPUT_DIR)/release.yaml
	git worktree remove --force $(OUTPUT_DIR)/foo

.PHONY: clean
## Removes temp directories
clean:
	$(Q)-rm -rf ${V_FLAG} $(OUTPUT_DIR)
