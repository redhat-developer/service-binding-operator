#!/bin/bash -xe

[ ! -z "$TEST_NAMESPACE" ] && [ -z "$(kubectl get ns "$TEST_NAMESPACE" -o jsonpath="{.metadata.name}" --ignore-not-found)" ] && exit 0

HACK_DIR=${HACK_DIR:-$(dirname $0)}

# Remove SBR finalizers
NAMESPACE=$TEST_NAMESPACE $HACK_DIR/remove-sbr-finalizers.sh

[ ! -z "$TEST_NAMESPACE" ] && kubectl delete namespace ${TEST_NAMESPACE} --ignore-not-found --timeout=45s --wait
