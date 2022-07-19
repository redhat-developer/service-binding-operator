#!/bin/bash

set -x
OPNS="$(kubectl get namespaces -o name | grep -E '.*/(.+-)?operators$' || test $? = 1)"
OPERATOR_NAMESPACE=${OPERATOR_NAMESPACE:-${OPNS#"namespace/"}}
OLM_NAMESPACE="${OLM_NAMESPACE:-$(kubectl get catalogsources.operators.coreos.com --all-namespaces -o jsonpath='{.items[0].metadata.namespace}' --ignore-not-found=true)}"
OUTPUT_PATH=${OUTPUT_PATH:-out/acceptance-tests/resources}
TEST_NAMESPACE_FILE=${TEST_NAMESPACE_FILE:-$OUTPUT_DIR/test-namespace}
set +x

echo "Collecting Kubernetes resources..."

kubectl api-resources --verbs=list --namespaced -o name >resources.namespaced.list
kubectl api-resources --verbs=list --namespaced=false -o name >resources.cluster-wide.list

for ns in ${OPERATOR_NAMESPACE} ${OLM_NAMESPACE}; do
    OUTPUT=${OUTPUT_PATH}/${ns}
    mkdir -p ${OUTPUT}
    for res in $(cat resources.namespaced.list | grep -v secrets); do
        kubectl get ${res} --ignore-not-found -n ${ns} -o yaml >${OUTPUT}/${res}.yaml
    done
    mkdir -p $OUTPUT/pod
    for p in $(kubectl get pods -n ${ns} -o name); do
        for c in $(kubectl get $p -n ${ns} -o jsonpath='{.spec.containers[].name}'); do
            kubectl logs -n ${ns} $p -c $c >$OUTPUT/$p.$c.log
        done
    done
    find ${OUTPUT} -size 0 -delete
done

if [ -f ${TEST_NAMESPACE_FILE} ]; then
    ns=$(cat ${TEST_NAMESPACE_FILE})
    OUTPUT=${OUTPUT_PATH}/${ns}
    mkdir -p ${OUTPUT}
    for res in $(cat resources.namespaced.list); do
        kubectl get ${res} --ignore-not-found -n ${ns} -o yaml >${OUTPUT}/${res}.yaml
    done
    mkdir -p $OUTPUT/pod
    for p in $(kubectl get pods -n ${ns} -o name); do
        for c in $(kubectl get $p -n ${ns} -o jsonpath='{.spec.containers[].name}'); do
            kubectl logs -n ${ns} $p -c $c >$OUTPUT/$p.$c.log
        done
    done
    find ${OUTPUT} -size 0 -delete
fi

OUTPUT=${OUTPUT_PATH}
for res in $(cat resources.cluster-wide.list); do
    kubectl get ${res} --ignore-not-found -o yaml >${OUTPUT}/${res}.yaml
done

echo "Collecting operator logs..."
mkdir -p ${OUTPUT_PATH}
operator_pod=$(kubectl get pods --no-headers -o custom-columns=":metadata.name" -n ${OPERATOR_NAMESPACE} | tail -1)
kubectl logs ${operator_pod} -n ${OPERATOR_NAMESPACE} >${OUTPUT_PATH}/${operator_pod}.log
