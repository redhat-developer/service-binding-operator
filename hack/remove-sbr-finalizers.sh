#!/bin/bash -xe

if [[ -z $NAMESPACE ]]; then
    USE_NS="--all-namespaces"
else
    USE_NS="-n $NAMESPACE"
fi

for file in $(find config/crd/bases -name *_servicebindings.yaml); do
  # Remove SBR finalizers if CRD exists
  CRD_NAME=$(kubectl get -f $file -o jsonpath="{.metadata.name}" --ignore-not-found)
  [ -z $CRD_NAME ] && continue

  SBRS=($(kubectl get $CRD_NAME $USE_NS -o jsonpath="{.items[*].metadata.name}"))
  SBR_NS=($(kubectl get $CRD_NAME $USE_NS -o jsonpath="{.items[*].metadata.namespace}"))

  for i in "${!SBRS[@]}"; do
      kubectl patch $CRD_NAME/${SBRS[$i]} -p '{"metadata":{"finalizers":[]}}' --type=merge --namespace=${SBR_NS[$i]};
  done

done