#!/usr/bin/env bash

failed=""
CHECK_PYTHON=./hack/check-python
$CHECK_PYTHON/detect-common-errors.sh || failed="$failed\n - detect-common-errors"
$CHECK_PYTHON/detect-dead-code.sh || failed="$failed\n - detect-dead-code"
$CHECK_PYTHON/check-PEP8-style.sh || failed="$failed\n - check-PEP8-style"
#$CHECK_PYTHON/check-python-docstyle.sh || failed="$failed\n - check-python-docstyle"
$CHECK_PYTHON/measure-cyclomatic-complexity.sh || failed="$failed\n - measure-cyclomatic-complexity"
$CHECK_PYTHON/measure-maintainability-index.sh || failed="$failed\n - measure-maintainability-index"

if [ ! -z "$failed" ]; then
    echo -e "\nERROR: Following python checks FAILED:$failed\n"
    exit 1
else
    echo -e "\nAll python checks PASSED"
fi
