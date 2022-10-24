SHELL = /usr/bin/env bash -o pipefail
SHELLFLAGS = -ec

OS = $(shell go env GOOS)
ARCH = $(shell go env GOARCH)

CGO_ENABLED ?= 0
GO111MODULE ?= on
GOCACHE ?= "$(PROJECT_DIR)/out/gocache"
GOFLAGS ?= -mod=vendor

ARTIFACT_DIR ?= $(PROJECT_DIR)/out
HACK_DIR ?= $(PROJECT_DIR)/hack
OUTPUT_DIR ?= $(PROJECT_DIR)/out
PYTHON_VENV_DIR = $(OUTPUT_DIR)/venv3

CONTAINER_RUNTIME ?= docker

QUAY_USERNAME ?= redhat-developer+travis
REGISTRY_USERNAME ?= $(QUAY_USERNAME)
REGISTRY_NAMESPACE ?= $(QUAY_USERNAME)
QUAY_TOKEN ?= ""
REGISTRY_PASSWORD ?= $(QUAY_TOKEN)

GO ?= CGO_ENABLED=$(CGO_ENABLED) GOCACHE=$(GOCACHE) GOFLAGS="$(GOFLAGS)" GO111MODULE=$(GO111MODULE) go

.DEFAULT_GOAL := help

## Print help message for all Makefile targets
## Run `make` or `make help` to see the help
.PHONY: help
help: ## Credit: https://gist.github.com/prwhite/8168133#gistcomment-2749866

	@printf "Usage:\n  make <target>\n\n";

	@awk '{ \
			if ($$0 ~ /^.PHONY: [a-zA-Z\-_0-9]+$$/) { \
				helpCommand = substr($$0, index($$0, ":") + 2); \
				if (helpMessage) { \
					printf "\033[36m%-20s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^[a-zA-Z\-_0-9.]+:/) { \
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
ZAP_ENCODER_FLAG = --zap-log-level=debug --zap-encoder=console
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
endif
ifeq ($(VERBOSE),3)
	Q_FLAG =
	QUIET_FLAG =
	S_FLAG =
	V_FLAG = -v
	VERBOSE_FLAG = --verbose
	X_FLAG = -x
endif

.PHONY: setup-venv
# Setup virtual environment
setup-venv:
	$(Q)python3 -m venv $(PYTHON_VENV_DIR)
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install --upgrade setuptools
	$(Q)$(PYTHON_VENV_DIR)/bin/pip install --upgrade pip

.PHONY: clean
## Removes temp directories
clean:
	$(Q)-rm -rf ${V_FLAG} $(OUTPUT_DIR)

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

gen-mocks: mockgen
	PATH=$(shell pwd)/bin:$(shell printenv PATH) $(GO) generate $(V_FLAG) ./...

# go-install-tool will 'go install' any package $2 and install it to $1.
define go-install-tool
[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp 2>/dev/null ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2)@v$(3) ;\
rm -rf $$TMP_DIR ;\
echo "$$(basename $(1))@v$(3) installed" ;\
}
endef

define output-install
echo "$(1)@v$(2) installed"
endef

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
CONTROLLER_GEN_VERSION ?= 0.7.0
CONTROLLER_GEN_LOCAL_VERSION := $(shell [ -f $(CONTROLLER_GEN) ] && $(CONTROLLER_GEN) --version | cut -d' ' -f 2)
CONTROLLER_GEN_HOST_VERSION := $(shell command -v controller-gen >/dev/null && controller-gen --version | cut -d' ' -f 2)
controller-gen:
	$(Q)mkdir -p $(dir $(CONTROLLER_GEN))
ifneq (v$(CONTROLLER_GEN_VERSION), $(CONTROLLER_GEN_LOCAL_VERSION))
	$(Q)rm -f $(CONTROLLER_GEN)
ifeq (v$(CONTROLLER_GEN_VERSION),$(CONTROLLER_GEN_HOST_VERSION))
	$(Q)ln -s $$(command -v controller-gen) $(CONTROLLER_GEN)
	@echo "controller-gen found at $$(command -v controller-gen)"
else
	$(Q)$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_GEN_VERSION))
endif
endif


YQ = $(shell pwd)/bin/yq
YQ_VERSION ?= 4.26.1
YQ_LOCAL_VERSION := $(shell [ -f $(YQ) ] && $(YQ) --version | cut -d' ' -f 3)
YQ_HOST_VERSION := $(shell command -v yq >/dev/null && yq --version | cut -d' ' -f 3)
yq:
	$(Q)mkdir -p $(dir $(YQ))
