SHELL = /usr/bin/env bash -o pipefail
SHELLFLAGS = -ec

# Current Operator version
VERSION ?= 1.1.0

GIT_COMMIT_ID ?= $(shell git rev-parse --short=8 HEAD)

OPERATOR_REGISTRY ?= quay.io
OPERATOR_REPO_REF ?= $(OPERATOR_REGISTRY)/redhat-developer/servicebinding-operator
OPERATOR_IMAGE_REF ?= $(OPERATOR_REPO_REF):$(GIT_COMMIT_ID)
OPERATOR_IMAGE_SHA_REF ?= $(shell $(CONTAINER_RUNTIME) inspect --format='{{index .RepoDigests 0}}' $(OPERATOR_IMAGE_REF) | cut -f 2 -d '@')
OPERATOR_BUNDLE_IMAGE_REF ?= $(OPERATOR_REPO_REF):bundle-$(VERSION)-$(GIT_COMMIT_ID)
OPERATOR_INDEX_IMAGE_REF ?= $(OPERATOR_REPO_REF):index
