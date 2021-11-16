import yaml
import polling2
import json
from behave import step, when, then
from openshift import Openshift
from util import substitute_scenario_id


class ServiceBinding(object):

    openshift = Openshift()
    yamlContent = ""
    crdName = ""
    name = ""
    namespace = ""

    def __init__(self, yamlContent, namespace=None):
        self.yamlContent = yamlContent
        res = yaml.full_load(yamlContent)
        self.name = res["metadata"]["name"]
        self.namespace = namespace
        apiVersion = res["apiVersion"]
        self.crdName = f"servicebindings.{apiVersion.split('/')[0]}"
        if apiVersion == "servicebinding.io":
            self.secretPath = '{.status.binding.name}'
        else:
            self.secretPath = '{.status.secret}'

    def create(self, user):
        return self.openshift.apply(self.yamlContent, self.namespace, user)

    def attempt_to_create_invalid(self):
        return self.openshift.apply_invalid(self.yamlContent, self.namespace)

    def get_info_by_jsonpath(self, json_path):
        if json_path.startswith("{"):
            return self.openshift.get_resource_info_by_jsonpath(self.crdName, self.name, self.namespace, json_path)
        else:
            return self.openshift.get_resource_info_by_jq(self.crdName, self.name, self.namespace, json_path)

    def get_secret_name(self):
        output = self.get_info_by_jsonpath(self.secretPath)
        assert output is not None, "Failed to fetch secret name from ServiceBinding"
        return output.strip().strip('"')

    def delete(self):
        self.openshift.delete(self.yamlContent, self.namespace)


@step(u'Service Binding is applied')
@step(u'user {user} applies Service Binding')
def sbr_is_applied(context, user=None):
    if "application" in context and "application_type" in context:
        application = context.application
        if context.application_type == "nodejs":
            context.application_original_generation = application.get_observed_generation()
            context.application_original_pod_name = application.get_running_pod_name()
        elif context.application_type == "knative":
            context.application_original_generation = context.application.get_generation()
        else:
            assert False, f"Invalid application type in context.application_type={context.application_type}, valid are 'nodejs', 'knative'"
    if "namespace" in context:
        ns = context.namespace.name
    else:
        ns = None
    resource = substitute_scenario_id(context, context.text)
    binding = ServiceBinding(resource, ns)
    assert binding.create(user) is not None, "Service binding not created"
    context.bindings[binding.name] = binding
    context.sb_secret = ""


@when(u'Invalid Service Binding is applied')
@then(u'Service Binding is unable to be applied')
def invalid_sbr_is_applied(context):
    resource = substitute_scenario_id(context, context.text)
    sbr = ServiceBinding(resource, context.namespace.name)
    # Get resource version of sbr if sbr is available
    if sbr.name in context.bindings.keys():
        context.resource_version = sbr.get_info_by_jsonpath("{.metadata.resourceVersion}")
    context.expected_error = sbr.attempt_to_create_invalid()


@step(u'Service Binding "{sbr_name}" is ready')
@step(u'Service Binding is ready')
def sbo_is_ready(context, sbr_name=None):
    if sbr_name is None:
        sbr_name = list(context.bindings.values())[0].name
    else:
        sbr_name = substitute_scenario_id(context, sbr_name)
    sbo_jq_is(context, '.status.conditions[] | select(.type=="CollectionReady").status', sbr_name, 'True')
    sbo_jq_is(context, '.status.conditions[] | select(.type=="InjectionReady").status', sbr_name, 'True')
    sbo_jq_is(context, '.status.conditions[] | select(.type=="Ready").status', sbr_name, 'True')
    sb = context.bindings[sbr_name]
    if sb.crdName == "servicebindings.servicebinding.io":
        assert sb.get_info_by_jsonpath("{.metadata.generation") == sb.get_info_by_jsonpath("{.status.observedGeneration"), \
            f"Service binding {sb.name} observed generation not equal to generation"
    context.sb_secret = context.bindings[sbr_name].get_secret_name()


# STEP
@step(u'jq "{jq_expression}" of Service Binding "{sbr_name}" should be changed to "{json_value}"')
@step(u'jq "{jq_expression}" of Service Binding should be changed to "{json_value}"')
def sbo_jq_is(context, jq_expression, sbr_name=None, json_value=""):
    if sbr_name is None:
        sbr_name = list(context.bindings.values())[0].name
    else:
        sbr_name = substitute_scenario_id(context, sbr_name)
    json_value = substitute_scenario_id(context, json_value)
    polling2.poll(lambda: json.loads(
        context.bindings[sbr_name].get_info_by_jsonpath(jq_expression)) == json_value,
                  step=5, timeout=800, ignore_exceptions=(json.JSONDecodeError,))


@when(u'Service binding "{sb_name}" is deleted')
def service_binding_is_deleted(context, sb_name):
    resource = substitute_scenario_id(context, sb_name)
    sb = context.bindings[resource]
    context.sb_secret = sb.get_secret_name()
    sb.delete()


@then(u'Service Binding "{sb_name}" is not updated')
def validate_persistent_sb(context, sb_name):
    resource = substitute_scenario_id(context, sb_name)
    json_path = "{.metadata.resourceVersion}"
    assert context.resource_version == context.bindings[resource].get_info_by_jsonpath(json_path), "Service Binding got update"


@step(u'Service Binding "{sbr_name}" has the binding secret name set in the status')
@step(u'Service Binding has the binding secret name set in the status')
def sbo_secret_name_has_been_set(context, sbr_name=None):
    if sbr_name is None:
        sbr_name = list(context.bindings.values())[0].name
    else:
        sbr_name = substitute_scenario_id(context, sbr_name)
    polling2.poll(lambda: context.bindings[sbr_name].get_secret_name() != "",
                  step=5, timeout=800,  ignore_exceptions=(json.JSONDecodeError,))


@step(u'Service Binding {condition}.{field} is "{field_value}"')
def check_sb_condition_field_value(context, condition, field, field_value):
    sb = list(context.bindings.values())[0]
    sbo_jq_is(context, f'.status.conditions[] | select(.type=="{condition}").{field}', sb.name, field_value)
