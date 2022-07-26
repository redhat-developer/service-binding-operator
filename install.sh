#!/bin/bash -e

OLM_VERSION=0.21.2
OPERATOR_CHANNEL=${OPERATOR_CHANNEL:-beta}
OPERATOR_PACKAGE=${OPERATOR_PACKAGE:-service-binding-operator}
DOCKER_CFG=${DOCKER_CFG:-$HOME/.docker/config.json}
CONTAINER_RUNTIME=${CONTAINER_RUNTIME:-docker}

kubectl get crd clusterserviceversions.operators.coreos.com || \
  curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v${OLM_VERSION}/install.sh | bash -s v${OLM_VERSION}

CATSRC_NAMESPACE="${CATSRC_NAMESPACE:-$(kubectl get catalogsources.operators.coreos.com --all-namespaces -o jsonpath='{.items[0].metadata.namespace}' --ignore-not-found=true)}"

OPNS="$(kubectl get namespaces -o name | grep -E '.*/(.+-)?operators$' || test $? = 1)"
OPERATOR_NAMESPACE=${OPNS#"namespace/"}

if [ -n "$OPERATOR_INDEX_IMAGE" ]; then
  if [ -r "$DOCKER_CFG" ]; then
    kubectl create secret generic sbo-operators-secrets -n $CATSRC_NAMESPACE --from-file=.dockerconfigjson=$DOCKER_CFG --type=kubernetes.io/dockerconfigjson --dry-run=client -o json | kubectl apply -f -
    kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "sbo-operators-secrets"}]}' -n=$CATSRC_NAMESPACE
  fi

  OPERATOR_REGISTRY_REF=$(echo $OPERATOR_INDEX_IMAGE | cut -f 1 -d '/')

  if [ -z "$SKIP_REGISTRY_LOGIN" ]; then
    if [ -z "$OPERATOR_REGISTRY_USERNAME" ]; then
      ${CONTAINER_RUNTIME} login $OPERATOR_REGISTRY_REF
    else
      ${CONTAINER_RUNTIME} login -u "$OPERATOR_REGISTRY_USERNAME" --pasword-stdin $OPERATOR_REGISTRY_REF <<<$OPERATOR_REGISTRY_PASSWORD
    fi
  fi

  CATSRC_ALREADY_FOUND=$(kubectl get catsrc -n $CATSRC_NAMESPACE -o json | jq -rc '.items[] | select(.spec.image=="'$OPERATOR_INDEX_IMAGE'").metadata.name')
  if [ -n "$CATSRC_ALREADY_FOUND" ]; then
    echo "Catalog source with a given index image already found in namespace '$CATSRC_NAMESPACE': '$CATSRC_ALREADY_FOUND', using it for subscription."
    CATSRC_NAME=$CATSRC_ALREADY_FOUND
  else
    CATSRC_NAME=${CATSRC_NAME:-catsrc-$(echo -n "$OPERATOR_INDEX_IMAGE" | sed -e 's,[/:\.],-,g')}
    #Apply CatalogSource for obtaining catalog of SBO operators
    kubectl apply -f - << EOD
---
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: $CATSRC_NAME
  namespace: $CATSRC_NAMESPACE
spec:
  displayName: SBO Operators
  image: $OPERATOR_INDEX_IMAGE
  sourceType: grpc
  publisher: Red Hat
  updateStrategy:
    registryPoll:
      interval: 10m0s
EOD
  fi
else
  CATSRC_CANDIDATES=(
    operatorhubio-catalog
    redhat-operators
  )
  for i in ${CATSRC_CANDIDATES[@]} ; do
    if [ "$(kubectl get catsrc $i -n $CATSRC_NAMESPACE -o name --ignore-not-found | wc -l)" -gt 0 ]; then
      CATSRC_NAME=${CATSRC_NAME:-$i}
      break
    fi
  done
fi

#Apply subscription for subscribing to the requested channel of the service binding operator
SUB_ALREADY_FOUND=$(kubectl get sub -n $OPERATOR_NAMESPACE -o json --ignore-not-found | jq -rc '.items[] | select(.spec.source == "'$CATSRC_NAME'" and .spec.sourceNamespace == "'$CATSRC_NAMESPACE'" and .spec.channel == "'$OPERATOR_CHANNEL'" and .spec.name == "'$OPERATOR_PACKAGE'").metadata.name')
if [ -n "$SUB_ALREADY_FOUND" ]; then
  echo "Subscription to the given channel '$OPERATOR_CHANNEL' of the given Catalog Source '$CATSRC_NAMESPACE/$CATSRC_NAME)' already found: '$SUB_ALREADY_FOUND'."
  echo "Skipping creation of the subscription to avoid duplicities."
else
  kubectl apply -f - << EOD
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: $OPERATOR_PACKAGE
  namespace: $OPERATOR_NAMESPACE
spec:
  channel: $OPERATOR_CHANNEL
  installPlanApproval: Automatic
  name: $OPERATOR_PACKAGE
  source: $CATSRC_NAME
  sourceNamespace: $CATSRC_NAMESPACE
EOD
fi

#Wait for the operator to get up and running
retries=50
until [[ $retries == 0 ]]; do
  kubectl get deployment/service-binding-operator -n $OPERATOR_NAMESPACE >/dev/null 2>&1 && break
  echo "Waiting for service-binding-operator to be created in $OPERATOR_NAMESPACE namespace"
  sleep 5
  retries=$(($retries - 1))
done
kubectl rollout status -w deployment/service-binding-operator -n $OPERATOR_NAMESPACE
