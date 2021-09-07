#!/usr/bin/env bash

[ -z "$MINIKUBE_HOME" ] && MINIKUBE_HOME=$HOME/.minikube

MINIKUBE_TOKEN_PATH=/etc/ca-certificates/tokens.csv

HOST_TOKEN_PATH=$MINIKUBE_HOME/files$MINIKUBE_TOKEN_PATH

USER=acceptance-tests-dev
USER_TOKEN=topsecret-token

mkdir -p $(dirname $HOST_TOKEN_PATH)

cat > $HOST_TOKEN_PATH <<EOF
$USER_TOKEN,$USER,${USER}1
EOF

minikube start --addons=registry --insecure-registry=0.0.0.0/0 --extra-config="apiserver.token-auth-file=$MINIKUBE_TOKEN_PATH" "$@"

kubectl config set-credentials $USER --token=$USER_TOKEN
