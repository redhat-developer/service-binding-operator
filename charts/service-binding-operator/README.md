
# Service Binding Operator Helm Chart

This Helm chart defines the Service Binding Operator. You can install Service Binding Operators using this Helm chart. 

Installing the Service Binding Operator Helm chart creates the following custom resource definitions (CRDs):
- bindablekinds.binding.operators.coreos.com
- servicebindings.binding.operators.coreos.com
- servicebindings.servicebinding.io
- clusterworkloadresourcemappings.servicebinding.io

The resources required for the Service Binding Operator are also installed.

## Introduction

The `values.yaml` file contains the following values that can be customized when installing the chart:

- `image.pullPolicy`
- `image.repository`
- `image.testRepository`
- `keepTestResources`

You can define values for the `image.pullPolicy`, `image.repository`, and `image.testRepository` values. If you are not able to pull an image from the quay.io registry, then copy the image to your own container registry.
As part of the Helm test, objects such as deployment, service binding resources, and secrets used for testing are deleted. To view them, you can install the chart with the `keepTestResources` flag value set to `true`.


## Helm Chart Installation 

The Helm chart installation involves the following steps:

1. Adding the `service-binding-operator-helm-chart` repository.
2. Installing the Service Binding Operator Helm chart.
3. Running a Helm test.

**Note:** If you are not installing the Service Binding Operator through Operator Lifecycle Manager (OLM), you must install cert-manager on the cluster. Installing the cert-manager automates TLS certificates for Kubernetes and OpenShift workloads. Cert-manager ensures that the certificates are valid and up-to-date, and attempts to renew certificates at a configured time before expiry. You can install cert-manager by running the following command:

```
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.0/cert-manager.yaml
```

### Adding the Helm chart repository

1. Add the `service-binding-operator-helm-chart` repository to your local repository and name the repository as per your convenience: 

```
helm repo add service-binding-operator-helm-chart https://redhat-developer.github.io/service-binding-operator-helm-chart/
```

### Installing the Helm chart

1. Search the repository for the Service Binding Operator Helm chart:

```
helm search repo service-binding-operator-helm-chart
```

2. Create a Helm chart release and specify a namespace with the `--create-namespace` flag:

```
helm install service-binding-operator-release \
service-binding-operator-helm-chart/service-binding-operator \
--namespace service-binding-operator --create-namespace
```

3. Optional: If you wish to install the chart on the default namespace, remove the `--namespace` and `--create-namespace` flags.

4. Optional: To view the resources created for testing, set the `keepTestResources` flag value to `true`: 

```
helm install service-binding-operator-release \
service-binding-operator-helm-chart/service-binding-operator \
--namespace service-binding-operator --create-namespace \
--set keepTestResources=true
```

5. Verify that the chart is successfully installed:

```
kubectl get pods --namespace service-binding-operator
```

### Running a Helm test

1. Optional: If you are installing the chart on the Amazon Elastic Kubernetes Service (Amazon EKS) cluster, then perform the following steps:

    1. Modify the `aws-auth` config map:

    ```
    kubectl edit -n kube-system cm/aws-auth 
    ```

    2. Add `-system:masters` to mapRoles and save.

    3. After editing the config map, update the EKS `kubeConfig` file:

    ```
    aws eks update-kubeconfig --name  <cluster-name>
    ```

2. Create a `my-k-config` secret and specify the namespace if applicable from your `kubeconfig` file.

```
kubectl create secret generic my-k-config --from-file=kubeconfig=<PATH TO YOUR KUBECONFIG> --namespace service-binding-operator
```

3. Run the Helm test and specify the namespace if applicable:

```
helm test service-binding-operator-release --namespace service-binding-operator
```

The `Succeeded` phase from the output indicates that the Helm test has run successfully.

4. Verify that the Helm test has run successfully:

```
kubectl get pods  --namespace service-binding-operator
```

5. As a safety measure, delete the secret that you created and specify the namespace if applicable:

```
kubectl delete secret my-k-config --namespace service-binding-operator
```

## Additional Help
Please reach out to us for any additional queries by creating an issue at https://github.com/redhat-developer/service-binding-operator/issues.
