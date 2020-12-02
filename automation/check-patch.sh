#!/bin/bash -xe

source "${0%/*}/generic-check-patch" || true

# override generic function so that we can use our versioning schemas
ci_get_xyz_version() {
    echo "0 3 0"
}

ci_get_branch() {
    echo "${GERRIT_BRANCH}-container"
}

for container in $(find distgit/containers -mindepth 1 -maxdepth 1 -type d); do
  cat automation/labels >> "$container/Dockerfile.in"
  cp automation/distgit-gitignore "$container/.gitignore"
done

case "$0" in
    *check-patch*|scratch-build)
        CI_DIGEST_PINNING=false
        ;;
    *check-merged*|push)
        CI_DIGEST_PINNING=true
        ;;
    *)
        echo "Unkown action $action" 1>&2
        echo "Can be $allowed_actions" 1>&2
        exit 1
esac

export CI_DIGEST_PINNING

ci_main "$0"