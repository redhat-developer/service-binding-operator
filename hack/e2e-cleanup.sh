#!/bin/bash -xe

HACK_DIR=${HACK_DIR:-$(dirname $0)}

# Remove SBR finalizers
NAMESPACE=$TEST_NAMESPACE $HACK_DIR/remove-sbr-finalizers.sh
