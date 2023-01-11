import re

from command import Command
from subscription_install_mode import InstallMode
from openshift import Openshift


class Operator(object):

    openshift = Openshift()
    cmd = Command()

    def __init__(self,
                 name="",
                 operator_namespace="",
                 package_name="",
                 operator_catalog_source_name="",
                 operator_catalog_image="",
                 operator_catalog_channel="",
                 operator_catalog_namespace="",
                 operator_subscription_csv_version=None,
                 pod_name_pattern="{name}.*"):
        self.name = name
        self.operator_namespace = operator_namespace if operator_namespace != "" else self.openshift.operators_namespace
        self.package_name = package_name
        self.operator_catalog_source_name = operator_catalog_source_name
        self.operator_catalog_image = operator_catalog_image
        self.operator_catalog_channel = operator_catalog_channel
        self.operator_catalog_namespace = operator_catalog_namespace if operator_catalog_namespace != "" else self.openshift.olm_namespace
        self.operator_subscription_csv_version = operator_subscription_csv_version
        self.pod_name_pattern = pod_name_pattern

    def is_running(self, wait=False):
        if wait:
            pod_name = self.openshift.wait_for_pod(self.pod_name_pattern.format(name=self.name), self.operator_namespace)
        else:
            pod_name = self.openshift.search_pod_in_namespace(self.pod_name_pattern.format(name=self.name), self.operator_namespace)
        if pod_name is not None:
            operator_pod_status = self.openshift.check_pod_status(pod_name, self.operator_namespace)
            print("The pod {} is running: {}".format(self.name, operator_pod_status))
            return operator_pod_status
        else:
            return False

    def install_catalog_source(self):
        if self.operator_catalog_image != "":
            install_src_output = self.openshift.create_catalog_source(
                self.operator_catalog_source_name, self.operator_catalog_image, self.operator_catalog_namespace)
            if re.search(r'.*catalogsource.operators.coreos.com/%s\s(unchanged|created)' % self.operator_catalog_source_name, install_src_output) is None:
                print("Failed to create {} catalog source".format(self.operator_catalog_source_name))
                return False
        return self.openshift.wait_for_package_manifest(self.package_name, self.operator_catalog_source_name, self.operator_catalog_channel)

    def csv_version_resolved(self, csv_version=None):
        if csv_version is None:
            if self.operator_subscription_csv_version is None:
                return self.openshift.get_current_csv(self.package_name, self.operator_catalog_source_name, self.operator_catalog_channel)
            else:
                return self.operator_subscription_csv_version
        else:
            return csv_version

    def install_operator_subscription(self, csv_version=None, install_mode=InstallMode.Automatic):
        csv_version_resolved = self.csv_version_resolved(csv_version)
        install_sub_output = self.openshift.create_operator_subscription(
            self.package_name, self.operator_catalog_source_name, self.operator_catalog_channel, self.operator_catalog_namespace,
            csv_version_resolved, install_mode)
        if re.search(r'.*subscription.operators.coreos.com/%s\s(unchanged|created)' % self.package_name, install_sub_output) is None:
            print("Failed to create {} operator subscription".format(self.package_name))
            return False
        self.openshift.approve_operator_subscription(self.package_name)
        return True
