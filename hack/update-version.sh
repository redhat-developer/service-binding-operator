#!/bin/bash
# Update the operator version to a new version at various places across the repository.
# Refer https://semver.org/

set -e
set -u

MANIFESTS_DIR="./../manifests"
NEW_VERSION=$1

function current_version {
    filename="../Makefile"
    OPERATOR_VERSION=$(grep -m 1 OPERATOR_VERSION $filename | sed 's/^.*= //g')
}

current_version
OLD_VERSION="${OPERATOR_VERSION}"

function replace {
    LOCATION=$1
    if [ -e $LOCATION ] ; then
        sed -i -e 's/'${OLD_VERSION}'/'${NEW_VERSION}'/g' $LOCATION
    else
        echo ERROR: Failed to find $LOCATION
        exit 1 #terminate and indicate error
    fi
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
