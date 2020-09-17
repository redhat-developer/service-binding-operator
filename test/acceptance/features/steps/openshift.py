import re
import time
from pyshould import should
from command import Command


class Openshift(object):
    def __init__(self):
        self.cmd = Command()
        self.catalog_source_yaml_template = '''
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
    name: {name}
    namespace: openshift-marketplace
spec:
    sourceType: grpc
    image: {catalog_image}
    displayName: {name} OLM registry
    updateStrategy:
        registryPoll:
            interval: 30m
'''
        self.image_stream_template = '''
---
apiVersion: image.openshift.io/v1
kind: ImageStream
metadata:
    name: {name}
    namespace: {namespace}
'''
        self.build_config_template = '''
---
apiVersion: build.openshift.io/v1
kind: BuildConfig
metadata:
  name: {name}
  namespace: {namespace}
spec:
  source:
    git:
      ref: master
      uri: {application_source}
    type: Git
  strategy:
    sourceStrategy:
      from:
        kind: ImageStreamTag
        name: {image_name_with_tag}
        namespace: openshift
    type: Source
  output:
    to:
      kind: ImageStreamTag
      name: {name}:latest
  triggers:
    - type: ConfigChange
'''
        self.service_template = '''
---
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {name}
  namespace: {namespace}
spec:
  template:
    spec:
      containers:
        - image: {image_repository}
'''
        self.operator_subscription_to_namespace_yaml_template = '''
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: '{name}'
  namespace: {namespace}
spec:
  channel: '{channel}'
  installPlanApproval: Automatic
  name: '{name}'
  source: '{operator_source_name}'
  sourceNamespace: openshift-marketplace
  startingCSV: '{csv_version}'
'''

    def get_pod_lst(self, namespace):
        return self.get_resource_lst("pods", namespace)

    def get_resource_lst(self, resource_plural, namespace):
        output, exit_code = self.cmd.run(f'oc get {resource_plural} -n {namespace} -o "jsonpath={{.items[*].metadata.name}}"')
        assert exit_code == 0, f"Getting resource list failed as the exit code is not 0 with output - {output}"
        return output

    def search_item_in_lst(self, lst, search_pattern):
        lst_arr = lst.split(" ")
        for item in lst_arr:
            if re.fullmatch(search_pattern, item) is not None:
                print(f"item matched {item}")
                return item
        print("Given item not matched from the list of pods")
        return None

    def search_pod_in_namespace(self, pod_name_pattern, namespace):
        return self.search_resource_in_namespace("pods", pod_name_pattern, namespace)

    def search_resource_in_namespace(self, resource_plural, name_pattern, namespace):
        print(f"Searching for {resource_plural} that matches {name_pattern} in {namespace} namespace")
        lst = self.get_resource_lst(resource_plural, namespace)
        if len(lst) != 0:
            print("Resource list is {}".format(lst))
            return self.search_item_in_lst(lst, name_pattern)
        else:
            print('Resource list is empty under namespace - {}'.format(namespace))
            return None

    def is_resource_in(self, resource_type):
        output, exit_code = self.cmd.run(f'oc get {resource_type}')
        return exit_code == 0

    def wait_for_pod(self, pod_name_pattern, namespace, interval=5, timeout=60):
        pod = self.search_pod_in_namespace(pod_name_pattern, namespace)
        start = 0
        if pod is not None:
            return pod
        else:
            while ((start + interval) <= timeout):
                pod = self.search_pod_in_namespace(pod_name_pattern, namespace)
                if pod is not None:
                    return pod
                time.sleep(interval)
                start += interval
        return None

    def check_pod_status(self, pod_name, namespace, wait_for_status="Running"):
        cmd = f'oc get pod {pod_name} -n {namespace} -o "jsonpath={{.status.phase}}"'
        status_found, output, exit_status = self.cmd.run_wait_for_status(cmd, wait_for_status)
        return status_found

    def get_pod_status(self, pod_name, namespace):
        cmd = f'oc get pod {pod_name} -n {namespace} -o "jsonpath={{.status.phase}}"'
        output, exit_status = self.cmd.run(cmd)
        print(f"Get pod status: {output}, {exit_status}")
        if exit_status == 0:
            return output
        return None

    def oc_apply(self, yaml):
        (output, exit_code) = self.cmd.run("oc apply -f -", yaml)
        print(output)
        return output

    def create_catalog_source(self, name, catalog_image):
        catalog_source = self.catalog_source_yaml_template.format(name=name, catalog_image=catalog_image)
        return self.oc_apply(catalog_source)

    def get_current_csv(self, package_name, catalog, channel):
        cmd = f'oc get packagemanifests -o json | jq -r \'.items[] \
            | select(.metadata.name=="{package_name}") \
            | select(.status.catalogSource=="{catalog}").status.channels[] \
            | select(.name=="{channel}").currentCSV\''
        current_csv, exit_code = self.cmd.run(cmd)
        if exit_code != 0:
            print(f"\nNon-zero exit code ({exit_code}) returned while getting currentCSV: {current_csv}")
            return None
        current_csv = current_csv.strip("\n")
        return current_csv

    def wait_for_package_manifest(self, package_name, operator_source_name, operator_channel, interval=5, timeout=120):
        current_csv = self.get_current_csv(package_name, operator_source_name, operator_channel)
        start = 0
        if current_csv is not None:
            return True
        else:
            while ((start + interval) <= timeout):
                current_csv = self.get_current_csv(package_name, operator_source_name, operator_channel)
                if current_csv is not None:
                    return True
                time.sleep(interval)
                start += interval
        return False

    def expose_service_route(self, service_name, namespace):
        output, exit_code = self.cmd.run(f'oc expose svc/{service_name} -n {namespace} --name={service_name}')
        return re.search(r'.*%s\sexposed' % service_name, output)

    def get_route_host(self, name, namespace):
        output, exit_code = self.cmd.run(f'oc get route {name} -n {namespace} -o "jsonpath={{.status.ingress[0].host}}"')
        assert exit_code == 0, f"Getting route host failed as the exit code is not 0 with output - {output}"
        return output

    def get_deployment_status(self, deployment_name, namespace, wait_for_status=None):
        deployment_status_cmd = f'oc get deployment {deployment_name} -n {namespace} -o json' \
            + ' | jq -rc \'.status.conditions[] | select(.type=="Available").status\''
        output = None
        exit_code = -1
        if wait_for_status is not None:
            status_found, output, exit_code = self.cmd.run_wait_for_status(deployment_status_cmd, wait_for_status, 5, 400)
            if exit_code == 0:
                assert status_found is True, f"Deployment {deployment_name} result after waiting for status is {status_found}"
        else:
            output, exit_code = self.cmd.run(deployment_status_cmd)
        assert exit_code == 0, "Getting deployment status failed as the exit code is not 0"

        return output

    def get_deployment_env_info(self, name, namespace):
        env_cmd = f'oc get deploy {name} -n {namespace} -o "jsonpath={{.spec.template.spec.containers[0].env}}"'
        env, exit_code = self.cmd.run(env_cmd)
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned while getting deployment's env: {env}"
        return env

    def get_deployment_envFrom_info(self, name, namespace):
        env_from_cmd = f'oc get deploy {name} -n {namespace} -o "jsonpath={{.spec.template.spec.containers[0].envFrom}}"'
        env_from, exit_code = self.cmd.run(env_from_cmd)
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned while getting deployment's envFrom: {env_from}"
        return env_from

    def get_resource_info_by_jsonpath(self, resource_type, name, namespace, json_path, wait=False, interval=5, timeout=120):
        oc_cmd = f'oc get {resource_type} {name} -n {namespace} -o "jsonpath={json_path}"'
        output, exit_code = self.cmd.run(oc_cmd)
        if exit_code != 0 or output == "" or output == "<nil>":
            if wait:
                attempts = timeout/interval
                while (exit_code != 0 or output == "" or output == "<nil>") and attempts > 0:
                    output, exit_code = self.cmd.run(oc_cmd)
                    attempts -= 1
                    time.sleep(interval)
        exit_code | should.be_equal_to(0).desc(f'Exit code should be 0:\n OUTPUT:\n{output}')
        return output

    def get_resource_info_by_jq(self, resource_type, name, namespace, jq_expression, wait=False, interval=5, timeout=120):
        output, exit_code = self.cmd.run(f'oc get {resource_type} {name} -n {namespace} -o json | jq -rc \'{jq_expression}\'')
        if exit_code != 0:
            if wait:
                attempts = timeout/interval
                while exit_code != 0 and attempts > 0:
                    output, exit_code = self.cmd.run(f'oc get {resource_type} {name} -n {namespace} -o json | jq -rc \'{jq_expression}\'')
                    attempts -= 1
                    time.sleep(interval)
        exit_code | should.be_equal_to(0).desc(f'Exit code should be 0:\n OUTPUT:\n{output}')
        return output.rstrip("\n")

    def create_image_stream(self, name, registry_namespace):
        image_stream = self.image_stream_template.format(name=name, namespace=registry_namespace)
        return self.oc_apply(image_stream)

    def get_docker_image_repository(self, name, namespace):
        cmd = f'oc get is {name} -n {namespace} -o "jsonpath={{.status.dockerImageRepository}}"'
        (output, exit_code) = self.cmd.run(cmd)
        assert exit_code == 0, f"cmd-{cmd} result for getting docker image repository is {output} with exit code-{exit_code} not equal to 0"
        return output

    def create_build_config(self, name, namespace, application_source, image_name_with_tag):
        build_config_yaml = self.build_config_template.format(
            name=name, namespace=namespace, application_source=application_source, image_name_with_tag=image_name_with_tag)
        return self.oc_apply(build_config_yaml)

    def create_knative_service(self, name, namespace, image):
        knative_service_yaml = self.service_template.format(name=name, namespace=namespace, image_repository=image)
        return self.oc_apply(knative_service_yaml)

    def wait_for_build_pod_status(self, build_pod_name, namespace, wait_for_status="Succeeded", timeout=780):
        cmd = f'oc get pod {build_pod_name} -n {namespace} -o "jsonpath={{.status.phase}}"'
        status_found, output, exit_status = self.cmd.run_wait_for_status(cmd, wait_for_status, timeout=timeout)
        return status_found, output

    def get_deployment_name_in_namespace(self, deployment_name_pattern, namespace, wait=False, interval=5, timeout=120):
        if wait:
            start = 0
            while ((start + interval) <= timeout):
                deployment = self.search_resource_in_namespace("deployment", deployment_name_pattern, namespace)
                if deployment is not None:
                    return deployment
                time.sleep(interval)
                start += interval
            return None
        else:
            return self.search_resource_in_namespace("deployment", deployment_name_pattern, namespace)

    def get_knative_route_host(self, name, namespace):
        cmd = f'oc get rt {name} -n {namespace} -o "jsonpath={{.status.url}}"'
        output, exit_code = self.cmd.run(cmd)
        assert exit_code == 0, f"cmd-{cmd} result for getting knative route is {output} with exit code not equal to 0"
        return output

    def get_revisions(self, namespace):
        return self.get_resource_lst("rev", namespace)

    def get_last_revision_status(self, revision, namespace):
        cmd = f'oc get rev {revision} -n {namespace} -o "jsonpath={{.status.conditions[*].status}}"'
        (output, exit_code) = self.cmd.run(cmd)
        assert exit_code == 0, f"cmd-{cmd} for getting last revision status is {output} with exit code not equal to 0"
        last_revision_status = output.split(" ")[-1]
        return last_revision_status

    def create_operator_subscription_to_namespace(self, package_name, namespace, operator_source_name, channel):
        operator_subscription = self.operator_subscription_to_namespace_yaml_template.format(
            name=package_name, namespace=namespace, operator_source_name=operator_source_name,
            channel=channel, csv_version=self.get_current_csv(package_name, operator_source_name, channel))
        return self.oc_apply(operator_subscription)

    def create_operator_subscription(self, package_name, operator_source_name, channel):
        return self.create_operator_subscription_to_namespace(package_name, "openshift-operators", operator_source_name, channel)

    def get_resource_list_in_namespace(self, resource_plural, name_pattern, namespace):
        print(f"Searching for {resource_plural} that matches {name_pattern} in {namespace} namespace")
        lst = self.get_resource_lst(resource_plural, namespace)
        if len(lst) != 0:
            print("Resource list is {}".format(lst))
            return self.get_all_matched_pattern_from_lst(lst, name_pattern)
        else:
            print('Resource list is empty under namespace - {}'.format(namespace))
            return None

    def get_all_matched_pattern_from_lst(self, lst, search_pattern):
        lst_arr = lst.split(" ")
        output_arr = []
        for item in lst_arr:
            if re.fullmatch(search_pattern, item) is not None:
                print(f"item matched {item}")
                output_arr.append(item)
        if not output_arr:
            print("Given item not matched from the list of pods")
            return None
        else:
            return output_arr

    def search_resource_lst_in_namespace(self, resource_plural, name_pattern, namespace):
        print(f"Searching for {resource_plural} that matches {name_pattern} in {namespace} namespace")
        lst = self.get_resource_list_in_namespace(resource_plural, name_pattern, namespace)
        if len(lst) != 0:
            print("Resource list is {}".format(lst))
            return lst
        print('Resource list is empty under namespace - {}'.format(namespace))
        return None

    def oc_apply_yaml_file(self, yaml):
        (output, exit_code) = self.cmd.run("oc apply -f " + yaml)
        assert exit_code == 0, "Applying yaml file failed as the exit code is not 0"
        return output

    def get_deployment_names_of_given_pattern(self, deployment_name_pattern, namespace):
        return self.search_resource_lst_in_namespace("deployment", deployment_name_pattern, namespace)

    def get_deployment_with_intermediate_secret_of_given_pattern(self, intermediate_secret_name, deployment_name_pattern,
                                                                 namespace, wait=False, interval=5, timeout=300):
        # Expected result from 'oc' (openshift client) v4.5
        expected_secretRef_oc_45 = f'secretRef:map[name:{intermediate_secret_name}]'
        # Expected result from 'oc' (openshift client) v4.6+
        expected_secretRef_oc_46 = f'{{"secretRef":{{"name":"{intermediate_secret_name}"}}}}'
        if wait:
            start = 0
            while ((start + interval) <= timeout):
                deployment_list = self.get_deployment_names_of_given_pattern(deployment_name_pattern, namespace)
                if deployment_list is not None:
                    for deployment in deployment_list:
                        result = self.get_deployment_envFrom_info(deployment, namespace)
                        if re.search(re.escape(expected_secretRef_oc_46), result) is not None or \
                                re.search(re.escape(expected_secretRef_oc_45), result) is not None:
                            return deployment
                        else:
                            print(
                                f"\nUnexpected deployment's envFrom info: \nExpected: {expected_secretRef_oc_46} or \
                                {expected_secretRef_oc_45} \nbut was: {result}\n")
                else:
                    print(f"No deployment that matches {deployment_name_pattern} found.\n")
                time.sleep(interval)
                start += interval
        else:
            deployment_list = self.get_deployment_names_of_given_pattern(deployment_name_pattern)
            if deployment_list is not None:
                for deployment in deployment_list:
                    result = self.get_deployment_envFrom_info(deployment, namespace)
                    if result == expected_secretRef_oc_46 or result == expected_secretRef_oc_45:
                        return deployment
                    else:
                        print(
                            f"\nUnexpected deployment's envFrom info: \nExpected: {expected_secretRef_oc_46} or \
                            {expected_secretRef_oc_45} \nbut was: {result}\n")
            else:
                print(f"No deployment that matches {deployment_name_pattern} found.\n")
        return None
