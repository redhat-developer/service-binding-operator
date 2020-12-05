from app import App
import requests
import json
import polling2
from behave import given, step, then
from util import scenario_id


class GenericTestApp(App):

    def __init__(self, name, namespace, app_image="quay.io/redhat-developer/sbo-generic-test-app:20200923"):
        App.__init__(self, name, namespace, app_image, "8080")

    def get_env_var_value(self, name):
        resp = polling2.poll(lambda: requests.get(url=f"http://{self.route_url}/env/{name}"),
                             check_success=lambda r: r.status_code in [200, 404], step=5, timeout=400, ignore_exceptions=(requests.exceptions.ConnectionError,))
        print(f'env endpoint response: {resp.text} code: {resp.status_code}')
        if resp.status_code == 200:
            return json.loads(resp.text)
        else:
            return None

    def get_file_value(self, file_path):
        resp = polling2.poll(lambda: requests.get(url=f"http://{self.route_url}{file_path}"),
                             check_success=lambda r: r.status_code == 200, step=5, timeout=400, ignore_exceptions=(requests.exceptions.ConnectionError,))
        print(f'file endpoint response: {resp.text} code: {resp.status_code}')
        return resp.text


@given(u'Generic test application "{application_name}" is running')
@given(u'Generic test application "{application_name}" is running with binding root as "{bindingRoot}"')
@given(u'Generic test application is running')
def is_running(context, application_name=None, bindingRoot=None):
    if application_name is None:
        application_name = scenario_id(context)
    application = GenericTestApp(application_name, context.namespace.name)
    if not application.is_running():
        print("application is not running, trying to import it")
        application.install(bindingRoot=bindingRoot)
    context.application = application


@step(u'The application env var "{name}" has value "{value}"')
def check_env_var_value(context, name, value):
    found = polling2.poll(lambda: context.application.get_env_var_value(name) == value, step=5, timeout=400)
    assert found, f'Env var "{name}" should contain value "{value}"'


@step(u'The env var "{name}" is not available to the application')
def check_env_var_existence(context, name):
    output = polling2.poll(lambda: context.application.get_env_var_value(name) is None, step=5, timeout=400)
    assert output, f'Env var "{name}" should not exist'


@then(u'Content of file "{file_path}" in application pod is')
def check_file_value(context, file_path):
    value = context.text.strip()
    polling2.poll(lambda: context.application.get_file_value(file_path) == value, step=5, timeout=400)