ifneq ($(YQ_VERSION), $(YQ_LOCAL_VERSION))
	$(Q)rm -f $(YQ)
ifeq ($(YQ_VERSION),$(YQ_HOST_VERSION))
	$(Q)ln -s $$(command -v yq) $(YQ)
	@echo "yq found at $$(command -v yq)"
else
	$(Q)$(call go-install-tool,$(YQ),github.com/mikefarah/yq/v4,$(YQ_VERSION))
endif
endif

KUBECTL_SLICE = $(shell pwd)/bin/kubectl-slice
KUBECTL_SLICE_VERSION ?= 1.1.0
KUBECTL_SLICE_LOCAL_VERSION := $(shell [ -f $(KUBECTL_SLICE) ] && $(KUBECTL_SLICE) --version | cut -d' ' -f 3)
KUBECTL_SLICE_HOST_VERSION := $(shell command -v kubectl-slice >/dev/null && kubectl-slice --version | cut -d' ' -f 3)
kubectl-slice:
	$(Q)mkdir -p $(dir $(KUBECTL_SLICE))
ifneq ($(KUBECTL_SLICE_VERSION), $(KUBECTL_SLICE_LOCAL_VERSION))
	$(Q)rm -f $(KUBECTL_SLICE)
ifeq ($(KUBECTL_SLICE_VERSION),$(KUBECTL_SLICE_HOST_VERSION))
	$(Q)ln -s $$(command -v kubectl-slice) $(KUBECTL_SLICE)
	@echo "kubectl-slice found at $$(command -v kubectl-slice)"
else
	$(Q){ \
		rm -f $(KUBECTL_SLICE) ;\
		arch=$$(case "$(ARCH)" in "amd64") echo "x86_64" ;; *) echo "$(ARCH)" ;; esac) ;\
		mkdir -p $(KUBECTL_SLICE)-install ;\
		curl -sSLo $(KUBECTL_SLICE)-install/kubectl-slice.tar.gz https://github.com/patrickdappollonio/kubectl-slice/releases/download/v$(KUBECTL_SLICE_VERSION)/kubectl-slice_$(KUBECTL_SLICE_VERSION)_$(OS)_$${arch}.tar.gz ;\
		tar xvfz $(KUBECTL_SLICE)-install/kubectl-slice.tar.gz -C $(KUBECTL_SLICE)-install/ > /dev/null ;\
		mv $(KUBECTL_SLICE)-install/kubectl-slice $(KUBECTL_SLICE) ;\
		rm -rf $(KUBECTL_SLICE)-install ;\
		$(call output-install,kubectl-slice,$(KUBECTL_SLICE_VERSION)) ;\
	}
endif
endif

MOCKGEN = $(shell pwd)/bin/mockgen
MOCKGEN_VERSION ?= 1.6.0
MOCKGEN_LOCAL_VERSION := $(shell [ -f $(MOCKGEN) ] && $(MOCKGEN) --version | cut -d' ' -f 3)
MOCKGEN_HOST_VERSION := $(shell command -v mockgen >/dev/null && mockgen --version | cut -d' ' -f 3)
mockgen:
	$(Q)mkdir -p $(dir $(MOCKGEN))
ifneq (v$(MOCKGEN_VERSION), $(MOCKGEN_LOCAL_VERSION))
	$(Q)rm -f $(MOCKGEN)
ifeq (v$(MOCKGEN_VERSION),$(MOCKGEN_HOST_VERSION))
	$(Q)ln -s $$(command -v mockgen) $(MOCKGEN)
	@echo "mockgen found at $$(command -v mockgen)"
else
	$(Q)$(call go-install-tool,$(MOCKGEN),github.com/golang/mock/mockgen,$(MOCKGEN_VERSION))
endif
endif


KUSTOMIZE = $(shell pwd)/bin/kustomize
KUSTOMIZE_VERSION ?= 4.5.4
KUSTOMIZE_LOCAL_VERSION := $(shell [ -f $(KUSTOMIZE) ] && $(KUSTOMIZE) version | cut -d' ' -f 1 | cut -d'/' -f 2)
KUSTOMIZE_HOST_VERSION := $(shell command -v kustomize >/dev/null && kustomize version | cut -d' ' -f 1 | cut -d'/' -f 2)
kustomize:
	$(Q)mkdir -p $(dir $(KUSTOMIZE))
