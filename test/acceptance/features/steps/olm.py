import re

from command import Command
from subscription_install_mode import InstallMode
from openshift import Openshift


class Operator(object):

    openshift = Openshift()
    cmd = Command()
    pod_name_pattern = "{name}.*"

    name = ""
    operator_catalog_source_name = ""
    operator_catalog_image = ""
    operator_catalog_channel = ""
    operator_subscription_csv_version = None
    package_name = ""

    def is_running(self, wait=False):
        if wait:
            pod_name = self.openshift.wait_for_pod(self.pod_name_pattern.format(name=self.name), self.openshift.operators_namespace)
        else:
            pod_name = self.openshift.search_pod_in_namespace(self.pod_name_pattern.format(name=self.name), self.openshift.operators_namespace)
        if pod_name is not None:
            operator_pod_status = self.openshift.check_pod_status(pod_name, self.openshift.operators_namespace)
            print("The pod {} is running: {}".format(self.name, operator_pod_status))
            return operator_pod_status
        else:
            return False

    def install_catalog_source(self):
        if self.operator_catalog_image != "":
            install_src_output = self.openshift.create_catalog_source(self.operator_catalog_source_name, self.operator_catalog_image)
            if re.search(r'.*catalogsource.operators.coreos.com/%s\s(unchanged|created)' % self.operator_catalog_source_name, install_src_output) is None:
                print("Failed to create {} catalog source".format(self.operator_catalog_source_name))
                return False
        return self.openshift.wait_for_package_manifest(self.package_name, self.operator_catalog_source_name, self.operator_catalog_channel)

    def install_operator_subscription(self, csv_version=None, install_mode=InstallMode.Automatic):
        install_sub_output = self.openshift.create_operator_subscription(
            self.package_name, self.operator_catalog_source_name, self.operator_catalog_channel,
            self.operator_subscription_csv_version if csv_version is None else csv_version, install_mode)
        if re.search(r'.*subscription.operators.coreos.com/%s\s(unchanged|created)' % self.package_name, install_sub_output) is None:
            print("Failed to create {} operator subscription".format(self.package_name))
            return False
        self.openshift.approve_operator_subscription(self.package_name)
        return True
