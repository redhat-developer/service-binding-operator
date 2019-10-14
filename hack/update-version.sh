#!/bin/bash
# Update the operator version from a version 0.0.(n) to a version 0.0.(n+1) at various places across the repository.
OPERATOR_VERSION="0.0.20"
MANIFESTS_DIR="./../manifests"
OPERATOR_VERSION_UPDATE=$1

if [ "${OPERATOR_VERSION_UPDATE}" != "${OPERATOR_VERSION}" ] ; then
	NEW_VERSION="$(echo ${OPERATOR_VERSION_UPDATE} | tr '.' '\n' | tail -1)"
	OLD_VERSION="$(echo ${OPERATOR_VERSION} | tr '.' '\n' | tail -1)"
	EXPECTED_VERSION="$(echo ${OLD_VERSION} + 1 | bc )"
	if [ "${NEW_VERSION}" ==  "${EXPECTED_VERSION}" ] ; then
		sed -i -e 's/0.0.'${OLD_VERSION}'/0.0.'${NEW_VERSION}'/g' ../Makefile
		sed -i -e 's/0.0.'${OLD_VERSION}'/0.0.'${NEW_VERSION}'/g' ./update-version.sh
		sed -i -e 's/'${OLD_VERSION}'/'${NEW_VERSION}'/g' ${MANIFESTS_DIR}/service-binding-operator.package.yaml
		sed -i -e 's/'${OLD_VERSION}'/'${NEW_VERSION}'/g' ${MANIFESTS_DIR}/service-binding-operator.v0.0.${OLD_VERSION}.clusterserviceversion.yaml
		mv ${MANIFESTS_DIR}/service-binding-operator.v0.0.${OLD_VERSION}.clusterserviceversion.yaml \
		${MANIFESTS_DIR}/service-binding-operator.v0.0.${NEW_VERSION}.clusterserviceversion.yaml
		sed -i 's/'${OLD_VERSION}'/'${NEW_VERSION}'/g' ./../openshift-ci/Dockerfile.registry.build
		echo -e "\n\033[0;32m \xE2\x9C\x94 Operator version upgraded from ${OLD_VERSION} to ${NEW_VERSION} \033[0m\n"
	else
		echo -e "\n\e[1;35m Enter a suitable version number as a value for the script argument\e[0m \n"
		echo -e "\n\e[1;36m If the previous version is 0.0.n then the new version should be 0.0.n+1\e[0m \n"
	fi
else
	echo -e "\n\e[1;35m Enter a suitable version number as a value for the parameter OPERATOR_VERSION_UPDATE\e[0m \n"
fi
