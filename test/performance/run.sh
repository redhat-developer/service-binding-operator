#!/bin/bash -x

WS=${WS:-$(readlink -m $(dirname $0))}
OUTPUT_DIR=${OUTPUT_DIR:-$WS/out}
WAIT_BEFORE=${WAIT_BEFORE:-60s}
WAIT_AFTER=${WAIT_AFTER:-60s}

mkdir -p $OUTPUT_DIR

$WS/collect-metrics.sh 30 $OUTPUT_DIR/metrics &
METRICS_PID=$!

trap "kill -9 $METRICS_PID" EXIT

sleep $WAIT_BEFORE

pushd $WS/toolchain-e2e.git

USER_NS_PREFIX=${1:-entanglement}
USERS_PER_SCENARIO=${2:-25}
DEFAULT_WORKLOADS=${DEFAULT_WORKLOADS:-$USERS_PER_SCENARIO}

SBO_NAMESPACE=${SBO_NAMESPACE:-openshift-operators}
SBO_DEPLOYMENT=${SBO_DEPLOYMENT:-service-binding-operator}

yes | go run setup/main.go --template=$WS/user-workloads/valid/sbo-test.with-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-sb-val > $OUTPUT_DIR/$USER_NS_PREFIX-sb-val.log

yes | go run setup/main.go --template=$WS/user-workloads/valid/sbo-test.without-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-nosb-val > $OUTPUT_DIR/$USER_NS_PREFIX-nosb-val.log
$WS/deploy-sbr.sh $WS/user-workloads/valid/sbo-test.sbr.yaml $USER_NS_PREFIX-nosb-val > $OUTPUT_DIR/$USER_NS_PREFIX-nosb-val.deploy-sbr.log
$WS/deploy-sbr.sh $WS/user-workloads/valid/sbo-test.sbr.spec.yaml $USER_NS_PREFIX-nosb-val >> $OUTPUT_DIR/$USER_NS_PREFIX-nosb-val.deploy-sbr.log

yes | go run setup/main.go --template=$WS/user-workloads/invalid/sbo-test.with-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-sb-inv > $OUTPUT_DIR/$USER_NS_PREFIX-sb-inv.log

yes | go run setup/main.go --template=$WS/user-workloads/invalid/sbo-test.without-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-nosb-inv > $OUTPUT_DIR/$USER_NS_PREFIX-nosb-inv.log
$WS/deploy-sbr.sh $WS/user-workloads/invalid/sbo-test.sbr.yaml $USER_NS_PREFIX-nosb-inv > $OUTPUT_DIR/$USER_NS_PREFIX-nosb-inv.deploy-sbr.log
$WS/deploy-sbr.sh $WS/user-workloads/invalid/sbo-test.sbr.spec.yaml $USER_NS_PREFIX-nosb-inv >> $OUTPUT_DIR/$USER_NS_PREFIX-nosb-inv.deploy-sbr.log

yes | go run setup/main.go --template=$WS/user-workloads/incomplete/sbo-test.with-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-sb-inc > $OUTPUT_DIR/$USER_NS_PREFIX-sb-inc.log

sleep $WAIT_AFTER
