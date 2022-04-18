#!/usr/bin/env bash

set -x

CLI="kubectl -n ${TEST_NAMESPACE}"
${CLI} apply -f ${WORKSPACE}/secret.yaml --wait 
${CLI} apply -f ${WORKSPACE}/application.yaml --wait

${CLI} rollout status -w deployment/test-app

# get the .status.observedGeneration of test-app deployment

${CLI} apply -f ${WORKSPACE}/sbo.yaml --wait
${CLI} wait --for=condition=Ready=True servicebindings.binding.operators.coreos.com/test-sbo-chart-binding --timeout=15s

# wait for deployment to re-deploy
${CLI} rollout status -w deployment/test-app

exit_code=0

# Assertions
binding_data=$(curl test-app.$TEST_NAMESPACE.svc.cluster.local:8080/bindings/test-sbo-chart-binding/username)
if [ "$binding_data" != "foo" ]; then
    echo "Incorrect binding data ..."
    exit_code=1
fi

# get the .status.observedGeneration of test-app deployment and it should be > original

# Clean test resources
if [ "$KEEP_TEST_RESOURCES" != "true" ]; then
${CLI} delete -f ${WORKSPACE}/secret.yaml
${CLI} delete -f ${WORKSPACE}/application.yaml
${CLI} delete -f ${WORKSPACE}/sbo.yaml
fi

# Exit with exit code
exit $exit_code