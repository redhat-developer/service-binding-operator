from app import App
import requests
import json
import polling2
from behave import step
from openshift import Openshift
from util import substitute_scenario_id
from string import Template


class GenericTestApp(App):

    deployment_name_pattern = "{name}"

    def __init__(self, name, namespace, app_image="quay.io/service-binding/generic-test-app:20211112"):
        App.__init__(self, name, namespace, app_image, "8080")

    def get_env_var_value(self, name):
        resp = polling2.poll(lambda: requests.get(url=f"http://{self.route_url}/env/{name}"),
                             check_success=lambda r: r.status_code in [200, 404], step=5, timeout=400, ignore_exceptions=(requests.exceptions.ConnectionError,))
        print(f'env endpoint response: {resp.text} code: {resp.status_code}')
        if resp.status_code == 200:
            return json.loads(resp.text)
        else:
            return None

    def format_pattern(self, pattern):
        return pattern.format(name=self.name)

    def get_file_value(self, file_path):
        resp = polling2.poll(lambda: requests.get(url=f"http://{self.route_url}{file_path}"),
                             check_success=lambda r: r.status_code == 200, step=5, timeout=400, ignore_exceptions=(requests.exceptions.ConnectionError,))
        print(f'file endpoint response: {resp.text} code: {resp.status_code}')
        return resp.text

    def assert_file_not_exist(self, file_path):
        polling2.poll(lambda: requests.get(url=f"http://{self.route_url}{file_path}"),
                      check_success=lambda r: r.status_code == 404, step=5, timeout=400, ignore_exceptions=(requests.exceptions.ConnectionError,))

    def set_label(self, label):
        self.openshift.set_label(self.name, label, self.namespace)


@step(u'Generic test application "{application_name}" is running')
@step(u'Generic test application "{application_name}" is running with binding root as "{bindingRoot}"')
@step(u'Generic test application is running')
@step(u'Generic test application is running with binding root as "{bindingRoot}"')
def is_running(context, application_name=None, bindingRoot=None, asDeploymentConfig=False):
    if application_name is None:
        application_name = substitute_scenario_id(context)
    application = GenericTestApp(application_name, context.namespace.name)
    if asDeploymentConfig:
        application.resource = "deploymentconfig"
    if not application.is_running():
        print("application is not running, trying to import it")
        application.install(bindingRoot=bindingRoot)
    context.application = application

    # save the generation number
    context.original_application_generation = application.get_generation()
    context.latest_application_generation = application.get_generation()


@step(u'Generic test application is running as deployment config')
def is_running_deployment_config(context):
    is_running(context, asDeploymentConfig=True)


@step(u'The application env var "{name}" has value "{value}"')
def check_env_var_value(context, name, value):
    value = substitute_scenario_id(context, value)
    found = polling2.poll(lambda: context.application.get_env_var_value(name) == value, step=5, timeout=400)
    assert found, f'Env var "{name}" should contain value "{value}"'


@step(u'The env var "{name}" is not available to the application')
def check_env_var_existence(context, name):
    output = polling2.poll(lambda: context.application.get_env_var_value(name) is None, step=5, timeout=400)
    assert output, f'Env var "{name}" should not exist'


@step(u'Content of file "{file_path}" in application pod is')
def check_file_value(context, file_path):
    value = Template(context.text.strip()).substitute(NAMESPACE=context.namespace.name)
    resource = substitute_scenario_id(context, file_path)
    polling2.poll(lambda: context.application.get_file_value(resource) == value, step=5, timeout=400)


@step(u'File "{file_path}" exists in application pod')
def check_file_exists(context, file_path):
    resource = substitute_scenario_id(context, file_path)
    polling2.poll(lambda: context.application.get_file_value(resource) != "", step=5, timeout=400)


@step(u'File "{file_path}" is unavailable in application pod')
def check_file_unavailable(context, file_path):
    context.application.assert_file_not_exist(file_path)


@step(u'Test applications "{first_app_name}" and "{second_app_name}" is running')
def are_two_apps_running(context, first_app_name, second_app_name, bindingRoot=None):
    application1 = GenericTestApp(first_app_name, context.namespace.name)
    if not application1.is_running():
        print("application1 is not running, trying to import it")
        application1.install(bindingRoot=bindingRoot)
    context.application1 = application1

    application2 = GenericTestApp(second_app_name, context.namespace.name)
    if not application2.is_running():
        print("application2 is not running, trying to import it")
        application2.install(bindingRoot=bindingRoot)
    context.application2 = application2


@step(u'The common label "{label}" is set for both apps')
def set_common_label(context, label):
    context.application1.set_label(f"{label}")
    context.application2.set_label(f"{label}")


@step(u'The application env var "{name}" has value "{value}" in both apps')
def check_env_var_value_in_both_apps(context, name, value):
    polling2.poll(lambda: context.application1.get_env_var_value(name) == value, step=5, timeout=400)
    polling2.poll(lambda: context.application2.get_env_var_value(name) == value, step=5, timeout=400)


@step(u'The container declared in application resource contains env "{envVar}" set only once')
@step(u'The container declared in application "{app_name}" resource contains env "{envVar}" set only once')
def check_env_var_count_set_on_container(context, envVar, app_name=None):
    openshift = Openshift()
    if app_name is None:
        app_name = context.application.name
    app_name = substitute_scenario_id(context, app_name)
    env = openshift.get_deployment_env_info(app_name, context.namespace.name)
    assert str(env).count(envVar) == 1
