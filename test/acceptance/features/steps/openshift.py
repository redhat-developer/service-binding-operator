import re
import time
import base64
import json
from environment import ctx
from command import Command


class Openshift(object):
    def __init__(self):
        self.cmd = Command()
        self.olm_namespace = "olm" if ctx.cli == "kubectl" else "openshift-marketplace"
        self.operators_namespace = "operators" if ctx.cli == "kubectl" else "openshift-operators"
        self.catalog_source_yaml_template = '''
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
    name: {name}
    namespace: {olm_namespace}
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
  sourceNamespace: {olm_namespace}
  startingCSV: '{csv_version}'
'''
        self.deployment_template = '''
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: '{name}'
  namespace: {namespace}
  labels:
    app: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: {image_name}
        env:
        - name: SERVICE_BINDING_ROOT
          value: {bindingRoot}
'''

    def get_pod_lst(self, namespace):
        return self.get_resource_lst("pods", namespace)

    def get_resource_lst(self, resource_plural, namespace):
        output, exit_code = self.cmd.run(f'{ctx.cli} get {resource_plural} -n {namespace} -o "jsonpath={{.items[*].metadata.name}}"')
        assert exit_code == 0, f"Getting resource list failed as the exit code is not 0 with output - {output}"
        return output.split(" ")

    def search_item_in_lst(self, lst, search_pattern):
        for item in lst:
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
        output, exit_code = self.cmd.run(f'{ctx.cli} get {resource_type}')
        return exit_code == 0

    def wait_for_pod(self, pod_name_pattern, namespace, interval=5, timeout=600):
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
        cmd = f'{ctx.cli} get pod {pod_name} -n {namespace} -o "jsonpath={{.status.phase}}"'
        status_found, output, exit_status = self.cmd.run_wait_for_status(cmd, wait_for_status)
        return status_found

    def get_pod_status(self, pod_name, namespace):
        cmd = f'{ctx.cli} get pod {pod_name} -n {namespace} -o "jsonpath={{.status.phase}}"'
        output, exit_status = self.cmd.run(cmd)
        print(f"Get pod status: {output}, {exit_status}")
        if exit_status == 0:
            return output
        return None

    def apply(self, yaml, namespace=None, user=None):
        if namespace is not None:
            ns_arg = f"-n {namespace}"
        else:
            ns_arg = ""
        if user is not None:
            user_arg = f"--user={user}"
        else:
            user_arg = ""
        (output, exit_code) = self.cmd.run(f"{ctx.cli} apply {ns_arg} {user_arg} -f -", yaml)
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) while applying a YAML: {output}"
        return output

    def apply_invalid(self, yaml, namespace=None):
        if namespace is not None:
            ns_arg = f"-n {namespace}"
        else:
            ns_arg = ""
        (output, exit_code) = self.cmd.run(f"{ctx.cli} apply {ns_arg} -f -", yaml)
        assert exit_code != 0, f"the command should fail but it did not, output: {output}"
        return output

    def delete(self, yaml, namespace=None):
        if namespace is not None:
            ns_arg = f"-n {namespace}"
        else:
            ns_arg = ""
        (output, exit_code) = self.cmd.run(f"{ctx.cli} delete {ns_arg} -f -", yaml)
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) while deleting a YAML: {output}"
        return output

    def create_catalog_source(self, name, catalog_image):
        catalog_source = self.catalog_source_yaml_template.format(name=name, catalog_image=catalog_image, olm_namespace=self.olm_namespace)
        return self.apply(catalog_source)

    def get_current_csv(self, package_name, catalog, channel):
        cmd = f'{ctx.cli} get packagemanifests -o json | jq -r \'.items[] \
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

    def expose_service_route(self, name, namespace, port=""):
        if ctx.cli == "oc":
            output, exit_code = self.cmd.run(f'{ctx.cli} expose svc/{name} -n {namespace} --name={name}')
        else:
            output, exit_code = self.cmd.run(f'{ctx.cli} expose deployment {name} -n {namespace} --port={port} --type=NodePort')
        assert exit_code == 0, f"Could not expose deployment: {output}"

    def get_route_host(self, name, namespace):
        if ctx.cli == "oc":
            output, exit_code = self.cmd.run(f'{ctx.cli} get route {name} -n {namespace} -o "jsonpath={{.status.ingress[0].host}}"')
            host = output
        else:
            addr = self.get_node_address()
            output, exit_code = self.cmd.run(f'{ctx.cli} get service {name} -n {namespace} -o "jsonpath={{.spec.ports[0].nodePort}}"')
            host = f"{addr}:{output}"

        assert exit_code == 0, f"Getting route host failed as the exit code is not 0 with output - {output}"

        return host

    def get_node_address(self):
        output, exit_code = self.cmd.run(f'{ctx.cli} get nodes -o "jsonpath={{.items[0].status.addresses}}"')
        assert exit_code == 0, f"Error accessing Node resources - {output}"
        addresses = json.loads(output)
        for addr in addresses:
            if addr['type'] in ["InternalIP", "ExternalIP"]:
                return addr['address']
        assert False, f"No IP addresses found in {output}"

    def get_deployment_status(self, deployment_name, namespace, wait_for_status=None, interval=5, timeout=400):
        deployment_status_cmd = f'{ctx.cli} get deployment {deployment_name} -n {namespace} -o json' \
            + ' | jq -rc \'.status.conditions[] | select(.type=="Available").status\''
        output = None
        exit_code = -1
        if wait_for_status is not None:
            status_found, output, exit_code = self.cmd.run_wait_for_status(deployment_status_cmd, wait_for_status, interval, timeout)
            if exit_code == 0:
                assert status_found is True, f"Deployment {deployment_name} result after waiting for status is {status_found}"
        else:
            output, exit_code = self.cmd.run(deployment_status_cmd)
        assert exit_code == 0, "Getting deployment status failed as the exit code is not 0"

        return output

    def get_deployment_env_info(self, name, namespace):
        env_cmd = f'{ctx.cli} get deploy {name} -n {namespace} -o "jsonpath={{.spec.template.spec.containers[0].env}}"'
        env, exit_code = self.cmd.run(env_cmd)
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned while getting deployment's env: {env}"
        return env

    def get_deployment_envFrom_info(self, name, namespace):
        env_from_cmd = f'{ctx.cli} get deploy {name} -n {namespace} -o "jsonpath={{.spec.template.spec.containers[0].envFrom}}"'
        env_from, exit_code = self.cmd.run(env_from_cmd)
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned while getting deployment's envFrom: {env_from}"
        return env_from

    def get_resource_info_by_jsonpath(self, resource_type, name, namespace, json_path="{.*}", user=None):
        oc_cmd = f'{ctx.cli} get {resource_type} {name} -n {namespace} -o "jsonpath={json_path}"'
        if user:
            oc_cmd += f" --user={user}"
        output, exit_code = self.cmd.run(oc_cmd)
        if exit_code == 0:
            if resource_type == "secrets":
                return base64.decodebytes(bytes(output, 'utf-8')).decode('utf-8')
            else:
                return output
        else:
            print(f'Error getting value for {resource_type}/{name} in {namespace} path={json_path}: {output}')
            return None

    def get_resource_info_by_jq(self, resource_type, name, namespace, jq_expression, wait=False, interval=5, timeout=120):
        output, exit_code = self.cmd.run(f'{ctx.cli} get {resource_type} {name} -n {namespace} -o json | jq  \'{jq_expression}\'')
        return output

    def create_image_stream(self, name, registry_namespace):
        image_stream = self.image_stream_template.format(name=name, namespace=registry_namespace)
        return self.apply(image_stream)

    def get_docker_image_repository(self, name, namespace):
        cmd = f'{ctx.cli} get is {name} -n {namespace} -o "jsonpath={{.status.dockerImageRepository}}"'
        (output, exit_code) = self.cmd.run(cmd)
        assert exit_code == 0, f"cmd-{cmd} result for getting docker image repository is {output} with exit code-{exit_code} not equal to 0"
        return output

    def create_build_config(self, name, namespace, application_source, image_name_with_tag):
        build_config_yaml = self.build_config_template.format(
            name=name, namespace=namespace, application_source=application_source, image_name_with_tag=image_name_with_tag)
        return self.apply(build_config_yaml)

    def create_knative_service(self, name, namespace, image):
        knative_service_yaml = self.service_template.format(name=name, namespace=namespace, image_repository=image)
        return self.apply(knative_service_yaml)

    def wait_for_build_pod_status(self, build_pod_name, namespace, wait_for_status="Succeeded", timeout=780):
        cmd = f'{ctx.cli} get pod {build_pod_name} -n {namespace} -o "jsonpath={{.status.phase}}"'
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
        cmd = f'{ctx.cli} get rt {name} -n {namespace} -o "jsonpath={{.status.url}}"'
        output, exit_code = self.cmd.run(cmd)
        assert exit_code == 0, f"cmd-{cmd} result for getting knative route is {output} with exit code not equal to 0"
        return output

    def get_revisions(self, namespace):
        return self.get_resource_lst("rev", namespace)

    def get_last_revision_status(self, revision, namespace):
        cmd = f'{ctx.cli} get rev {revision} -n {namespace} -o "jsonpath={{.status.conditions[*].status}}"'
        (output, exit_code) = self.cmd.run(cmd)
        assert exit_code == 0, f"cmd-{cmd} for getting last revision status is {output} with exit code not equal to 0"
        last_revision_status = output.split(" ")[-1]
        return last_revision_status

    def create_operator_subscription_to_namespace(self, package_name, namespace, operator_source_name, channel):
        operator_subscription = self.operator_subscription_to_namespace_yaml_template.format(
            name=package_name, namespace=namespace, operator_source_name=operator_source_name, olm_namespace=self.olm_namespace,
            channel=channel, csv_version=self.get_current_csv(package_name, operator_source_name, channel))
        return self.apply(operator_subscription)

    def create_operator_subscription(self, package_name, operator_source_name, channel):
        return self.create_operator_subscription_to_namespace(package_name, self.operators_namespace, operator_source_name, channel)

    def get_resource_list_in_namespace(self, resource_plural, name_pattern, namespace):
        print(f"Searching for {resource_plural} that matches {name_pattern} in {namespace} namespace")
        lst = self.get_resource_lst(resource_plural, namespace)
        if len(lst) != 0:
            print("Resource list is {}".format(lst))
            return self.get_all_matched_pattern_from_lst(lst, name_pattern)
        else:
            print('Resource list is empty under namespace - {}'.format(namespace))
            return []

    def get_all_matched_pattern_from_lst(self, lst, search_pattern):
        output_arr = []
        for item in lst:
            if re.fullmatch(search_pattern, item) is not None:
                print(f"item matched {item}")
                output_arr.append(item)
        if not output_arr:
            print("Given item not matched from the list of pods")
            return []
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

    def apply_yaml_file(self, yaml, namespace=None):
        if namespace is not None:
            ns_arg = f"-n {namespace}"
        else:
            ns_arg = ""
        (output, exit_code) = self.cmd.run(f"{ctx.cli} apply {ns_arg} -f " + yaml)
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
            deployment_list = self.get_deployment_names_of_given_pattern(deployment_name_pattern, namespace)
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

    def delete_service_binding(self, sb_name, namespace):
        (output, exit_code) = self.cmd.run(f"{ctx.cli} delete --wait=true --timeout=120s ServiceBinding {sb_name} -n {namespace}")
        assert exit_code == 0, f"Unexpected exit code ({exit_code}) while deleting service binding '{sb_name}': {output}"

    def new_app(self, name, image_name, namespace, bindingRoot=None):
        if ctx.cli == "oc":
            cmd = f"{ctx.cli} new-app --docker-image={image_name} --name={name} -n {namespace}"
            if bindingRoot:
                cmd = cmd + f" -e SERVICE_BINDING_ROOT={bindingRoot}"
            (output, exit_code) = self.cmd.run(cmd)
        else:
            cmd = f"{ctx.cli} create deployment {name} -n {namespace} --image={image_name}"
            if bindingRoot:
                (output, exit_code) = self.cmd.run(f"{ctx.cli} apply -f -",
                                                   self.deployment_template.format(name=name, image_name=image_name,
                                                                                   namespace=namespace, bindingRoot=bindingRoot))
            else:
                (output, exit_code) = self.cmd.run(cmd)
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned when attempting to create a new app using following command line {cmd}\n: {output}"

    def cli(self, cmd, namespace):
        output, exit_status = self.cmd.run(f'{ctx.cli} {cmd} -n {namespace}')
        assert exit_status == 0, "Exit should be zero"
        return output
