SHELL = /usr/bin/env bash -o pipefail
SHELLFLAGS = -ec

# -- Variables for acceptance tests
TEST_ACCEPTANCE_START_SBO ?= local
TEST_ACCEPTANCE_OUTPUT_DIR ?= $(OUTPUT_DIR)/acceptance-tests
TEST_ACCEPTANCE_REPORT_DIR ?= $(OUTPUT_DIR)/acceptance-tests-report
TEST_ACCEPTANCE_RESOURCES_DIR ?= $(TEST_ACCEPTANCE_OUTPUT_DIR)/resources
TEST_ACCEPTANCE_ARTIFACTS ?= $(ARTIFACT_DIR)
TEST_NAMESPACE = $(shell $(HACK_DIR)/get-test-namespace $(OUTPUT_DIR))
TEST_ACCEPTANCE_CLI ?= oc

TEST_ACCEPTANCE_TAGS ?=

ifdef TEST_ACCEPTANCE_TAGS
TEST_ACCEPTANCE_TAGS_ARG ?= --tags="~@disabled" --tags="~@examples" --tags="$(TEST_ACCEPTANCE_TAGS)"
else
TEST_ACCEPTANCE_TAGS_ARG ?= --tags="~@disabled" --tags="~@examples"
endif

OPERATOR_NAMESPACE ?=
OLM_NAMESPACE ?=

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

.PHONY: test-acceptance-setup
# Setup the environment for the acceptance tests
test-acceptance-setup: setup-venv
ifeq ($(TEST_ACCEPTANCE_START_SBO), local)
test-acceptance-setup: stop-local build test-cleanup create-test-namespace deploy-test-3rd-party-crds
	$(Q)echo "Starting local SBO instance"
	$(eval TEST_ACCEPTANCE_SBO_STARTED := $(shell ZAP_FLAGS="$(ZAP_FLAGS)" OUTPUT="$(TEST_ACCEPTANCE_OUTPUT_DIR)" RUN_IN_BACKGROUND=true ./hack/deploy-sbo-local.sh))
else ifeq ($(TEST_ACCEPTANCE_START_SBO), remote)
test-acceptance-setup: test-cleanup create-test-namespace
else ifeq ($(TEST_ACCEPTANCE_START_SBO), operator-hub)
test-acceptance-setup:
	$(eval TEST_ACCEPTANCE_SBO_STARTED := $(shell ./hack/deploy-sbo-operator-hub.sh))
endif
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install -q -r test/acceptance/features/requirements.txt
ifeq ($(TEST_ACCEPTANCE_CLI), oc)
	./test/acceptance/openshift-setup.sh
endif

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

.PHONY: test-acceptance-with-bundle
## Run acceptance tests with the operator installed from a given index image and channel
test-acceptance-with-bundle: deploy-from-index-image
	$(Q)TEST_ACCEPTANCE_START_SBO=remote $(MAKE) test-acceptance

.PHONY: test-acceptance-artifacts
# Collect artifacts from acceptance tests to be archived in CI
test-acceptance-artifacts: collect-kube-resources
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

.PHONY: collect-kube-resources
# Collect Kubernetes resources
collect-kube-resources:
	-$(Q)OUTPUT_PATH=$(TEST_ACCEPTANCE_RESOURCES_DIR) \
	OPERATOR_NAMESPACE=$(OPERATOR_NAMESPACE) \
	OLM_NAMESPACE=$(OLM_NAMESPACE) \
	$(HACK_DIR)/collect-kube-resources.sh