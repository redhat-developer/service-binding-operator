#!/bin/bash -e

DT=$(date "+%F_%T")
RESULTS=${RESULTS:-results-$DT}
mkdir -p $RESULTS

USER_NS_PREFIXES=${1:-entanglement}
PROCESS_ONLY=${PROCESS_ONLY:-}
SKIP_RESOURCE_COUNTS=${SKIP_RESOURCE_COUNTS:-}

ddiff_sec() {
    secdiff=$(echo "$(date -d "$2" +%s) - $(date -d "$1" +%s)" | bc | sed -e 's,^\.\([0-9]\+\),0.\1,')
    nanosecdiff=$(echo "$(date -d "$2" +%N) - $(date -d "$1" +%N)" | bc | sed -e 's,^\.\([0-9]\+\),0.\1,')
    echo "scale=9; $secdiff + ($nanosecdiff / 1000000000)" | bc | sed -e 's,^\.\([0-9]\+\),0.\1,'
}

# Resource counts
resource_counts() {
    echo -n "$1;"
    # All resource counts from user namespaces
    regex=$(echo $2 | sed -e 's,\([^ ]\) \([^ ]\),\1|\2,g')
    echo -n "$(oc get $1 --all-namespaces -o custom-columns=NAMESPACE:.metadata.namespace --ignore-not-found=true | grep -E "$regex"| wc -l)"
    echo -n ";"
    # All resource counts from all namespaces
    echo "$(oc get $1 --all-namespaces -o name --ignore-not-found=true | wc -l)"
}

# Dig various timestamps out
timestamps() {
    NS_PREFIX=$1
    SBR_JSON=$2
    SB_API=${3:-binding.operators.coreos.com}
    WORKLOADS_JSON=$4
    SBO_LOG=$5
    RESULTS=$6

    BINDING_TIMESTAMPS_CSV=${NS_PREFIX}.${SB_API}.binding-timestamps.csv
    SBR_TIMESTAMPS_CSV=${NS_PREFIX}.${SB_API}.timestamps.csv
    TMP_CSV=${NS_PREFIX}.${SB_API}.tmp.csv

    LOG_SEG_DIR=$RESULTS/sbo-log-segments
    mkdir -p $LOG_SEG_DIR

    rm -f $RESULTS/$TMP_CSV
    jq -rc 'select(.metadata.namespace | startswith("'$NS_PREFIX'")) | ((.metadata.namespace) + ";" + (.metadata.name) + ";" + (.metadata.creationTimestamp))' $SBR_JSON >$RESULTS/$TMP_CSV
    echo "ServiceBinding;Created;ReconciledTimestamp;AllDoneTimestamp;TimeToReconcile;ReconciledToDone;TimeToDone" >$RESULTS/$SBR_TIMESTAMPS_CSV
    for i in $(cat $RESULTS/$TMP_CSV); do
        ns=$(echo -n $i | cut -d ";" -f1)
        name=$(echo -n $i | cut -d ";" -f2)

        # ServiceBinding
        echo -n $ns/$name
        echo -n ";"

        # Created
        created=$(date -d $(echo -n $i | cut -d ";" -f3) "+%F %T")
        echo -n "$created"
        echo -n ";"

        # ReconciledTimestamp
        log=$LOG_SEG_DIR/$ns.log
        cat $SBO_LOG | grep $ns >$log
        reconciled=$(date -d @$(cat $log | jq -rc 'if .serviceBinding != null then select(.serviceBinding | contains("'$ns/$name'")) | select(.msg | contains("Reconciling")).ts else empty end' | head -n1) "+%F %T.%N" | tr -d "\n")
        echo -n "$reconciled"
        echo -n ";"

        # AllDoneTimestamp
        done_ts=$(cat "$log" | jq -rc 'select(.msg | contains("Done")) | select(.serviceBinding | contains("'$ns/$name'")) | select(.retry==false).ts' | tail -n 1)
        if [ -n "$done_ts" ]; then
            all_done=$(date -d "@$done_ts" "+%F %T.%N")
        else
            all_done=""
        fi
        echo -n "$all_done"
        echo -n ";"

        # TimeToReconciled
        echo -n "$(ddiff_sec "$created" "$reconciled")"
        echo -n ";"

        # ReconciledToDone
        echo -n "$(ddiff_sec "$reconciled" "$all_done")"
        echo -n ";"

        # TimeToDone
        echo -n "$(ddiff_sec "$created" "$all_done")"

        echo ""

    done \
        >>$RESULTS/$SBR_TIMESTAMPS_CSV

    rm -f $RESULTS/$TMP_CSV
    jq -rc 'select(.metadata.namespace | startswith("'$NS_PREFIX'")) | ((.metadata.namespace) + ";" + (.metadata.name))' $WORKLOADS_JSON >$RESULTS/$TMP_CSV
    echo "Namespace;Workload;SB_Name;SB_created;SB_ReconciledTimestamp;SB_AllDoneTimestamp;SB_TimeToReconcile;SB_ReconciledToDone;SB_TimeToDone" >$RESULTS/$BINDING_TIMESTAMPS_CSV
    for i in $(cat $RESULTS/$TMP_CSV); do
        echo -n $i
        echo -n ";"
        NS=$(echo -n $i | cut -d ";" -f1)
        cat $RESULTS/$SBR_TIMESTAMPS_CSV | grep $NS
    done >>$RESULTS/$BINDING_TIMESTAMPS_CSV
    rm -f $RESULTS/$TMP_CSV
}

