## Binding an Imported app to an In-cluster Operator Managed Etcd Database.

1. Install Etcd operator using operator hub,
   follow https://operatorhub.io/operator/etcd
   you can specify the following while installing etcd operator:
   catalog source = "community-operators"
   channel = "clusterwide-alpha"

2. Create an Etcd cluster.
 ```yaml
 apiVersion: "etcd.database.coreos.com/v1beta2"
 kind: "EtcdCluster"
 metadata:
  annotations:
   etcd.database.coreos.com/scope: clusterwide
  name: "etcd-cluster-example"
 spec:
  repository: quay.io/coreos/etcd
  size: 3
  version: "3.2.13"
 ```

3. Deploy application which uses Etcd client lib using OpenShift 4 devconsole.
Example application : https://github.com/akashshinde/node-todo.git
![](https://i.imgur.com/WGQZ1nj.png)

Note: This example assumes the app is deployed using a K8s Deployment. If using an OCP DeploymentConfig change the group and resource in the applicationSelector to:

```
    group: apps.openshift.io
    version: v1
    resource: deploymentconfigs
```

4. Add the labels to the deployment or deployment config (Optional)

```
oc label deployment node-todo-git 'connects-to=etcd' 'environment=demo'
```

5. Create SBR.
```yaml
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
  name: binding-request
spec:
  applicationSelector:
    group: apps.openshift.io
    version: v1
    resource: deploymentconfigs
    resourceRef: node-todo-git
  backingServiceSelector:
    group: etcd.database.coreos.com
    version: v1beta2
    kind: EtcdCluster
    resourceRef: etcd-cluster-example
  detectBindingResources: true
```

6. Application should be binded to the Etcd database automatically.
![](https://i.imgur.com/JjORDrJ.png)
