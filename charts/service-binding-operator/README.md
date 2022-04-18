
# Service Binding Operator Helm Chart

This Helm chart defines the Service Binding Operator. You can install Service Binding Operators using this Helm chart. 

Installing the Service Binding Operator Helm chart creates the following custom resource definitions (CRDs):
- bindablekinds.binding.operators.coreos.com
- servicebindings.binding.operators.coreos.com
- servicebindings.servicebinding.io

The resources required for the Service Binding Operator will also be installed.

## Introduction

The values.yaml file contains the following values that can be customized when installing the chart:

- `image.pullPolicy`
- `image.repository`
- `image.testRepository`
- `keepTestResources`

A user can define values for the image  PullPolicy. 
A user can define values for `image.repository` and `image.testRepository`. If user is not able to pull image from quay.io registry, they can copy the image  to their own container registry.
As part of Helm test we delete the deploymemt,service binding resource and secret used for testing. If a user is interested to view them, then he has to install the chart with keepTestResources set to `true`.


## Helm Chart Installation 

The Helm chart installation involves the following steps:
1. Adding the `service-binding-operator-helm-chart` repository.
2. Installing the Service Binding Operator chart.
3. Running a Helm test.

**Note:** If you are not installing the Service Binding Operator through Operator Lifecycle Manager (OLM), you must install cert-manager on the cluster. Installing the cert-manager automates TLS certificates for Kubernetes and OpenShift workloads. Cert-manager ensures that the certificates are valid and up-to-date, and attempts to renew certificates at a configured time before expiry. You can install cert-manager by running the following command:

kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.0/cert-manager.yaml

### Adding the Helm chart repository
You need to add our helm repository to your local repository. Name the repository as per your convenience.  

```
helm repo add service-binding-operator-helm-chart https://redhat-developer.github.io/service-binding-operator-helm-chart/
```

### Installing the Helm chart
In order to install the chart you need to search the repository, with the following command:

```
helm search repo service-binding-operator-helm-chart
```
```
helm install service-binding-operator-release \
service-binding-operator-helm-chart/service-binding-operator \
--namespace service-binding-operator --create-namespace
```
Remove --namespace and --create-namespace flag if you wish to install the chart on default namespace.

In order to view the resources created on helm test , set the keepTestResources to true. 

```
helm install service-binding-operator-release \
service-binding-operator-helm-chart/service-binding-operator \
--namespace service-binding-operator --create-namespace \
--set keepTestResources=true
```

You can check whether the chart is succesfully installed by running the following command

```
kubectl get pods --namespace service-binding-operator
```

### Helm test

In order to test the chart the user is expected to create a secret (specify the namespace if applicable), named my-k-config from his kubeconfig .

**NOTE**:
In case you are installing the chart on AWS eks cluster then you need to modify the aws-auth configmap.
```
kubectl edit -n kube-system cm/aws-auth 
```
Please add -system:masters to mapRoles and save.
After editing the config map you need to update the eks kubeconfig
```
aws eks update-kubeconfig --name <cluster-name>
```
Then Continue with the following steps.

```
kubectl create secret generic my-k-config --from-file=kubeconfig=<PATH TO YOUR KUBECONFIG> -namespace service-binding-operator
```

Run the Helm test (specify the namespace if applicable) using :

```
helm test service-binding-operator-release --namespace service-binding-operator
```

Please ensure to delete the secret (specify the namespace if applicable) created :
```
kubectl delete secret my-k-config --namespace service-binding-operator
```

## Additional Help
Please reach out to us for any additional queries by creating an issue on https://github.com/redhat-developer/service-binding-operator/issues.
