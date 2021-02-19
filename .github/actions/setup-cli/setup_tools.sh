#!/usr/bin/env bash

curl -Lo operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v${SDK_VERSION}/operator-sdk_linux_amd64
chmod +x operator-sdk
mv -v operator-sdk $GITHUB_WORKSPACE/bin/

curl -Lo opm https://github.com/operator-framework/operator-registry/releases/download/v${OPM_VERSION}/linux-amd64-opm
chmod +x opm
mv -v opm $GITHUB_WORKSPACE/bin/

curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v${K8S_VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
mv -v kubectl $GITHUB_WORKSPACE/bin/

curl -Lo minikube https://storage.googleapis.com/minikube/releases/v${MINIKUBE_VERSION}/minikube-linux-amd64
chmod +x minikube
mv -v minikube $GITHUB_WORKSPACE/bin/
