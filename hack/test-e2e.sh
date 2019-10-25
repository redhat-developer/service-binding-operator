#!/bin/bash
set +x
export GO111MODULE=${GO111MODULE}
export GOCACHE=${GOCACHE}
export SERVICE_BINDING_OPERATOR_DISABLE_ELECTION=${SERVICE_BINDING_OPERATOR_DISABLE_ELECTION:-true}
export TEST_NAMESPACE=${TEST_NAMESPACE}
export ZAP_FLAGS=${ZAP_FLAGS}
export OPERATOR_SDK_EXTRA_ARGS=${OPERATOR_SDK_EXTRA_ARGS}
export LOGS_DIR=${LOGS_DIR}
export HACK_DIR=${HACK_DIR}
set -x

#Run the e2e tests with local operator
operator-sdk --verbose test local ./test/e2e \
	--namespace $TEST_NAMESPACE \
	--up-local \
	--go-test-flags "-timeout=15m" \
	--local-operator-flags "$ZAP_FLAGS" \
	${OPERATOR_SDK_EXTRA_ARGS} \
	| tee ${LOGS_DIR}/e2e/test-e2e.log

#Get test's exit code
TEST_RTN_CODE=${PIPESTATUS[0]}

#Extract the local operator log and replace the escape sequences by the actual whitespaces
${HACK_DIR}/e2e-log-parser.sh ${LOGS_DIR}/e2e/test-e2e.log > ${LOGS_DIR}/e2e/local-operator.log

#Exit with the tests' exit code
exit $TEST_RTN_CODE
