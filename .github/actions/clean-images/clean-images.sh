#!/usr/bin/env bash

REPO=$1
# REPO_USERNAME=...
# REPO_PASSWORD=...
TAGS=${2:-}

if [ -z "$TAGS" ] || [ -z "$REPO" ]; then
    echo "Usage: $0 <repo> <images regex>"
    echo ""
    echo "Optionally set REPO_USERNAME and REPO_PASSWORD env variables to provide repo credentials."
    echo ""
    exit 1
fi

if [ -n "$REPO_USERNAME" ]; then
    REPO_CREDS="--creds ${REPO_USERNAME}:${REPO_PASSWORD}"
fi

for tag in $(skopeo list-tags --tls-verify=false docker://${REPO} | jq -r ".Tags[] | select(.? | match(\"${TAGS}\"))"); do
    echo "Deleting docker://${REPO}:${tag}"
    skopeo delete ${REPO_CREDS} docker://${REPO}:${tag}
done
