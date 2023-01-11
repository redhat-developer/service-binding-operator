#!/bin/bash -x

OUTPUT_DIR=${OUTPUT_DIR:-out}
mkdir -p $OUTPUT_DIR

export TEST_ACCEPTANCE_START_SBO=scenarios
export TEST_OPERATOR_INDEX_IMAGE=${OPERATOR_INDEX_IMAGE_REF:-quay.io/redhat-developer/servicebinding-operator:index}
export TEST_OPERATOR_CHANNEL=candidate
operator_index_yaml=$OUTPUT_DIR/operator-index.yaml

opm render ${TEST_OPERATOR_INDEX_IMAGE} -o yaml > $operator_index_yaml
yq_exp='select(.schema=="olm.channel") | select(.name=="'${TEST_OPERATOR_CHANNEL}'").entries[] | select(.replaces == null).name'
export TEST_OPERATOR_CSV=$(yq eval "$yq_exp" "$operator_index_yaml")
yq_exp='select(.schema=="olm.channel") | select(.name=="'${TEST_OPERATOR_CHANNEL}'").package'
export TEST_OPERATOR_PACKAGE=$(yq eval "$yq_exp" "$operator_index_yaml")

env | grep TEST_OPERATOR