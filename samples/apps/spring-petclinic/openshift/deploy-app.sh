#!/bin/bash -e

NAMESPACE=${NAMESPACE:-my-postgresql}

sed -e 's,quay.io/service-binding/spring-petclinic:latest,'${PETCLINIC_APP_IMAGE}',g' app-deployment.yaml | \
oc apply -f - -n ${NAMESPACE} --wait
oc expose service/spring-petclinic -n ${NAMESPACE} || true

echo ""
echo "Petclinic application is running and available at http://$(oc get route spring-petclinic -n ${NAMESPACE} -o jsonpath={.spec.host})"
echo ""
