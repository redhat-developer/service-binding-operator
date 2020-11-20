from command import Command
import re
import time
import requests
from openshift import Openshift


class QuarkusApplication(object):

    cmd = Command()

    image_name_with_tag = "quay.io/pmacik/using-spring-data-jqa-quarkus:latest"
    openshift = Openshift()

    name = ""
    namespace = ""
    deployment_name_pattern = "{name}-\\w+-deployment"

    def __init__(self, name, namespace):
        self.name = name
        self.namespace = namespace

    def install(self):
        knative_service_output = self.openshift.create_knative_service(self.name, self.namespace, self.image_name_with_tag)
        output = re.search(r'.*service.serving.knative.dev/%s\s(created|configured|unchanged)' % self.name, knative_service_output)
        assert output is not None, f"Knative serving is not created as the result is {knative_service_output}"
        return True

    def get_pod_name_running(self, pod_name_pattern, wait=False):
        if wait:
            pod_name = self.openshift.wait_for_pod(self.format_pattern(pod_name_pattern), self.namespace, timeout=500)
        else:
            pod_name = self.openshift.search_pod_in_namespace(self.format_pattern(pod_name_pattern), self.namespace)
        return pod_name

    def is_imported(self, wait=False, interval=5, timeout=600):
        deployment_name = self.openshift.get_deployment_name_in_namespace(
            self.format_pattern(self.deployment_name_pattern), self.namespace, wait=wait, timeout=timeout)
        if deployment_name is None:
            return False
        else:
            deployment_replicas = self.openshift.get_resource_info_by_jsonpath("deployment", deployment_name, self.namespace, "{.status.replicas}")
            assert deployment_replicas.isnumeric(
            ), f"Number of replicas of deployment '{deployment_name}' should be a numerical value, but is actually: '{deployment_replicas}"
            assert int(str(deployment_replicas)) > 0, "Number of replicas of deployment '{deployment_name}' " + \
                "should be greater than 0, but is actually: '{deployment_replicas}'."
            return True

    def get_response_from_api(self, endpoint, wait=False, interval=5, timeout=300):
        route_url = self.openshift.get_knative_route_host(self.name, self.namespace)
        if route_url is None:
            return None
        url = f"{route_url}{endpoint}"
        if wait:
            start = 0
            while ((start + interval) <= timeout):
                db_name = requests.get(url)
                if db_name.status_code == 200:
                    return db_name.text
                time.sleep(interval)
                start += interval
        else:
            db_name = requests.get(url)
            if db_name.status_code == 200:
                return db_name.text
        return None

    def get_observed_generation(self):
        deployment_name = self.openshift.get_deployment_name_in_namespace(self.format_pattern(self.deployment_name_pattern), self.namespace)
        return self.openshift.get_resource_info_by_jsonpath("deployment", deployment_name, self.namespace, "{.status.observedGeneration}")

    def format_pattern(self, pattern):
        return pattern.format(name=self.name)

    def get_redeployed_rev_name(self, old_rev_name, interval=5, timeout=300):
        start = 0
        while ((start + interval) <= timeout):
            revisions = self.openshift.get_revisions(self.namespace)
            for rev in revisions:
                if rev != old_rev_name and re.match(self.name, rev) is not None:
                    new_revision = self.openshift.get_last_revision_status(rev, self.namespace)
                    if new_revision == 'True':
                        return rev
            time.sleep(interval)
            start += interval
        return None

    def get_rev_name_redeployed_by_generation(self, old_generation, interval=5, timeout=300):
        start = 0
        while ((start + interval) <= timeout):
            current_generation = self.get_generation()
            revisions = self.openshift.get_revisions(self.namespace)
            for rev in revisions:
                if (current_generation > old_generation) and (re.match(self.name, rev) is not None):
                    new_revision = self.openshift.get_last_revision_status(rev, self.namespace)
                    if new_revision == 'True':
                        return rev
            time.sleep(interval)
            start += interval
        return None

    def get_generation(self):
        deployment_name = self.openshift.get_deployment_name_in_namespace(self.format_pattern(self.deployment_name_pattern), self.namespace)
        return self.openshift.get_resource_info_by_jsonpath("deployment", deployment_name, self.namespace, "{.metadata.generation}")

    def get_deployment_with_intermediate_secret(self, intermediate_secret_name):
        return self.openshift.get_deployment_with_intermediate_secret_of_given_pattern(
            intermediate_secret_name, self.format_pattern(self.deployment_name_pattern), self.namespace, wait=True, timeout=120)
