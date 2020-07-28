#!/bin/bash

OPERATOR_NAMESPACE=${OPERATOR_NAMESPACE:-}
ZAP_FLAGS=${ZAP_FLAGS:-}

OUTPUT="${OUTPUT:-out/acceptance-tests}"
mkdir -p "$OUTPUT"

SBO_LOCAL_LOG="$OUTPUT/sbo-local.log"

_killall(){
    which killall &> /dev/null
    if [ $? -eq 0 ]; then
        killall $1
    else
        for i in "$(ps -l | grep $1)"; do if [ -n "$i" ]; then kill $(echo "$i" | sed -e 's,\s\+,#,g' | cut -d "#" -f4); fi; done
    fi
}

_killall operator-sdk
_killall service-binding-operator-local

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
