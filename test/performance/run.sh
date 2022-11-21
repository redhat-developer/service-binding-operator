#!/bin/bash -x

WS=${WS:-$(readlink -m $(dirname $0))}
OUTPUT_DIR=${OUTPUT_DIR:-$WS/out}
WAIT_BEFORE=${WAIT_BEFORE:-60s}
WAIT_AFTER=${WAIT_AFTER:-60s}

mkdir -p $OUTPUT_DIR

$WS/collect-metrics.sh 30 $OUTPUT_DIR/metrics &
METRICS_PID=$!

trap "kill -9 $METRICS_PID" EXIT

echo "Waiting for $WAIT_BEFORE to gather resource utilization before the load is started"
sleep $WAIT_BEFORE

pushd $WS/toolchain-e2e.git

USER_NS_PREFIX=${1:-entanglement}
USERS_PER_SCENARIO=${2:-25}
DEFAULT_WORKLOADS=${DEFAULT_WORKLOADS:-$USERS_PER_SCENARIO}

SBO_NAMESPACE=${SBO_NAMESPACE:-openshift-operators}
SBO_DEPLOYMENT=${SBO_DEPLOYMENT:-service-binding-operator}

echo "Applying cluster-wide resources"
oc apply -f $WS/user-workloads/cluster-wide/sbo-test.yaml

echo "Running the performance scenario: sb-val"
yes | go run setup/main.go --template=$WS/user-workloads/valid/sbo-test.with-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-sb-val > $OUTPUT_DIR/$USER_NS_PREFIX-sb-val.log

echo "Running the performance scenario: nosb-val"
yes | go run setup/main.go --template=$WS/user-workloads/valid/sbo-test.without-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-nosb-val > $OUTPUT_DIR/$USER_NS_PREFIX-nosb-val.log
$WS/deploy-sbr.sh $WS/user-workloads/valid/sbo-test.sbr.yaml $USER_NS_PREFIX-nosb-val > $OUTPUT_DIR/$USER_NS_PREFIX-nosb-val.deploy-sbr.log
$WS/deploy-sbr.sh $WS/user-workloads/valid/sbo-test.sbr.spec.yaml $USER_NS_PREFIX-nosb-val >> $OUTPUT_DIR/$USER_NS_PREFIX-nosb-val.deploy-sbr.log

echo "Running the performance scenario: sb-inv"
yes | go run setup/main.go --template=$WS/user-workloads/invalid/sbo-test.with-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-sb-inv > $OUTPUT_DIR/$USER_NS_PREFIX-sb-inv.log

echo "Running the performance scenario: nosb-inv"
yes | go run setup/main.go --template=$WS/user-workloads/invalid/sbo-test.without-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-nosb-inv > $OUTPUT_DIR/$USER_NS_PREFIX-nosb-inv.log
$WS/deploy-sbr.sh $WS/user-workloads/invalid/sbo-test.sbr.yaml $USER_NS_PREFIX-nosb-inv > $OUTPUT_DIR/$USER_NS_PREFIX-nosb-inv.deploy-sbr.log
$WS/deploy-sbr.sh $WS/user-workloads/invalid/sbo-test.sbr.spec.yaml $USER_NS_PREFIX-nosb-inv >> $OUTPUT_DIR/$USER_NS_PREFIX-nosb-inv.deploy-sbr.log

echo "Running the performance scenario: sb-inc"
yes | go run setup/main.go --template=$WS/user-workloads/incomplete/sbo-test.with-sbr.user-workloads.yaml --users $USERS_PER_SCENARIO --default $DEFAULT_WORKLOADS --custom $USERS_PER_SCENARIO --operators-limit 0 --workloads $SBO_NAMESPACE:$SBO_DEPLOYMENT --username $USER_NS_PREFIX-sb-inc > $OUTPUT_DIR/$USER_NS_PREFIX-sb-inc.log

echo "Waiting for $WAIT_AFTER to gather resource utilization after the load is done"
sleep $WAIT_AFTER
