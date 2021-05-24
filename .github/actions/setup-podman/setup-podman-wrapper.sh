#!/usr/bin/env bash

set -x

mkdir -p $HOME/.config/containers
sed -e "s,REGISTRY_PREFIX,${REGISTRY_PREFIX},g" ./.github/actions/setup-podman/registries_template.conf > $HOME/.config/containers/registries.conf
cp -rvf ./.github/actions/setup-podman/podman ${GITHUB_WORKSPACE}/bin/podman
