#!/bin/bash -xe

HACK_DIR=${HACK_DIR:-$(dirname $0)}

# Remove SBR finalizers
$HACK_DIR/remove-sbr-finalizers.sh

# Delete deployed resources
RES_FILES=(
        crds/apps.openshift.io_servicebindingrequests_crd.yaml
        operator.yaml
        role_binding.yaml
        role.yaml
        service_account.yaml
)

for rf in ${RES_FILES[@]} ; do
    kubectl delete -f "deploy/$rf" --ignore-not-found
done
