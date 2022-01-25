#!/usr/bin/env bash

echo "Downloading requested CLI"

if [ "$OPERATOR_SDK" == true ]; then
    echo "Downloading operator-sdk..."
    curl -Lo operator-sdk https://github.com/operator-framework/operator-sdk/releases/download/v${SDK_VERSION}/operator-sdk_linux_amd64
    chmod +x operator-sdk
    mv -v operator-sdk $GITHUB_WORKSPACE/bin/
fi

if [ "$KUBECTL" == true ] || [ "$START_MINIKUBE" == true ]; then
    echo "Downloading kubectl..."
    curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v${K8S_VERSION}/bin/linux/amd64/kubectl
    chmod +x kubectl
    mv -v kubectl $GITHUB_WORKSPACE/bin/
fi

if [ "$MINIKUBE" == true ] || [ "$START_MINIKUBE" == true ]; then
    echo "Downloading minikube..."
    curl -Lo minikube https://storage.googleapis.com/minikube/releases/v${MINIKUBE_VERSION}/minikube-linux-amd64
    chmod +x minikube
    mv -v minikube $GITHUB_WORKSPACE/bin/
fi

echo "All requested CLI downloaded!"