# Collect timestamps
{

    # ServiceBiding operator log
    if [ -z "$PROCESS_ONLY" ]; then
        oc logs $(oc get $(oc get pods -n openshift-operators -o name | grep service-binding-operator) -n openshift-operators -o jsonpath='{.metadata.name}') -n openshift-operators >$RESULTS/service-binding-operator.log
    fi

    
    for ns_prefix in $USER_NS_PREFIXES; do
        workloads_json=$RESULTS/$ns_prefix.workloads.json
        if [ -z "$PROCESS_ONLY" ]; then
            # Workload resources in user namespaces
            oc get deployment --all-namespaces -o json | jq -r '.items[] | select(.metadata.namespace | startswith("'$ns_prefix'" )) | select(.metadata.name | contains("sbo-perf-app"))' > $workloads_json
            oc get cronjobs --all-namespaces -o json | jq -r '.items[] | select(.metadata.namespace | startswith("'$ns_prefix'" )) | select(.metadata.name | contains("sbo-perf-cronjob"))' >> $workloads_json
            
            # ServiceBinding resources in user namespaces
            for sb_api in binding.operators.coreos.com servicebinding.io; do
                oc get servicebindings.$sb_api --all-namespaces -o json | jq -r '.items[] | select(.metadata.namespace | startswith("'$ns_prefix'"))' >$RESULTS/$ns_prefix.$sb_api.json
                # Service Binding Timestamps 
                timestamps $ns_prefix $RESULTS/$ns_prefix.$sb_api.json $sb_api $workloads_json $RESULTS/service-binding-operator.log $RESULTS
            done
        fi
    done
} &

# Collect resource counts
{
    if [ -z "$PROCESS_ONLY" ] && [ -z "$SKIP_RESOURCE_COUNTS" ]; then
        oc api-resources --verbs=list --namespaced -o name | sort >resource-list.namespaced
        oc api-resources --verbs=list --namespaced=false -o name | sort >resource-list.cluster

        RESOURCE_COUNTS_OUT=$RESULTS/resource-count.csv
        echo "Resource;UserNamespaces;AllNamespaces" >$RESOURCE_COUNTS_OUT
        for i in $(cat resource-list.namespaced resource-list.cluster | sort); do
            resource_counts $i "$USER_NS_PREFIXES" >>$RESOURCE_COUNTS_OUT
        done
    fi
} &

wait
