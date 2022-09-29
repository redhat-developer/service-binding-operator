#!/bin/bash -x

pushd $(dirname $0)
source ./env.sh

if [ -d ../../test/performance/toolchain-e2e.git ]; then
    pushd ../../test/performance/toolchain-e2e.git
        make clean-e2e-resources;
    popd;
fi

oc delete pod $(oc get pods -n openshift-operators -o json| jq -rc '.items[] | select(.metadata.name | startswith("service-binding-operator")).metadata.name') -n openshift-operators
popd
