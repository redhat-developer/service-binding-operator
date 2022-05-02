#!/bin/bash

PERIOD="${1:-30}"
RESULTS=${2:-metrics-$(date "+%F_%T")}

mkdir -p "$RESULTS"

strip_unit() {
    echo -n $1 | sed -e 's,\([0-9]\+\)m,\1,g' | sed -e 's,\([0-9]\+\)Mi,\1,g' | sed -e 's,\([0-9]\+\)%,\1,g'
}

# Nodes
oc get nodes >$RESULTS/nodes.yaml
oc describe nodes >$RESULTS/nodes.info

node_info_file() {
    readlink -m "$RESULTS/node-info.$1.csv"
}

node_line() {
    node=$1
    node_json="$(oc get node $node -o json)"
    echo -n "$(echo "$node_json" | jq -rc '.status.conditions[] | select(.type=="MemoryPressure").status');"
    echo -n "$(echo "$node_json" | jq -rc '.status.conditions[] | select(.type=="DiskPressure").status');"
    echo -n "$(echo "$node_json" | jq -rc '.status.conditions[] | select(.type=="PIDPressure").status');"
    echo -n "$(echo "$node_json" | jq -rc '.status.conditions[] | select(.type=="Ready").status');"
    node_info=($(oc adm top node $node --no-headers))
    echo -n "$(strip_unit ${node_info[1]});"
    echo -n "$(strip_unit ${node_info[2]});"
    echo -n "$(strip_unit ${node_info[3]});"
    echo "$(strip_unit ${node_info[4]})"
}

NODES=($(oc get nodes -o json | jq -rc '.items[].metadata.name' | sort))
for node in "${NODES[@]}"; do
    echo "Time;MemoryPressure;DiskPressure;PIDPressure;Ready;CPU_millicores;CPU_percent;Memory_MiB;Memory_percent" >$(node_info_file $node)
done

# Operator pods
pod_info_file() {
    readlink -m "$RESULTS/pod-info.$1.csv"
}

pod_line() {
    pod=$1
    ns=$2
    pod_info=($(oc adm top pod $pod -n $ns --no-headers))
    echo -n "$(strip_unit ${pod_info[1]});"
    echo "$(strip_unit ${pod_info[2]})"
}

for namespace in openshift-operators openshift-monitoring openshift-apiserver openshift-kube-apiserver openshift-sdn openshift-operator-lifecycle-manager; do
    PODS=($(oc get pods -n $namespace -o json | jq -rc '.items[].metadata.name' | grep -E 'operator|prometheus|apiserver|sdn|ovs|olm|packageserver' | sort))
    for pod in "${PODS[@]}"; do
        echo "Time;CPU_millicores;Memory_MiB" >$(pod_info_file $pod)
    done
done

echo "Collecting metrics"
# Periodical collection
while true; do
    echo -n "."
    for namespace in openshift-operators openshift-monitoring openshift-apiserver openshift-kube-apiserver openshift-sdn openshift-operator-lifecycle-manager; do
        PODS=($(oc get pods -n $namespace -o json | jq -rc '.items[].metadata.name' | grep -E 'operator|prometheus|apiserver|sdn|ovs|olm|packageserver' | sort))
        for pod in "${PODS[@]}"; do
            pod_file=$(pod_info_file $pod)
            echo -n "$(date -u '+%F %T.%N');" >>$pod_file
            pod_line $pod $namespace >>$pod_file
        done
    done
    for node in ${NODES[@]}; do
        node_file=$(node_info_file $node)
        echo -n "$(date -u '+%F %T.%N');" >>$node_file
        node_line $node >>$node_file
    done
    sleep ${PERIOD}s
done