ifneq (v$(KUSTOMIZE_VERSION), $(KUSTOMIZE_LOCAL_VERSION))
	$(Q)rm -f $(KUSTOMIZE)
ifeq (v$(KUSTOMIZE_VERSION),$(KUSTOMIZE_HOST_VERSION))
	$(Q)ln -s $$(command -v kustomize) $(KUSTOMIZE)
	@echo "kustomize found at $$(command -v kustomize)"
else
	$(Q){ \
		set -e ;\
		mkdir -p $(dir $(KUSTOMIZE)) ;\
		rm -f $(KUSTOMIZE) ; \
		mkdir -p $(KUSTOMIZE)-install ;\
		curl -sSLo $(KUSTOMIZE)-install/kustomize.tar.gz https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv$(KUSTOMIZE_VERSION)/kustomize_v$(KUSTOMIZE_VERSION)_$(OS)_$(ARCH).tar.gz ;\
		tar xzvf $(KUSTOMIZE)-install/kustomize.tar.gz -C $(KUSTOMIZE)-install/ >/dev/null ;\
		mv $(KUSTOMIZE)-install/kustomize $(KUSTOMIZE) ;\
		rm -rf $(KUSTOMIZE)-install ;\
		chmod +x $(KUSTOMIZE) ;\
		$(call output-install,kustomize,$(KUSTOMIZE_VERSION)) ;\
	}
endif
endif

.PHONY: opm
OPM ?=  $(shell pwd)/bin/opm
OPM_VERSION ?= 1.22.0
OPM_LOCAL_VERSION := $(shell [ -f $(OPM) ] && $(OPM) version | cut -d'"' -f 2)
OPM_HOST_VERSION := $(shell command -v opm >/dev/null && opm version | cut -d'"' -f 2)
opm:
	$(Q)mkdir -p $(dir $(OPM))
ifneq (v$(OPM_VERSION), $(OPM_LOCAL_VERSION))
	$(Q)rm -f $(OPM)
ifeq (v$(OPM_VERSION),$(OPM_HOST_VERSION))
	$(Q)ln -s $$(command -v opm) $(OPM)
	@echo "opm found at $$(command -v opm)"
else
	$(Q){ \
		set -e ;\
		mkdir -p $(dir $(OPM)) ;\
		rm -f $(OPM) ; \
		curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v$(OPM_VERSION)/$(OS)-$(ARCH)-opm ;\
		chmod +x $(OPM) ;\
		$(call output-install,opm,$(OPM_VERSION)) ;\
	}
endif
endif

.PHONY: minikube
MINIKUBE ?=  $(shell pwd)/bin/minikube
MINIKUBE_VERSION ?= 1.26.1
MINIKUBE_LOCAL_VERSION = $(shell [ -f $(MINIKUBE) ] && $(MINIKUBE) version --short)
MINIKUBE_HOST_VERSION = $(shell command -v minikube >/dev/null && minikube version --short)
minikube:
	$(Q)mkdir -p $(dir $(MINIKUBE))
ifneq (v$(MINIKUBE_VERSION), $(MINIKUBE_LOCAL_VERSION))
	$(Q)rm -f $(MINIKUBE)
ifeq (v$(MINIKUBE_VERSION),$(MINIKUBE_HOST_VERSION))
	$(Q)ln -s $$(command -v minikube) $(MINIKUBE)
	@echo "minikube found at $$(command -v minikube)"
else
	$(Q){ \
		set -e ;\
		mkdir -p $(dir $(MINIKUBE)) ;\
		rm -f $(MINIKUBE) ; \
		curl -sSLo $(MINIKUBE)  https://storage.googleapis.com/minikube/releases/v$(MINIKUBE_VERSION)/minikube-$(OS)-$(ARCH) ;\
		chmod +x $(MINIKUBE) ;\
		$(call output-install,minikube,$(MINIKUBE_VERSION)) ;\
	}
endif
endif

.PHONY: operator-sdk
OPERATOR_SDK ?=  $(shell pwd)/bin/operator-sdk
OPERATOR_SDK_VERSION ?= 1.24.0
OPERATOR_SDK_LOCAL_VERSION := $(shell [ -f $(OPERATOR_SDK) ] && $(OPERATOR_SDK) version | cut -d'"' -f 2)
OPERATOR_SDK_HOST_VERSION := $(shell command -v operator-sdk >/dev/null && operator-sdk version | cut -d'"' -f 2)
operator-sdk:
	$(Q)mkdir -p $(dir $(OPERATOR_SDK))
