#!/bin/bash -xe

if [[ -z $NAMESPACE ]]; then
    USE_NS="--all-namespaces"
else
    USE_NS="-n $NAMESPACE"
fi

# Remove SBR finalizers if CRD exists
[ -z "$(kubectl get -f deploy/crds/*crd.yaml -o jsonpath="{.metadata.name}" --ignore-not-found)" ] && exit 0

SBRS=($(kubectl get sbrs $USE_NS -o jsonpath="{.items[*].metadata.name}"))
SBR_NS=($(kubectl get sbrs $USE_NS -o jsonpath="{.items[*].metadata.namespace}"))

for i in "${!SBRS[@]}"; do
    kubectl patch sbr/${SBRS[$i]} -p '{"metadata":{"finalizers":[]}}' --type=merge --namespace=${SBR_NS[$i]};
done
