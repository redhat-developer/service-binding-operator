GOLANGCI_LINT_BIN=$(OUTPUT_DIR)/golangci-lint

.PHONY: lint
## Runs linters
lint: setup-venv lint-go-code lint-yaml lint-python-code lint-feature-files lint-conflicts

YAML_FILES := $(shell find . -path ./vendor -prune -o -path ./config -prune -o -path ./test/performance -prune -o -type f -regex ".*\.y[a]ml" -print)
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
	$(Q)curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./out v1.45.2

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