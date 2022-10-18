# -- Variables for performance tests
TEST_PERFORMANCE_OUTPUT_DIR ?= $(OUTPUT_DIR)/performance
TEST_PERFORMANCE_ARTIFACTS ?= $(ARTIFACT_DIR)
TEST_PERFORMANCE_NS_PREFIX ?= entanglement
TEST_PERFORMANCE_USERS_PER_SCENARIO ?= 400
OPENSHIFT_API ?=
OPENSHIFT_USERNAME ?=
OPENSHIFT_PASSWORD ?=

TEST_PERFORMANCE_AVG_MEMORY ?= 150 # value in MiB
TEST_PERFORMANCE_MAX_MEMORY ?= 200 # valie in MiB
TEST_PERFORMANCE_AVG_CPU ?= 20 # value in millicores of vCPU
TEST_PERFORMANCE_MAX_CPU ?= 100 # value in millicores of vCPU

.PHONY: test-performance-setup
## Setup OpenShift cluster for performance test
test-performance-setup:
	@oc login $(OPENSHIFT_API) -u $(OPENSHIFT_USERNAME) -p $(OPENSHIFT_PASSWORD) --insecure-skip-tls-verify=true
	QUAY_NAMESPACE=$(REGISTRY_NAMESPACE) ./test/performance/setup.sh

.PHONY: test-performance
## Run performance test
test-performance: export CATSRC_NAME=sbo-test-performance
test-performance: test-performance-setup deploy-from-index-image
	OUTPUT_DIR=$(TEST_PERFORMANCE_OUTPUT_DIR) ./test/performance/run.sh $(TEST_PERFORMANCE_NS_PREFIX) $(TEST_PERFORMANCE_USERS_PER_SCENARIO)

.PHONY: test-performance-collect-kpi
## Collect KPI (Key Performance Indicators)
test-performance-collect-kpi:
	METRICS=$(TEST_PERFORMANCE_OUTPUT_DIR)/metrics RESULTS=$(TEST_PERFORMANCE_OUTPUT_DIR)/results  NS_PREFIX=$(TEST_PERFORMANCE_NS_PREFIX) ./test/performance/collect-kpi.sh

.PHONY: test-performance-artifacts
# Collect artifacts from performance test to be archived in CI
test-performance-artifacts:
	$(Q)echo "Gathering performance test artifacts"
	$(Q)mkdir -p $(TEST_PERFORMANCE_ARTIFACTS) \
		&& cp -rvf $(TEST_PERFORMANCE_OUTPUT_DIR) $(TEST_PERFORMANCE_ARTIFACTS)/

.PHONY: test-performance-thresholds
# Compare various KPIs to the threshold values and fail if any threshold is exceeded
test-performance-thresholds: yq
	@echo Evaluating KPI:
	@$(YQ) eval '.kpi[] | select(.name == "usage")' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml
	@echo "Checking if average value of memory "$$($(YQ) eval '(.kpi[] | select(.name == "usage").metrics.[] | select(.name == "Memory_MiB").average)' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml)" < $(TEST_PERFORMANCE_AVG_MEMORY) MiB"
	@[ $$($(YQ) eval '(.kpi[] | select(.name == "usage").metrics.[] | select(.name == "Memory_MiB").average) < $(TEST_PERFORMANCE_AVG_MEMORY)' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml) == "true" ]
	@echo "Checking if maximal value of memory "$$($(YQ) eval '(.kpi[] | select(.name == "usage").metrics.[] | select(.name == "Memory_MiB").maximum)' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml)" < $(TEST_PERFORMANCE_MAX_MEMORY) MiB"
	@[ $$($(YQ) eval '(.kpi[] | select(.name == "usage").metrics.[] | select(.name == "Memory_MiB").maximum) < $(TEST_PERFORMANCE_MAX_MEMORY)' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml) == "true" ]
	@echo "Checking if average value of CPU "$$($(YQ) eval '(.kpi[] | select(.name == "usage").metrics.[] | select(.name == "CPU_millicores").average)' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml)" < $(TEST_PERFORMANCE_AVG_CPU) milicores of vCPU"
	@[ $$($(YQ) eval '(.kpi[] | select(.name == "usage").metrics.[] | select(.name == "CPU_millicores").average) < $(TEST_PERFORMANCE_AVG_CPU)' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml) == "true" ]
	@echo "Checking if maximal value of CPU "$$($(YQ) eval '(.kpi[] | select(.name == "usage").metrics.[] | select(.name == "CPU_millicores").maximum)' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml)" < $(TEST_PERFORMANCE_MAX_CPU) milicores of vCPU"
	@[ $$($(YQ) eval '(.kpi[] | select(.name == "usage").metrics.[] | select(.name == "CPU_millicores").maximum) < $(TEST_PERFORMANCE_MAX_CPU)' $(TEST_PERFORMANCE_OUTPUT_DIR)/results/kpi.yaml) == "true" ]
