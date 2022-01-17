#!/usr/bin/env -S bash -e

# Required CLIs
# - https://github.com/opencontainers/umoci
# - https://github.com/containers/skopeo

# Usage:
# prepare-operatorhu-pr.sh <version> <bundle-image-ref>

mkdir -p out
TMP_OCI_PATH=$(mktemp -d out/sbo-bundle-oci.XXX)

skopeo copy --src-tls-verify=false --src-no-creds docker://$2 oci:$TMP_OCI_PATH:bundle
umoci unpack --image $TMP_OCI_PATH:bundle --rootless ${TMP_OCI_PATH}-unpacked
rm -rf out/operatorhub-pr-files
mkdir out/operatorhub-pr-files/service-binding-operator -p
mv ${TMP_OCI_PATH}-unpacked/rootfs out/operatorhub-pr-files/service-binding-operator/$1

cat <<EOD
Done.

Now you can submit the content of out/operatorhub-pr-files/service-binding-operator
to https://github.com/k8s-operatorhub/community-operators/tree/main/operators/service-binding-operator

Please check the content of out/operatorhub-pr-files/service-binding-operator/VERSION/metadata/annotations.yaml
and verify the channel where the operator will be published. Usually, the operator should be published to `stable`
channel.
EOD