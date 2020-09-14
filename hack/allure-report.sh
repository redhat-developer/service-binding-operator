#!/bin/bash -ex

CONTAINER_RUNTIME=${CONTAINER_RUNTIME:-docker}

$CONTAINER_RUNTIME info 2>/dev/null >/dev/null || NO_CONTAINER_RUNTIME=$?
if [ -n "$NO_CONTAINER_RUNTIME" ]; then
    echo "WARNING $CONTAINER_RUNTIME is required but not installed, skipping..."
    exit 1
fi

ACTION=${1:-generate}

NAME=test-acceptance-report

REPORT_DIR=$(pwd)/out/acceptance-tests-report
mkdir -p $REPORT_DIR

$CONTAINER_RUNTIME build -f test/acceptance/Dockerfile.allure -t $NAME test/acceptance

$CONTAINER_RUNTIME rm -f $NAME || true

if [ $ACTION == "generate" ]; then
    $CONTAINER_RUNTIME run --rm --name $NAME -v $(pwd)/out/acceptance-tests:/allure/results -v $REPORT_DIR:/allure/report $NAME $ACTION
    echo -e "\nAcceptance tests report was generated into $REPORT_DIR directory" 
elif [ $ACTION == "serve" ]; then
    $CONTAINER_RUNTIME run -d --rm --name $NAME -v $(pwd)/out/acceptance-tests:/allure/results -p 8088:8080 $NAME $ACTION
    CID=$($CONTAINER_RUNTIME ps -a -f name=$NAME -q)
    echo -e "\nAcceptance tests report is running in background in $CONTAINER_RUNTIME container called '$NAME' ($CID) and can be accessed at http://localhost:8088"
fi
