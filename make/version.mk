SHELL = /usr/bin/env bash -o pipefail
SHELLFLAGS = -ec

# Current Operator version
VERSION ?= 1.3.1

GIT_COMMIT_ID ?= $(shell git rev-parse --short=8 HEAD)

OPERATOR_REGISTRY ?= quay.io
OPERATOR_REPO_REF ?= $(OPERATOR_REGISTRY)/redhat-developer/servicebinding-operator
OPERATOR_IMAGE_REF ?= $(OPERATOR_REPO_REF):$(GIT_COMMIT_ID)
OPERATOR_IMAGE_SHA_REF ?= $(shell $(CONTAINER_RUNTIME) inspect --format='{{index .RepoDigests 0}}' $(OPERATOR_IMAGE_REF) | cut -f 2 -d '@')
OPERATOR_BUNDLE_IMAGE_REF ?= $(OPERATOR_REPO_REF):bundle-$(VERSION)-$(GIT_COMMIT_ID)
OPERATOR_INDEX_IMAGE_REF ?= $(OPERATOR_REPO_REF):index

.PHONY: operator-repo-ref
# Prints operator repo ref
operator-repo-ref:
	@echo $(OPERATOR_REPO_REF)

.PHONY: operator-image-ref
# Prints operator image ref
operator-image-ref:
	@echo $(OPERATOR_IMAGE_REF)

.PHONY: operator-bundle-image-ref
# Prints operator bundle image ref
operator-bundle-image-ref:
	@echo $(OPERATOR_BUNDLE_IMAGE_REF)

.PHONY: operator-index-image-ref
# Prints operator index image ref
operator-index-image-ref:
	@echo $(OPERATOR_INDEX_IMAGE_REF)
