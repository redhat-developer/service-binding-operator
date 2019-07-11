#!/bin/bash

function check_crds () {
	local crd_name="$1"

	for i in  1..120 ; do
		if kubectl get crds $crd_name -o wide ; then
			echo "CRD is found: ${crd_name}"
			return 0
		fi

		sleep 3
	done

	echo "CRD doesn't exist: ${crd_name}"
	return 1
}

if ! check_crds servicebindingrequests.apps.openshift.io ; then
	exit 1
fi
