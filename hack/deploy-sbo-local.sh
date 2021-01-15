#!/bin/bash -x

ZAP_FLAGS=${ZAP_FLAGS:-}

OUTPUT="${OUTPUT:-out/acceptance-tests}"
mkdir -p "$OUTPUT"

SBO_LOCAL_LOG="$OUTPUT/sbo-local.log"

RUN_IN_BACKGROUND="${RUN_IN_BACKGROUND:-false}"

_run_operator(){
    SERVICE_BINDING_OPERATOR_DISABLE_ELECTION=false WATCH_NAMESPACE="" ./bin/manager $ZAP_FLAGS
}

if [ "$RUN_IN_BACKGROUND" == "true" ]; then
    _run_operator > $SBO_LOCAL_LOG 2>&1 &

    SBO_PID=$!

    attempts=24
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
else
    _run_operator
fi