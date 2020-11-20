from app import App
import re
import requests
import time
import polling2


class NodeJSApp(App):

    pod_name_pattern = "{name}.*$(?<!-build)"

    def __init__(self, name, namespace, nodejs_app_image="quay.io/pmacik/nodejs-rest-http-crud"):
        App.__init__(self, name, namespace, nodejs_app_image, "8080")

    def get_response_from_api(self, endpoint, interval=10, timeout=300):
        resp = polling2.poll(lambda: requests.get(url=f"http://{self.route_url}{endpoint}"),
                             check_success=lambda r: r.status_code in [200], step=interval, timeout=timeout,
                             ignore_exceptions=(requests.exceptions.ConnectionError,))
        return resp.text

    def get_observed_generation(self):
        return self.openshift.get_resource_info_by_jsonpath("deployment", self.name, self.namespace, "{.status.observedGeneration}")

    def get_running_pod_name(self, interval=5, timeout=300):
        start = 0
        while ((start + interval) <= timeout):
            pod_list = self.openshift.get_pod_lst(self.namespace)
            for pod in pod_list:
                if re.fullmatch(self.get_pod_name_pattern(), pod) is not None:
                    if self.openshift.get_pod_status(pod, self.namespace) == "Running":
                        return pod
            time.sleep(interval)
            start += interval
        return None

    def get_redeployed_pod_name(self, old_pod_name, interval=5, timeout=300):
        start = 0
        while ((start + interval) <= timeout):
            pod_list = self.openshift.get_pod_lst(self.namespace)
            for pod in pod_list:
                if pod != old_pod_name and re.fullmatch(self.get_pod_name_pattern(), pod) is not None:
                    if self.openshift.get_pod_status(pod, self.namespace) == "Running":
                        return pod
            time.sleep(interval)
            start += interval
        return None

    def get_pod_name_pattern(self):
        return self.pod_name_pattern.format(name=self.name)

    def is_redeployed(self, old_generation, interval=5, timeout=300):
        start = 0
        while ((start + interval) <= timeout):
            current_generation = self.get_generation()
            pod_list = self.openshift.get_pod_lst(self.namespace)
            for pod in pod_list:
                if (current_generation > old_generation) and (re.fullmatch(self.get_pod_name_pattern(), pod) is not None):
                    if self.openshift.get_pod_status(pod, self.namespace) == "Running":
                        return pod
            time.sleep(interval)
            start += interval
        return None

    def get_generation(self):
        return self.openshift.get_resource_info_by_jsonpath("deployment", self.name, self.namespace, "{.metadata.generation}")

    def get_deployment_with_intermediate_secret(self, intermediate_secret_name):
        return self.openshift.get_deployment_with_intermediate_secret_of_given_pattern(
            intermediate_secret_name, self.name, self.namespace, wait=True, timeout=120)
