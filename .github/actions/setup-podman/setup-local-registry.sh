#!/usr/bin/env bash

set -x

mkdir -p ${GITHUB_WORKSPACE}/registry

podman run -d -p 5000:5000 --rm -v ${GITHUB_WORKSPACE}/registry:/var/lib/registry:Z --name reg registry:2.7
