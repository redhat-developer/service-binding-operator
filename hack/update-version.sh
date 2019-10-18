#!/bin/bash
# Update the operator version to a new version at various places across the repository.
# Refer https://semver.org/

MANIFESTS_DIR="./../manifests"
OPERATOR_VERSION="0.0.20"
OLD_VERSION="${OPERATOR_VERSION}"
NEW_VERSION=$1

function replace {
    LOCATION=$1
    sed -i -e 's/'${OLD_VERSION}'/'${NEW_VERSION}'/g' $LOCATION
}
replace ../Makefile
replace ./update-version.sh
replace ${MANIFESTS_DIR}/service-binding-operator.package.yaml
replace ${MANIFESTS_DIR}/service-binding-operator.v${OLD_VERSION}.clusterserviceversion.yaml
mv ${MANIFESTS_DIR}/service-binding-operator.v${OLD_VERSION}.clusterserviceversion.yaml \
${MANIFESTS_DIR}/service-binding-operator.v${NEW_VERSION}.clusterserviceversion.yaml
replace ./../openshift-ci/Dockerfile.registry.build
echo -e "\n\033[0;32m \xE2\x9C\x94 Operator version upgraded from \
${OLD_VERSION} to ${NEW_VERSION} \033[0m\n"
