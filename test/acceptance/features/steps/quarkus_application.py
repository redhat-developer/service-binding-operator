from command import Command
import re
import time
import requests
from openshift import Openshift


class QuarkusApplication(object):

    cmd = Command()

    image_name_with_tag = "quay.io/pmacik/using-spring-data-jqa-quarkus:latest"
    api_end_point = '{route_url}/api/status/dbNameCM'
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

    def is_imported(self):
        deployment_name = self.openshift.get_deployment_name_in_namespace(
            self.format_pattern(self.deployment_name_pattern), self.namespace, wait=True, timeout=400)
        if deployment_name is None:
            return False
        else:
            deployment_status = self.openshift.get_deployment_status(deployment_name, self.namespace, wait_for_status="True")
            print(f"The deployment {deployment_name} status is {deployment_status}")
            output_match = re.search(r'True', deployment_status)
            assert output_match is not None, "Matched deployment status is not True"
            return True

    def get_db_name_from_api(self, wait=False, interval=5, timeout=300):
        route_url = self.openshift.get_knative_route_host(self.name, self.namespace)
        if route_url is None:
            return None
        if wait:
            start = 0
            while ((start + interval) <= timeout):
                url = self.api_end_point.format(route_url=route_url)
                db_name = requests.get(url)
                if db_name.status_code == 200:
                    return db_name.text
                time.sleep(interval)
                start += interval
        else:
            url = self.api_end_point.format(route_url=route_url)
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
            for rev in revisions.split(" "):
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
            for rev in revisions.split(" "):
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

    def get_deployment_names(self):
        return self.openshift.search_resource_lst_in_namespace("deployment", self.format_pattern(self.deployment_name_pattern), self.namespace)

    def get_deployment_with_intermediate_secret(self, intermediate_secret_name, wait=False, interval=5, timeout=300):

        # Expected result from 'oc' (openshift client) v4.6
        expected_secretRef_oc_46 = f'[{{"secretRef":{{"name":"{intermediate_secret_name}"}}}}]'
        # Expected result from 'oc' (openshift client) v4.5
        expected_secretRef_oc_45 = f'[map[secretRef:map[name:{intermediate_secret_name}]]]'

        deployment_name_pattern = self.format_pattern(self.deployment_name_pattern)
        if wait:
            start = 0
            while ((start + interval) <= timeout):
                deployment_list = self.get_deployment_names()
                if deployment_list is not None:
                    for deployment in deployment_list:
                        result = self.openshift.get_deployment_envFrom_info(deployment, self.namespace)
                        if result == expected_secretRef_oc_45 or result == expected_secretRef_oc_46:
                            return deployment
                        else:
                            print("\nUnexpected deployment's envFrom info: \n" +
                                  f"Expected: {expected_secretRef_oc_45} or {expected_secretRef_oc_46} \nbut was: {result}\n")
                else:
                    print(f"No deployment that matches {deployment_name_pattern} found.\n")
                time.sleep(interval)
                start += interval
        else:
            deployment_list = self.get_deployment_names()
            if deployment_list is not None:
                for deployment in deployment_list:
                    result = self.openshift.get_deployment_envFrom_info(deployment, self.namespace)
                    if result == expected_secretRef_oc_45 or result == expected_secretRef_oc_46:
                        return deployment
                    else:
                        print("\nUnexpected deployment's envFrom info: \n" +
                              f"Expected: {expected_secretRef_oc_45} or {expected_secretRef_oc_46} \nbut was: {result}\n")
            else:
                print(f"No deployment that matches {deployment_name_pattern} found.\n")
        return None
