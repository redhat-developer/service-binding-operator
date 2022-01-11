#!/bin/bash -e

NAMESPACE=${NAMESPACE:-my-postgresql}

sed -e 's,quay.io/service-binding/spring-petclinic:latest,'${PETCLINIC_APP_IMAGE}',g' app-deployment.yaml | \
kubectl apply -f - -n ${NAMESPACE} --wait

echo ""
echo "Petclinic application is running and available at $(minikube service spring-petclinic -n ${NAMESPACE} --url)"
echo ""
