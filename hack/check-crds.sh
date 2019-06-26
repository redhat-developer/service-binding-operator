#!/bin/bash
function check_crds() {
	local crd_name=$1
	for i in {1..120}
	do
		oc get crds | grep $crd_name
		if [ $? == 0 ]
		then
			echo "CRD exists: " $crd_name
			return 0
		fi
		sleep 10s
	done
	echo "CRD doesn't exist: " $crd_name
	exit 1
}

check_crds servicebindingrequests.apps.openshift.io
