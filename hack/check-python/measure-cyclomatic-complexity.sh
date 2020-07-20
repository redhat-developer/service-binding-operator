#!/bin/bash

. ./hack/check-python/prepare-env.sh

[ "$NOVENV" == "1" ] || prepare_venv || exit 1

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

pushd "${SCRIPT_DIR}/.."
radon cc -s -a -i venv .
popd
