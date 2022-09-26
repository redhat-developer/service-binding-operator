#!/bin/bash -e
export RHOAS_CHANNEL=${RHOAS_CHANNEL:-beta}
export RHOAS_PACKAGE=${RHOAS_PACKAGE:-rhoas-operator}
export RHOAS_CATSRC_NAMESPACE=${RHOAS_CATSRC_NAMESPACE:-openshift-marketplace}
export RHOAS_CATSRC_NAME=${RHOAS_CATSRC_NAME:-community-operators}
export RHOAS_NAMESPACE=${RHOAS_NAMESPACE:-openshift-operators}

echo "Installing RHOAS Operator"
oc apply -f - <<EOD
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: $RHOAS_PACKAGE
  namespace: $RHOAS_NAMESPACE
spec:
  channel: $RHOAS_CHANNEL
  installPlanApproval: Automatic
  name: $RHOAS_PACKAGE
  source: $RHOAS_CATSRC_NAME
  sourceNamespace: $RHOAS_CATSRC_NAMESPACE
EOD

#Wait for the operator to get up and running
retries=50
until [[ $retries == 0 ]]; do
    kubectl get deployment/rhoas-operator -n $RHOAS_NAMESPACE >/dev/null 2>&1 && break
    echo "Waiting for rhoas-operator to be created in $RHOAS_NAMESPACE namespace"
    sleep 5
    retries=$(($retries - 1))
done
kubectl rollout status -w deployment/rhoas-operator -n $RHOAS_NAMESPACE
