#!/bin/bash -x

HACK_DIR=${HACK_DIR:-$(dirname $0)}

# Remove SBR finalizers
$HACK_DIR/remove-sbr-finalizers.sh

# Delete deployed resources
kubectl delete -f deploy/crds/apps_v1alpha1_servicebindingrequest_cr.yaml
kubectl delete -f deploy/crds/apps.openshift.io_servicebindingrequests_crd.yaml
kubectl delete -f deploy/operator.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/service_account.yaml