ifneq (v$(OPERATOR_SDK_VERSION), $(OPERATOR_SDK_LOCAL_VERSION))
	$(Q)rm -f $(OPERATOR_SDK)
ifeq (v$(OPERATOR_SDK_VERSION),$(OPERATOR_SDK_HOST_VERSION))
	$(Q)ln -s $$(command -v operator-sdk) $(OPERATOR_SDK)
	@echo "operator-sdk found at $$(command -v operator-sdk)"
else
	$(Q){ \
		set -e ;\
		mkdir -p $(dir $(OPERATOR_SDK)) ;\
		rm -f $(OPERATOR_SDK) ; \
		curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/v$(OPERATOR_SDK_VERSION)/operator-sdk_$(OS)_$(ARCH) ;\
		chmod +x $(OPERATOR_SDK) ;\
		$(call output-install,operator-sdk,$(OPERATOR_SDK_VERSION)) ;\
	}
endif
endif

.PHONY: kubectl
KUBECTL ?=  $(shell pwd)/bin/kubectl
KUBECTL_VERSION ?= 1.25.3
KUBECTL_LOCAL_VERSION := $(shell [ -f $(KUBECTL) ] && $(KUBECTL) version --client --output yaml | grep 'gitVersion:' | tr -d ' ' | cut -d':' -f 2)
KUBECTL_HOST_VERSION := $(shell command -v kubectl >/dev/null && kubectl version --client --output yaml | grep 'gitVersion:' | tr -d ' ' | cut -d':' -f 2)
kubectl:
	$(Q)mkdir -p $(dir $(KUBECTL))
ifneq ($(KUBECTL_HOST_VERSION),)
	$(Q)rm -f $(KUBECTL)
	$(Q)ln -s $$(command -v kubectl) $(KUBECTL)
	@echo "kubectl found at $$(command -v kubectl)"
else ifneq (v$(KUBECTL_VERSION), $(KUBECTL_LOCAL_VERSION))
	$(Q){ \
		set -e ;\
		mkdir -p $(dir $(KUBECTL)) ;\
		rm -f $(KUBECTL) ; \
		curl -sSLo $(KUBECTL) https://dl.k8s.io/release/v$(KUBECTL_VERSION)/bin/$(OS)/$(ARCH)/kubectl ;\
		chmod +x $(KUBECTL) ;\
		$(call output-install,kubectl,$(KUBECTL_VERSION)) ;\
	}
endif

.PHONY: helm
HELM ?=  $(shell pwd)/bin/helm
HELM_VERSION ?= 3.10.1
HELM_LOCAL_VERSION := $(shell [ -f $(HELM) ] && $(HELM) version --short | cut -d'+' -f 1)
HELM_HOST_VERSION := $(shell command -v helm >/dev/null && helm version --short | cut -d'+' -f 1)
helm:
	$(Q)mkdir -p $(dir $(HELM))
ifneq (v$(HELM_VERSION), $(HELM_LOCAL_VERSION))
	$(Q)rm -f $(HELM)
ifeq (v$(HELM_VERSION),$(HELM_HOST_VERSION))
	$(Q)ln -s $$(command -v helm) $(HELM)
	@echo "helm found at $$(command -v helm)"
else
	$(Q){ \
		set -e ;\
		mkdir -p $(dir $(HELM)) $(HELM)-install ;\
		curl -sSLo $(HELM)-install/helm.tar.gz https://get.helm.sh/helm-v$(HELM_VERSION)-$(OS)-$(ARCH).tar.gz ;\
		tar xvfz $(HELM)-install/helm.tar.gz -C $(HELM)-install >/dev/null ;\
		rm -f $(HELM) ; \
		cp $(HELM)-install/$(OS)-$(ARCH)/helm $(HELM) ;\
		rm -r $(HELM)-install ;\
		chmod +x $(HELM) ;\
		$(call output-install,helm,$(HELM_VERSION)) ;\
	}
endif
endif

.PHONY: install-tools
install-tools: controller-gen helm kubectl kubectl-slice kustomize minikube mockgen operator-sdk opm yq
	@echo
	@echo run '`eval $$(make local-env)`' to configure your shell to use tools in the ./bin folder

.PHONY: local-env
local-env:
	@echo export PATH=$(shell pwd)/bin:$$PATH

all: build
