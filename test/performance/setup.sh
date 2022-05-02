#!/bin/bash -e

# Environment
if [ -z "$QUAY_NAMESPACE" ]; then
    echo "QUAY_NAMESPACE environemnt variable needs to be set to a non-empty value"
    exit 1
fi

WS=${WS:-$(readlink -m $(dirname $0))}

# Ensure oc is loged in
OPENSHIFT_API_TOKEN=$(oc whoami -t) || (echo "a token is required to capture metrics, use 'oc login' to log into the cluster" && exit 1)

# Install Developer Sandbox
WSTC=$WS/toolchain-e2e.git
TOOLCHAIN_E2E_REPO=${TOOLCHAIN_E2E_REPO:-https://github.com/codeready-toolchain/toolchain-e2e}
TOOLCHAIN_E2E_BRANCH=${TOOLCHAIN_E2E_BRANCH:-master}
if [ ! -d $WSTC ]; then
    git clone $TOOLCHAIN_E2E_REPO $WSTC
fi
cd $WSTC
git reset --hard
git checkout $TOOLCHAIN_E2E_BRANCH
git pull
make dev-deploy-e2e

wait_for_deployment() {
    deployment=$1
    ns=$2
    #Wait for the operator to get up and running
    retries=50
    until [[ $retries == 0 ]]; do
        kubectl get deployment/$deployment -n $ns >/dev/null 2>&1 && break
        echo "Waiting for $deployment to be created in $ns namespace"
        sleep 5
        retries=$(($retries - 1))
    done
    kubectl rollout status -w deployment/$deployment -n $ns
}

wait_for_deployment host-operator-controller-manager toolchain-host-operator
wait_for_deployment registration-service toolchain-host-operator
wait_for_deployment member-operator-controller-manager toolchain-member-operator
wait_for_deployment member-operator-webhook toolchain-member-operator

# Install operators
$WS/install-operators.sh
