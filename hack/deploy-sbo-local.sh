#!/bin/bash

OPERATOR_NAMESPACE=${OPERATOR_NAMESPACE:-}
ZAP_FLAGS=${ZAP_FLAGS:-}

SBO_LOCAL_LOG=out/sbo-local.log

killall operator-sdk service-binding-operator-local

operator-sdk --verbose run --local --namespace="$OPERATOR_NAMESPACE" --operator-flags "$ZAP_FLAGS" > $SBO_LOCAL_LOG 2>&1 &

SBO_PID=$!

attempts=12
while [ -z "$(grep 'Starting workers' $SBO_LOCAL_LOG)" ]; do
    if [[ $attempts -ge 0 ]]; then
        sleep 5
        attempts=$((attempts-1))
    else
        echo "FAILED"
        kill $SBO_PID
        exit 1
    fi
done

echo $SBO_PID
