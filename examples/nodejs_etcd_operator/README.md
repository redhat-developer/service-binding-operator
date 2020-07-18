## Binding an Imported app to an In-cluster Operator Managed Etcd Database.

1. Install Etcd operator using operator hub,
   follow https://operatorhub.io/operator/etcd
2. Create an Etcd cluster.
 ```yaml
 apiVersion: "etcd.database.coreos.com/v1beta2"
 kind: "EtcdCluster"
 metadata:
  name: "etcd-cluster"
 spec:
  size: 5
  version: "3.2.13"
 ```
3. Deploy application which uses Etcd client lib using openshift 4 devconsole.
Test application : https://github.com/akashshinde/node-todo.git
![](https://i.imgur.com/WGQZ1nj.png)

Note: This example assumes the app is deployed using a K8s Deployment if using a OCP DeploymentConfig change the group and resource in the applicationSelector to

```
    group: apps.openshift.io
    version: v1
    resource: deploymentconfigs
```

4. Add the labels to the deployment or deployment config

```
oc label deployment node-todo-git 'connects-to=etcd' 'environment=demo'
```

5. Create SBR.
```yaml
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
  name: node-binding-request
spec:
  applicationSelector:
    labelSelector:
      matchLabels:
        connects-to: etcd
        environment: demo
    group: apps
    resource: deployments
    version: v1
    resourceRef: ""
  backingServiceSelector:
    group: etcd.database.coreos.com
    version: v1beta2
    kind: EtcdCluster
    resourceRef: etcd-cluster
  mountPathPrefix: “”
  customEnvVar: []
  detectBindingResources: true
```
6. Application should be binded to the Etcd database automatically.
![](https://i.imgur.com/JjORDrJ.png)


