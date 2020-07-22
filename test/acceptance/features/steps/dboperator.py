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
    operator_catalog_source_name = "sample-db-operators"
    operator_catalog_image = "quay.io/redhat-developer/sample-db-operators-olm:v1"
    operator_catalog_channel = "stable"
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

    def install_catalog_source(self):
        install_src_output = self.openshift.create_catalog_source(self.operator_catalog_source_name, self.operator_catalog_image)
        if re.search(r'.*catalogsource.operators.coreos.com/%s\s(unchanged|created)' % self.operator_catalog_source_name, install_src_output) is None:
            print("Failed to create {} catalog source".format(self.operator_catalog_source_name))
            return False
        return self.openshift.wait_for_package_manifest(self.package_name, self.operator_catalog_source_name, self.operator_catalog_channel)

    def install_operator_subscription(self):
        install_sub_output = self.openshift.create_operator_subscription(self.package_name, self.operator_catalog_source_name, self.operator_catalog_channel)
        if re.search(r'.*subscription.operators.coreos.com/%s\s(unchanged|created)' % self.package_name, install_sub_output) is None:
            print("Failed to create {} operator subscription".format(self.package_name))
            return False
        return True

    def get_package_manifest(self):
        cmd = 'oc get packagemanifest %s -o "jsonpath={.metadata.name}"' % self.pkgManifest
        manifest = self.cmd.run_check_for_status(
            cmd, status=self.pkgManifest)
        manifest | should_not.be_equal_to(None)
        manifest | should.equal(self.pkgManifest)
        return manifest
