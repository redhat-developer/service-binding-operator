#!/bin/bash

WS=${WS:-$(readlink -m $(dirname $0))}

indent() {
    sed 's/^/  /'
}

DT=$(date "+%F_%T")
export RESULTS=${RESULTS:-results-$DT}
mkdir -p $RESULTS

export METRICS=${METRICS:-$(find . -type d -name 'metrics-*')}

NS_PREFIX=${NS_PREFIX:-entanglement}

SBO_METRICS=$(find $METRICS -type f -name 'pod-info.service-binding-operator-*.csv')

SCENARIOS="nosb-inv nosb-val sb-inc sb-inv sb-val"
#SCENARIOS="nosb-val"

USER_NS_PREFIXES=""
for scenario in $SCENARIOS; do
    USER_NS_PREFIXES="$USER_NS_PREFIXES $NS_PREFIX-$scenario"
done

kpi_yaml=$RESULTS/kpi.yaml
echo "kpi:" >$kpi_yaml

output=$RESULTS/sbo-metrics.kpi.yaml
echo "- name: usage" >$output
echo "  metrics:" >>$output

python $WS/kpi.py -c $SBO_METRICS -x 0 -y 1 | indent >>$output
python $WS/kpi.py -c $SBO_METRICS -x 0 -y 2 | indent >>$output

cat $output >>$kpi_yaml

if [ -z $PROCESS_ONLY ]; then
    $WS/collect-results.sh "${USER_NS_PREFIXES}"
fi

for scenario in $SCENARIOS; do
    for sb_api in binding.operators.coreos.com servicebinding.io; do
        ns_prefix=$NS_PREFIX-$scenario
        output=$RESULTS/$scenario.$sb_api.kpi.yaml
        echo "- name: $scenario.$sb_api" >$output
        echo "  metrics:" >>$output
        for i in 4 5 6; do
            python $WS/kpi.py -c "$(readlink -m $RESULTS/$ns_prefix.$sb_api.timestamps.csv)" -x 1 -y $i -d "%Y-%m-%d %H:%M:%S" | indent >>$output
        done
        cat $output >>$kpi_yaml
    done
done
