from openshift import Openshift
from command import Command
import re
import requests
import time


class NodeJSApp(object):

    nodesj_app_image = "quay.io/pmacik/nodejs-rest-http-crud"
    api_end_point = 'http://{route_url}/api/status/dbNameCM'
    openshift = Openshift()

    pod_name_pattern = "{name}.*$(?<!-build)"

    name = ""
    namespace = ""

    def __init__(self, name, namespace):
        self.cmd = Command()
        self.name = name
        self.namespace = namespace

    def is_running(self, wait=False):
        deployment_flag = False

        if wait:
            pod_name = self.openshift.wait_for_pod(self.get_pod_name_pattern(), self.namespace, timeout=180)
        else:
            pod_name = self.openshift.search_pod_in_namespace(self.get_pod_name_pattern(), self.namespace)

        if pod_name is not None:
            application_pod_status = self.openshift.check_pod_status(pod_name, self.namespace, wait_for_status="Running")
            print("The pod {} is running: {}".format(pod_name, application_pod_status))

            deployment = self.openshift.search_resource_in_namespace("deployments", f"{self.name}.*", self.namespace)
            if deployment is not None:
                print("deployment is {}".format(deployment))
                deployment_flag = True

            if application_pod_status and deployment_flag:
                return True
            else:
                return False
        else:
            return False

    def install(self):
        create_new_app_output, exit_code = self.cmd.run(f"oc new-app --docker-image={self.nodesj_app_image} --name={self.name} -n {self.namespace}")
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned when attempting to create a new app: {create_new_app_output}"
        assert re.search(f'imagestream.image.openshift.io.*{self.name}.*created',
                         create_new_app_output) is not None, f"Unable to create imagestream: {create_new_app_output}"
        assert re.search(f'deployment.apps.*{self.name}.*created',
                         create_new_app_output) is not None, f"Unable to create deployment: {create_new_app_output}"
        assert re.search(f'service.*{self.name}.*created',
                         create_new_app_output) is not None, f"Unable to create service: {create_new_app_output}"
        assert self.openshift.expose_service_route(self.name, self.namespace) is not None, "Unable to expose service route"
        return self.is_running(wait=True)

    def get_db_name_from_api(self, interval=5, timeout=60):
        route_url = self.openshift.get_route_host(self.name, self.namespace)
        if route_url is None:
            return None
        start = 0
        while ((start + interval) <= timeout):
            db_name = requests.get(url=self.api_end_point.format(route_url=route_url))
            if db_name.status_code == 200:
                return db_name.text
            time.sleep(interval)
            start += interval
        return None

    def get_observed_generation(self):
        return self.openshift.get_resource_info_by_jsonpath("deployment", self.name, self.namespace, "{.status.observedGeneration}")

    def get_running_pod_name(self, interval=5, timeout=60):
        start = 0
        while ((start + interval) <= timeout):
            pod_list = self.openshift.get_pod_lst(self.namespace)
            for pod in pod_list.split(" "):
                if re.fullmatch(self.get_pod_name_pattern(), pod):
                    if self.openshift.get_pod_status(pod, self.namespace) == "Running":
                        return pod
            time.sleep(interval)
            start += interval
        return None

    def get_redeployed_pod_name(self, old_pod_name, interval=5, timeout=60):
        start = 0
        while ((start + interval) <= timeout):
            pod_list = self.openshift.get_pod_lst(self.namespace)
            for pod in pod_list.split(" "):
                if pod != old_pod_name and re.fullmatch(self.get_pod_name_pattern(), pod):
                    if self.openshift.get_pod_status(pod, self.namespace) == "Running":
                        return pod
            time.sleep(interval)
            start += interval
        return None

    def get_pod_name_pattern(self):
        return self.pod_name_pattern.format(name=self.name)
