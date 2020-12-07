#!/bin/bash -e

OLM_VERSION=0.17.0
OPERATOR_INDEX_IMAGE=${OPERATOR_INDEX_IMAGE:-registry.redhat.io/redhat/community-operator-index:latest}
OPERATOR_REGISTRY_REF=$(echo $OPERATOR_INDEX_IMAGE | cut -f 1 -d '/')
OPERATOR_CHANNEL=${OPERATOR_CHANNEL:-beta}
DOCKER_CFG=${DOCKER_CFG:-$HOME/.docker/config.json}
CONTAINER_RUNTIME=${CONTAINER_RUNTIME:-docker}

#The SBO can also be installed in a vanilla kubernetes cluster. A prerequisite for this would be to add credentials for the registry.redhat.io in the cluster. Steps to be followed to achieve the same are listed below:
#1) Follow [Red Hat Container Registry Authentication Steps](https://access.redhat.com/RegistryAuthentication)
#2) Verify that your credentials are correct using docker login -u <your_username> -p <your_passwd> registry.redhat.io

if [ -z "$SKIP_REGISTRY_LOGIN" ]; then
  if [ -z "$OPERATOR_REGISTRY_USERNAME" ]; then
    ${CONTAINER_RUNTIME} login $OPERATOR_REGISTRY_REF
  else
    ${CONTAINER_RUNTIME} login -u "$OPERATOR_REGISTRY_USERNAME" --pasword-stdin $OPERATOR_REGISTRY_REF <<<$OPERATOR_REGISTRY_PASSWORD
  fi
fi

#3) Start a kubernetes cluster. Any Kubernetes cluster can be used as well. The script assumes that a Kubernetes is up and running and the user has logged into it.

#4) Enable [OLM](https://github.com/operator-framework/operator-lifecycle-manager) in the cluster by running the following command
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v${OLM_VERSION}/install.sh | bash -s v${OLM_VERSION}
#On minikube you can alternatively install OLM by running `minikube addons enable olm`

#5) Create a new image pull secret out of your local .docker/config.json file
if [ -r "$DOCKER_CFG" ]; then
  kubectl create secret generic sbo-operators-secrets -n olm --from-file=.dockerconfigjson=$DOCKER_CFG --type=kubernetes.io/dockerconfigjson

  #6) Add that pull secret to the default account in olm namespace
  kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "sbo-operators-secrets"}]}' -n=olm
fi

#7) Install the operator by running the following commands
#Apply CatalogSource for obtaining catalog of SBO operators
kubectl apply -f - << EOD
---
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: sbo-operators
  namespace: olm
spec:
  displayName: SBO Operators
  image: $OPERATOR_INDEX_IMAGE
  sourceType: grpc
  publisher: Red Hat
  updateStrategy:
    registryPoll:
      interval: 10m0s
EOD
#Apply subscription for subscribing to the requested channel of the service binding operator
kubectl apply -f - << EOD
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: service-binding-operator
  namespace: operators
spec:
  channel: $OPERATOR_CHANNEL
  installPlanApproval: Automatic
  name: service-binding-operator
  source: sbo-operators
  sourceNamespace: olm
EOD
#This Operator will be installed in the "operators" namespace and will be usable from all namespaces in the cluster.

retries=50
until [[ $retries == 0 ]]; do
  kubectl get deployment/service-binding-operator -n operators >/dev/null 2>&1 && break
  echo "Waiting for service-binding-operator to be created"
  sleep 5
  retries=$(($retries - 1))
done
kubectl rollout status -w deployment/service-binding-operator -n operators