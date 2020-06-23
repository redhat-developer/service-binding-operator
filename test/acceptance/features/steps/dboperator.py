import re
from pyshould import should, should_not

from command import Command
from openshift import Openshift


class DbOperator():

    openshift = Openshift()
    cmd = Command()

    pod_name_pattern = "{name}.*"

    name = ""
    namespace = ""
    operator_source_name = "db-operators"
    operator_registry_namespace = "pmacik"
    operator_registry_channel = "stable"
    package_name = "db-operators"

    def __init__(self, name="postgresql-operator", namespace="openshift-operators"):
        self.name = name
        self.namespace = namespace

    def is_running(self, wait=False):
        if wait:
            pod_name = self.openshift.wait_for_pod(self.pod_name_pattern.format(name=self.name), self.namespace)
        else:
            pod_name = self.openshift.search_pod_in_namespace(self.pod_name_pattern.format(name=self.name), self.namespace)
        if pod_name is not None:
            operator_pod_status = self.openshift.check_pod_status(pod_name, self.namespace)
            print("The pod {} is running: {}".format(self.name, operator_pod_status))
            return operator_pod_status
        else:
            return False

    def install_operator_source(self):
        install_src_output = self.openshift.create_operator_source(self.operator_source_name, self.operator_registry_namespace)
        if not re.search(r'.*operatorsource.operators.coreos.com/%s\s(unchanged|created)' % self.operator_source_name, install_src_output):
            print("Failed to create {} operator source".format(self.operator_source_name))
            return False
        return self.openshift.wait_for_package_manifest(self.package_name, self.operator_source_name, self.operator_registry_channel)

    def install_operator_subscription(self):
        install_sub_output = self.openshift.create_operator_subscription(self.package_name, self.operator_source_name, self.operator_registry_channel)
        return re.search(r'.*subscription.operators.coreos.com/%s\s(unchanged|created)' % self.operator_source_name, install_sub_output)

    def get_package_manifest(self):
        cmd = 'oc get packagemanifest %s -o "jsonpath={.metadata.name}"' % self.pkgManifest
        manifest = self.cmd.run_check_for_status(
            cmd, status=self.pkgManifest)
        manifest | should_not.be_equal_to(None)
        manifest | should.equal(self.pkgManifest)
        return manifest
