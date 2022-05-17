PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

include make/*.mk
