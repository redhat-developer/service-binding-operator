#!/bin/bash -xe

GENERIC_CHECK_PATCH_PATH=$(which generic-check-patch)

source $GENERIC_CHECK_PATCH_PATH || true

# override generic function so that we can use our versioning schemas
ci_get_xyz_version() {
    echo "0 0 23"
}

for container in $(find distgit/containers -mindepth 1 -maxdepth 1 -type d); do
  cat automation/labels >> "$container/Dockerfile.in"
done

ci_main "$0"