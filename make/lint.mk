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

GO_LINT_CMD = $(Q)GOFLAGS="$(GOFLAGS)" GOGC=30 GOCACHE=$(GOCACHE) $(OUTPUT_DIR)/golangci-lint ${V_FLAG} run --concurrency=1 --verbose --deadline=30m --disable-all --enable

.PHONY: lint-go-code
## Checks GO code
lint-go-code: $(GOLANGCI_LINT_BIN) fmt vet
	# This is required for OpenShift CI enviroment
	# Ref: https://github.com/openshift/release/pull/3438#issuecomment-482053250
	$(GO_LINT_CMD) deadcode
	$(GO_LINT_CMD) gosimple
	$(GO_LINT_CMD) staticcheck
	$(GO_LINT_CMD) errcheck
	$(GO_LINT_CMD) govet
	$(GO_LINT_CMD) ineffassign
	$(GO_LINT_CMD) structcheck
	$(GO_LINT_CMD) typecheck
	$(GO_LINT_CMD) unused
	$(GO_LINT_CMD) varcheck

$(GOLANGCI_LINT_BIN):
	$(Q)curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./out v1.51.2

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
