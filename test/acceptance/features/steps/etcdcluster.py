import re
from environment import ctx

from command import Command
from openshift import Openshift


class EtcdCluster(object):

    openshift = Openshift()
    cmd = Command()

    name = ""
    namespace = ""

    def __init__(self, name, namespace):
        self.name = name
        self.namespace = namespace
        self.etcd_cluster_template = '''
apiVersion: "etcd.database.coreos.com/v1beta2"
kind: "EtcdCluster"
metadata:
  annotations:
    etcd.database.coreos.com/scope: clusterwide
  name: {etcd_cluster_name}
  namespace: {namespace}
spec:
  repository: quay.io/coreos/etcd
  size: 3
  version: "3.2.13"
'''

    def is_present(self):
        cmd = f'{ctx.cli} get etcdcluster -n {self.namespace}'
        output, exit_code = self.cmd.run(cmd)
        if exit_code != 0:
            print(f"cmd-{cmd} result for getting available knative serving is {output} with the exit code {exit_code}")
            return False
        if self.name in output:
            return True
        return False

    def create(self):
        etcd_cluster_output = self.openshift.apply(self.etcd_cluster_template.format(etcd_cluster_name=self.name, namespace=self.namespace))
        pattern = 'etcdcluster.etcd.database.coreos.com/%s\\screated.*' % self.name
        if re.search(pattern, etcd_cluster_output):
            return True
        print(f"Pattern {pattern} did not match as creating etcd cluster yaml output is {etcd_cluster_output}")
        return False
