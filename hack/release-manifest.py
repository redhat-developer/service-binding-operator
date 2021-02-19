import os
import yaml
import sys
from pathlib import Path

MANIFEST_DIR = str(sys.argv[1])
path = os.getcwd()
release_yaml = path + "/out/release.yaml"
Path(release_yaml).touch()

def release_manifest():
    res = list()
    ns = create_service_binding_operator_ns()
    res.append(ns)
    for filename in os.listdir(MANIFEST_DIR):
        with open(os.path.join(MANIFEST_DIR, filename), 'r') as f:
            data = yaml.load(f)
        if data["kind"] == "ClusterServiceVersion":
            cluster_role = create_cluster_role_yaml(os.path.join(MANIFEST_DIR, filename))
            res.append(cluster_role)
            depl = create_operator_deployment(os.path.join(MANIFEST_DIR, filename))
            res.append(depl)
        elif data["kind"] == "ConfigMap" or data["kind"] == "Service":
            metadata  = data["metadata"]
            metadata.update({"namespace" : 'service-binding-operator'})
            res.append(data)
        else:
            with open(os.path.join(MANIFEST_DIR, filename), "r") as stream:
                res.extend(list(yaml.safe_load_all(stream)))
    clusterrolebinding = create_clusterrolebinding_yaml()
    res.append(clusterrolebinding)
    service_account = create_service_account()
    res.append(service_account)
    with open(release_yaml, "a") as stream:
        yaml.dump_all(
            res,
            stream,
            default_flow_style=False
        )  

def create_cluster_role_yaml(csv):
    role_data = {
    "apiVersion" : 'rbac.authorization.k8s.io/v1',
    "kind" : 'ClusterRole',
    "metadata" : {
        "name" : 'service-binding-operator',
    },
    }
    with open(csv, 'r') as f:
        data = yaml.load(f)
    roles = data["spec"]["install"]["spec"]["clusterPermissions"][0]
    roles.pop('serviceAccountName')
    role_data.update(roles)
    return role_data

def create_clusterrolebinding_yaml():
    # namespace = str(sys.argv[3])
    role_binding = {
    "apiVersion" : 'rbac.authorization.k8s.io/v1',
    "kind" : 'ClusterRoleBinding',
    "metadata" : {
        "name" : 'service-binding-operator',
    },
    "roleRef" : {
        "apiGroup" : 'rbac.authorization.k8s.io',
        "kind" : 'ClusterRole',
        "name" : 'service-binding-operator',
    },
    "subjects":
        [{"kind" : 'ServiceAccount', "name": 'service-binding-operator', "namespace": 'service-binding-operator'},],
}
    return role_binding

def create_operator_deployment(csv):
    depl = {
    "apiVersion" : 'apps/v1',
    "kind" : 'Deployment',
    "metadata" : {
        "name" : 'service-binding-operator',
        "annotations": {"olm.targetNamespaces" : ""},
        "namespace" : 'service-binding-operator',
    },
}
    with open(csv, 'r') as f:
        data = yaml.load(f)
    roles = data["spec"]["install"]["spec"]["deployments"][0]
    del roles['name']
    depl.update(roles)
    return depl

def create_service_account():
    service_account = {
    "apiVersion" : 'v1',
    "kind" : 'ServiceAccount',
    "metadata" : {
        "name" : 'service-binding-operator',
        "namespace" : 'service-binding-operator'
    },
}
    return service_account

def create_service_binding_operator_ns():
    namespace = {
    "apiVersion" : 'v1',
    "kind" : 'Namespace',
    "metadata" : {
        "name" : 'service-binding-operator'
    },
}
    return namespace

## prepares release.yaml manifest to be added with GitHub releases/tags
release_manifest()
