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

#----------------------------------------------------------------
# HELP target
#----------------------------------------------------------------

# Based on https://gist.github.com/rcmachado/af3db315e31383502660
## Display this help text
help:/
	$(info Available targets)
	$(info -----------------)
	@awk '/^[a-zA-Z\-%\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		helpCommand = substr($$1, 0, index($$1, ":")-1); \
		if (helpMessage) { \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			gsub(/##/, "\n                                     ", helpMessage); \
			printf "%-35s - %s\n", helpCommand, helpMessage; \
			lastLine = "" \
		} \
	} \
	{ hasComment = match(lastLine, /^## (.*)/); \
		if(hasComment) { \
            lastLine=lastLine$$0; \
		} else { \
			lastLine = $$0 \
		} \
	}' $(MAKEFILE_LIST)

#-----------------------------------------------------------------------------
# Global Variables
#-----------------------------------------------------------------------------

# By default the project should be build under GOPATH/src/github.com/<orgname>/<reponame>
GO_PACKAGE_ORG_NAME ?= $(shell basename $$(dirname $$PWD))
GO_PACKAGE_REPO_NAME ?= $(shell basename $$PWD)
GO_PACKAGE_PATH ?= github.com/${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}

GIT_COMMIT_ID = $(shell git rev-parse --short HEAD)

OPERATOR_VERSION ?= 0.0.10
OPERATOR_GROUP ?= ${GO_PACKAGE_ORG_NAME}
OPERATOR_IMAGE ?= quay.io/${OPERATOR_GROUP}/${GO_PACKAGE_REPO_NAME}
OPERATOR_TAG_SHORT ?= $(OPERATOR_VERSION)
OPERATOR_TAG_LONG ?= $(OPERATOR_VERSION)-$(GIT_COMMIT_ID)
QUAY_TOKEN ?= ""

MANIFESTS_DIR ?= ./manifests
MANIFESTS_TMP ?= ./tmp/manifests

#---------------------------------------------------
# Lint targets
#---------------------------------------------------

GOLANGCI_LINT_BIN=./out/golangci-lint
.PHONY: lint
## Runs linters on Go code files and YAML files
lint: lint-go-code lint-yaml

YAML_FILES := $(shell find . -path ./vendor -prune -o -type f -regex ".*y[a]ml" -print)
.PHONY: lint-yaml
## runs yamllint on all yaml files
lint-yaml: ./vendor ${YAML_FILES}
	$(Q)yamllint -c .yamllint $(YAML_FILES)

.PHONY: lint-go-code
## Checks the code with golangci-lint
lint-go-code: ./vendor $(GOLANGCI_LINT_BIN)
	# This is required for OpenShift CI enviroment
	# Ref: https://github.com/openshift/release/pull/3438#issuecomment-482053250
	$(Q)GOCACHE=$(shell pwd)/out/gocache ./out/golangci-lint ${V_FLAG} run --deadline=30m

$(GOLANGCI_LINT_BIN):
	$(Q)curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./out v1.17.1


#------------------------------------------------------
# Test targets
#------------------------------------------------------

# Generate namespace name for test
./out/test-namespace:
	@echo -n "test-namespace-$(shell uuidgen | tr '[:upper:]' '[:lower:]')" > ./out/test-namespace

.PHONY: get-test-namespace
get-test-namespace: ./out/test-namespace
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
	$(Q)GO111MODULE=on GOCACHE=$(shell pwd)/out/gocache go test $(shell GOCACHE=$(shell pwd)/out/gocache go list ./...|grep -v e2e) -v -mod vendor

.PHONY: test-e2e-olm-ci
test-e2e-olm-ci:
	$(Q)sed -e "s,REPLACE_IMAGE,registry.svc.ci.openshift.org/${OPENSHIFT_BUILD_NAMESPACE}/stable:service-binding-operator-registry," ./test/e2e/catalog_source.yaml | kubectl apply -f -
	$(Q)kubectl apply -f ./test/e2e/subscription.yaml
	$(eval DEPLOYED_NAMESPACE := openshift-operators)
	$(Q)./hack/check-crds.sh
	$(Q)operator-sdk test local ./test/e2e --no-setup --go-test-flags "-v -timeout=15m"

#---------------------------------------------------------
# Build and vendor tarets
#---------------------------------------------------------

.PHONY: build 
## Build: compile the operator for Linux/AMD64.
build: ./out/operator

./out/operator:
	$(Q)CGO_ENABLED=0 GO111MODULE=on GOARCH=amd64 GOOS=linux go build ${V_FLAG} -o ./out/operator cmd/manager/main.go

## Build-Image: using operator-sdk to build a new image
build-image:
	$(Q)GO111MODULE=on operator-sdk build "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)"

## Vendor: "go mod vendor" resets the vendor folder to what's defined in go.mod
./vendor: go.mod go.sum
	$(Q)GOCACHE=$(shell pwd)/out/gocache GO111MODULE=on go mod vendor ${V_FLAG}

## Generate CSV: using oeprator-sdk generate cluster-service-version for current operator version
generate-csv:
	operator-sdk olm-catalog gen-csv --csv-version=$(OPERATOR_VERSION) --verbose

generate-olm:
	operator-courier --verbose flatten $(MANIFESTS_DIR) $(MANIFESTS_TMP)
	cp -vf deploy/crds/*_crd.yaml $(MANIFESTS_TMP)

#---------------------------------------------------------
# Deploy
#---------------------------------------------------------

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

## Push-Image: push docker image to upstream, including latest tag.
push-image: build-image
	docker tag "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)" "$(OPERATOR_IMAGE):latest"
	docker push "$(OPERATOR_IMAGE):$(OPERATOR_TAG_LONG)"
	docker push "$(OPERATOR_IMAGE):latest"
