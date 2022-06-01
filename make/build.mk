SHELL = /usr/bin/env bash -o pipefail
SHELLFLAGS = -ec

OPERATOR_CHANNELS ?= beta,candidate
DEFAULT_OPERATOR_CHANNEL ?= candidate
CSV_PACKAGE_NAME ?= service-binding-operator

BUNDLE_METADATA_OPTS ?= --channels=$(OPERATOR_CHANNELS) --default-channel=$(DEFAULT_OPERATOR_CHANNEL)

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
	$(OPM) index add -u $(CONTAINER_RUNTIME) -p $(CONTAINER_RUNTIME) --bundles $(OPERATOR_BUNDLE_IMAGE_REF) --tag $(OPERATOR_INDEX_IMAGE_REF)

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
