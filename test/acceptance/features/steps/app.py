from openshift import Openshift
from command import Command
from environment import ctx
from behave import step
from util import substitute_scenario_id
import polling2
import json


class App(object):
    openshift = Openshift()
    cmd = Command()
    name = ""
    namespace = ""
    app_image = ""
    route_url = ""
    port = ""
    bindingRoot = ""
    resource = ""

    def __init__(self, name, namespace, app_image, port="", resource="deployment"):
        self.name = name
        self.namespace = namespace
        self.app_image = app_image
        self.port = port
        self.resource = resource

    def is_running(self, wait=False):
        output, exit_code = self.cmd.run(
            f"{ctx.cli} wait --for=condition=Available=True {self.resource}/{self.name} -n {self.namespace} --timeout={300 if wait else 0}s")
        running = exit_code == 0
        if running:
            self.route_url = polling2.poll(lambda: self.base_url(),
                                           check_success=lambda v: v != "", step=1, timeout=100)
        return running

    def install(self, bindingRoot=None):
        self.openshift.new_app(self.name, self.app_image, self.namespace, bindingRoot, self.resource == "deploymentconfig")
        self.openshift.expose_service_route(self.name, self.namespace, self.port)
        return self.is_running(wait=True)

    def base_url(self):
        return self.openshift.get_route_host(self.name, self.namespace)

    def get_generation(self):
        deployment_name = self.openshift.get_deployment_name_in_namespace(
                            self.format_pattern(self.deployment_name_pattern), self.namespace, resource=self.resource)
        generation = polling2.poll(
                target=lambda: self.openshift.get_resource_info_by_jsonpath(self.resource, deployment_name, self.namespace, "{.metadata.generation}"),
                check_success=lambda x: x is not None,
                step=5, timeout=800)
        return int(generation)


@step(u'jsonpath "{json_path}" on "{res_name}" should return "{json_value}"')
@step(u'jsonpath "{json_path}" on "{res_name}" should return no value')
def resource_jsonpath_value(context, json_path, res_name, json_value=""):
    openshift = Openshift()
    json_path = substitute_scenario_id(context, json_path)
    res_name = substitute_scenario_id(context, res_name)
    json_value = substitute_scenario_id(context, json_value)
    (crdName, name) = res_name.split("/")
    polling2.poll(lambda: openshift.get_resource_info_by_jsonpath(crdName, name, context.namespace.name, json_path) == json_value,
                  step=5, timeout=800, ignore_exceptions=(json.JSONDecodeError,))


@step(u'jsonpath "{json_path}" on "{res_name}" should contain "{json_value}"')
def resource_jsonpath_contains(context, json_path, res_name, json_value):
    openshift = Openshift()
    json_path = substitute_scenario_id(context, json_path)
    res_name = substitute_scenario_id(context, res_name)
    json_value = json.loads(substitute_scenario_id(context, json_value))
    (crdName, name) = res_name.split("/")
    actual_value = polling2.poll(lambda: openshift.get_resource_info_by_jsonpath(crdName, name, context.namespace.name, json_path),
                                 step=5, timeout=800, ignore_exceptions=(json.JSONDecodeError))
    data = json.loads(actual_value)
    assert json_value in data, "Expected {str(json_value)} in {str(data)}"
