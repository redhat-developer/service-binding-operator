import re
import polling2

from command import Command
from openshift import Openshift


class ServerlessOperator():

    openshift = Openshift()
    cmd = Command()

    pod_name_pattern = "{name}.*"

    name = ""
    namespace = ""
    operator_catalog_source_name = "redhat-operators"
    operator_catalog_channel = "4.5"

    def __init__(self, name="serverless-operator", namespace="openshift-operators"):
        self.name = name
        self.namespace = namespace

    def is_running(self, wait=False):
        currentCSV = self.openshift.get_current_csv(self.name, self.operator_catalog_source_name, self.operator_catalog_channel)
        if wait:
            polling2.poll(lambda: self.openshift.search_resource_in_namespace("csvs", currentCSV,
                                                                              self.namespace), check_success=lambda v: v is not None, step=1, timeout=100)
        else:
            if self.openshift.search_resource_in_namespace("csvs", currentCSV, self.namespace) is None:
                return False

        expectedDeployments = self.openshift.get_resource_info_by_jsonpath(
            "csv", currentCSV, self.namespace, "{.spec.install.spec.deployments[*].name}").split()
        found_pod_names = []
        for deployment in expectedDeployments:
            if wait:
                found_pod_name = self.openshift.wait_for_pod(self.pod_name_pattern.format(name=deployment), self.namespace)
            else:
                found_pod_name = self.openshift.search_pod_in_namespace(self.pod_name_pattern.format(name=deployment), self.namespace)
            if found_pod_name is not None:
                operator_pod_status = self.openshift.check_pod_status(found_pod_name, self.namespace)
                print("The pod {} is running: {}".format(found_pod_name, operator_pod_status))
                found_pod_names.append(found_pod_name)
        if len(found_pod_names) == len(expectedDeployments):
            return True
        else:
            print(f"Not all pods from expected deployments [{expectedDeployments}] are running. Only following pods are: [{found_pod_names}]")
        return False

    def install_operator_subscription(self):
        install_sub_output = self.openshift.create_operator_subscription(self.name, self.operator_catalog_source_name, self.operator_catalog_channel)
        if re.search(r'.*subscription.operators.coreos.com/%s\s(unchanged|created)' % self.name, install_sub_output) is not None:
            return True
        return False
