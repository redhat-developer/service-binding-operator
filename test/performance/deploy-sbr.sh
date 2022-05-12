#!/bin/bash -x

SBR_YAML=${1:-sbo-test.sbr.yaml}

APP_NAME=$(yq e '.spec.application.name' $SBR_YAML)
SBR_NAME=$(yq e '.metadata.name' $SBR_YAML)
USER_NS_PREFIX=${2:-entanglement}
oc get deploy --all-namespaces -o json | jq -rc '.items[] | select(.metadata.name | contains("'$APP_NAME'")).metadata.namespace' | grep $USER_NS_PREFIX >workload.namespace.list

no_ns=$(cat workload.namespace.list | wc -l)

split -l $((no_ns / 5)) workload.namespace.list sbr-segment

for i in sbr-segment*; do
    for j in $(cat $i); do
        oc apply -f $SBR_YAML -n $j --server-side=true
        sleep 0.02s
    done &
done

wait

#Wait for the all the service bindings to get status
retries=360
until [[ $retries == 0 ]]; do
    sb_with_status_set=$(oc get sbr --all-namespaces -o json | jq -rc '.items[]| select(.metadata.namespace | contains("'$USER_NS_PREFIX'")) | select(.status != null).metadata.name' | wc -l)
    [ $no_ns != $sb_with_status_set ] || break
    echo "Waiting for all the Service Binding resources to be processed by operator... currently only $sb_with_status_set/$no_ns are"
    sleep 10
    retries=$(($retries - 1))
done

rm -rf sbr-segment*
rm -rf workload.namespace.list
