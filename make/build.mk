SHELL = /usr/bin/env bash -o pipefail
SHELLFLAGS = -ec

OPERATOR_CHANNELS ?= beta,candidate
DEFAULT_OPERATOR_CHANNEL ?= candidate
CSV_PACKAGE_NAME ?= service-binding-operator

BUNDLE_METADATA_OPTS ?= --channels=$(OPERATOR_CHANNELS) --default-channel=$(DEFAULT_OPERATOR_CHANNEL)

OPERATOR_INDEX_NAME ?= $(CSV_PACKAGE_NAME)-index
OPERATOR_INDEX_DIR ?= $(OPERATOR_INDEX_NAME)
OPERATOR_INDEX_YAML ?= $(OPERATOR_INDEX_DIR)/index.yaml

.PHONY: build
## Build operator binary
build:
	$(GO) build -o bin/manager main.go

.PHONY: manifests
## Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: bundle
# Generate bundle manifests and metadata, then validate generated files.
bundle: manifests kustomize yq kubectl-slice push-image
#	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(OPERATOR_REPO_REF)@$(OPERATOR_IMAGE_SHA_REF)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(YQ) e -i '.metadata.annotations.containerImage="$(OPERATOR_REPO_REF)@$(OPERATOR_IMAGE_SHA_REF)"' bundle/manifests/service-binding-operator.clusterserviceversion.yaml
	# this is needed because operator-sdk 1.16 filters out aggregated cluster role and the accompanied binding
	$(KUSTOMIZE) build config/manifests | $(YQ) e 'select((.kind == "ClusterRole" and .metadata.name == "service-binding-controller-role") or (.kind == "ClusterRoleBinding" and .metadata.name == "service-binding-controller-rolebinding"))' - | $(KUBECTL_SLICE) -o bundle/manifests -t '{{.metadata.name}}_{{.apiVersion | replace "/" "_"}}_{{.kind | lower}}.yaml'
	operator-sdk bundle validate ./bundle --select-optional name=operatorhub

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
index-image: opm push-bundle-image
	mkdir -p $(OPERATOR_INDEX_DIR)
	-$(OPM) generate dockerfile $(OPERATOR_INDEX_NAME)
	$(OPM) init $(CSV_PACKAGE_NAME) --default-channel=$(DEFAULT_OPERATOR_CHANNEL) --icon=$(PROJECT_DIR)/assets/icon/sbo-logo.svg --output=yaml > $(OPERATOR_INDEX_YAML)
	$(OPM) render $(OPERATOR_BUNDLE_IMAGE_REF) --output=yaml >> $(OPERATOR_INDEX_YAML)
	@echo "---" >> $(OPERATOR_INDEX_YAML)
	@echo "schema: olm.channel" >> $(OPERATOR_INDEX_YAML)
	@echo "package: $(CSV_PACKAGE_NAME)" >> $(OPERATOR_INDEX_YAML)
	@echo "name: $(DEFAULT_OPERATOR_CHANNEL)" >> $(OPERATOR_INDEX_YAML)
	@echo "entries:" >> $(OPERATOR_INDEX_YAML)
	@echo "- name: $(CSV_PACKAGE_NAME).v$(VERSION)" >> $(OPERATOR_INDEX_YAML)
	$(OPM) validate $(OPERATOR_INDEX_NAME)
	$(CONTAINER_RUNTIME) build -f $(OPERATOR_INDEX_NAME).Dockerfile -t $(OPERATOR_INDEX_IMAGE_REF) .

.PHONY: push-index-image
# push index image
push-index-image: index-image registry-login
	$(Q)$(CONTAINER_RUNTIME) push $(OPERATOR_INDEX_IMAGE_REF)

.PHONY: deploy-from-index-image
## deploy the operator from a given index image
deploy-from-index-image:
	$(info "Installing SBO using a Catalog Source from '$(OPERATOR_INDEX_IMAGE_REF)' index image")
	$(Q)OPERATOR_INDEX_IMAGE=$(OPERATOR_INDEX_IMAGE_REF) \
		OPERATOR_CHANNEL=$(DEFAULT_OPERATOR_CHANNEL) \
		OPERATOR_PACKAGE=$(CSV_PACKAGE_NAME) \
		SKIP_REGISTRY_LOGIN=true \
		./install.sh
