SHELL = /usr/bin/env bash -o pipefail
SHELLFLAGS = -ec

.PHONY: release-operator
## Build and release operator, bundle and index images to registry
release-operator: push-image push-bundle-image push-index-image

.PHONY: prepare-operatorhub-pr
## prepare files for OperatorHub PR
## use this target when the operator needs to be released as upstream operator
prepare-operatorhub-pr: yq
	PATH=$(PWD)/bin:$(PATH) ./hack/prepare-operatorhub-pr.sh $(VERSION) $(OPERATOR_BUNDLE_IMAGE_REF)

.PHONY: release-manifests
## prepare a manifest file for releasing operator on vanilla k8s cluster
release-manifests: REF=$(shell $(KUSTOMIZE) cfg grep "kind=ClusterServiceVersion" $(OUTPUT_DIR)/operatorhub-pr-files | $(YQ) e '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image' -)
release-manifests: prepare-operatorhub-pr kustomize yq
	git worktree add $(OUTPUT_DIR)/foo $(GIT_COMMIT_ID)
	cd $(OUTPUT_DIR)/foo/config/manager && $(KUSTOMIZE) edit set image controller=$(REF)
	$(KUSTOMIZE) build $(OUTPUT_DIR)/foo/config/default > $(OUTPUT_DIR)/release.yaml
	git worktree remove --force $(OUTPUT_DIR)/foo

.PHONY: branch-release
## prepare release branch
branch-release:
	OUTPUT_DIR=$(OUTPUT_DIR) VERSION=$(VERSION) $(HACK_DIR)/branch-release.sh
