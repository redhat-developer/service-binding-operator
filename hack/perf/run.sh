#!/bin/bash -x

pushd $(dirname $0)
    source ./env.sh
    pushd ../../
        time make clean test-performance test-performance-collect-kpi -o deploy-from-index-image
    popd
popd
