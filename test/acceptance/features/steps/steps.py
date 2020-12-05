# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
import ipaddress
import os
import json
import re
import time
import polling2
import parse
import binascii
import yaml

from string import Template
from behave import given, register_type, then, when, step
from dboperator import DbOperator
from etcdcluster import EtcdCluster
from etcdoperator import EtcdOperator
from knative_serving import KnativeServing
from namespace import Namespace
from nodejs_application import NodeJSApp
from openshift import Openshift
from postgres_db import PostgresDB
from quarkus_application import QuarkusApplication
from serverless_operator import ServerlessOperator
from service_binding import ServiceBinding
from servicebindingoperator import Servicebindingoperator
from app import App
from util import scenario_id


# STEP
@given(u'Namespace "{namespace_name}" exists')
def namespace_maybe_create(context, namespace_name):
    namespace = Namespace(namespace_name)
    if not namespace.is_present():
        print("Namespace is not present, creating namespace: {}...".format(namespace_name))
        assert namespace.create(), f"Unable to create namespace '{namespace_name}'"
    print("Namespace {} is created!!!".format(namespace_name))
    return namespace


@given(u'Namespace "{namespace_name}" is used')
def namespace_is_used(context, namespace_name):
    context.namespace = namespace_maybe_create(context, namespace_name)


# STEP
@given(u'Namespace [{namespace_env}] is used')
def given_namespace_from_env_is_used(context, namespace_env):
    env = os.getenv(namespace_env)
    assert env is not None, f"{namespace_env} environment variable needs to be set"
    print(f"{namespace_env} = {env}")
    namespace_is_used(context, env)


# STEP
sbo_is_running_in_namespace_step = u'Service Binding Operator is running in "{operator_namespace}" namespace'


@given(sbo_is_running_in_namespace_step)
@when(sbo_is_running_in_namespace_step)
def sbo_is_running_in_namespace(context, operator_namespace):
    """
    Checks if the SBO is up and running in the given namespace
    """
    sb_operator = Servicebindingoperator(namespace=operator_namespace)
    assert sb_operator.is_running(), "Service Binding Operator is not running"
    print("Service binding operator is running!!!")


# STEP
sbo_is_running_in_namespace_from_env_step = u'Service Binding Operator is running in [{operator_namespace_env}] namespace'


@given(sbo_is_running_in_namespace_from_env_step)
@when(sbo_is_running_in_namespace_from_env_step)
def sbo_is_running_in_namespace_from_env(context, operator_namespace_env):
    env = os.getenv(operator_namespace_env)
    assert env is not None, f"{operator_namespace_env} environment variable needs to be set"
    print(f"{operator_namespace_env} = {env}")
    sbo_is_running_in_namespace(context, env)


# STEP
sbo_is_running_step = u'Service Binding Operator is running'


@given(sbo_is_running_step)
@when(sbo_is_running_step)
def sbo_is_running(context):
    if "sbo_namespace" in context:
        sbo_is_running_in_namespace(context, context.sbo_namespace)
    else:
        assert context.namespace is not None, "Namespace is not set in context"
        sbo_is_running_in_namespace(context, context.namespace.name)


# STEP
@given(u'PostgreSQL DB operator is installed')
def given_db_operator_is_installed(context):
    db_operator = DbOperator()
    if not db_operator.is_running():
        print("DB operator is not installed, installing...")
        assert db_operator.install_catalog_source(), "Unable to install DB catalog source"
        assert db_operator.install_operator_subscription(), "Unable to install DB operator subscription"
        assert db_operator.is_running(wait=True), "Unable to launch DB operator"
    print("PostgresSQL DB operator is running!!!")


# STEP
nodejs_app_imported_from_image_is_running_step = u'Nodejs application "{application_name}" imported from "{application_image}" image is running'


@given(nodejs_app_imported_from_image_is_running_step)
@when(nodejs_app_imported_from_image_is_running_step)
def nodejs_app_imported_from_image_is_running(context, application_name, application_image):
    namespace = context.namespace
    application = NodeJSApp(application_name, namespace.name, application_image)
    if not application.is_running():
        print("application is not running, trying to import it")
        assert application.install(), f"Unable to install application '{application_name}' from image '{application_image}'"
        assert application.is_running(wait=True), f"Unable to start application '{application_name}' from image '{application_image}'"
    print("Nodejs application is running!!!")
    context.application = application
    context.application_type = "nodejs"


# STEP
app_endpoint_is_available_step = u'Application endpoint "{endpoint}" is available'


@given(app_endpoint_is_available_step)
@when(app_endpoint_is_available_step)
@then(app_endpoint_is_available_step)
def app_endpoint_is_available(context, endpoint):
    application = context.application
    assert application.get_response_from_api(endpoint=endpoint) is not None, f'Application endpoint "{endpoint}" is not available'


# STEP
default_nodejs_app_imported_from_image_is_running_step = u'Imported Nodejs application "{application_name}" is running'


@given(default_nodejs_app_imported_from_image_is_running_step)
@when(default_nodejs_app_imported_from_image_is_running_step)
def default_nodejs_app_imported_from_image_is_running(context, application_name):
    nodejs_app_imported_from_image_is_running(context, application_name, application_image="quay.io/pmacik/nodejs-rest-http-crud")
    app_endpoint_is_available(context, endpoint="/api/status/dbNameCM")


# STEP
imported_nodejs_app_is_not_running_step = u'Imported Nodejs application "{application_name}" is not running'


@given(imported_nodejs_app_is_not_running_step)
@when(imported_nodejs_app_is_not_running_step)
def imported_nodejs_app_is_not_running(context, application_name):
    namespace = context.namespace
    application = NodeJSApp(application_name, namespace.name)
    assert application.is_running() is False, "Application is running already"


# STEP
db_instance_is_running_step = u'DB "{db_name}" is running'


@given(db_instance_is_running_step)
@when(db_instance_is_running_step)
@given(u'DB "{db_name}" is running in "{namespace}" namespace')
def db_instance_is_running(context, db_name, namespace=None):
    if namespace is None:
        namespace = context.namespace.name

    db = PostgresDB(db_name, namespace)
    if not db.is_running():
        assert db.create(), f"Unable to create DB '{db_name}'"
        assert db.is_running(wait=True), f"Unable to launch DB '{db_name}'"
    print(f"DB {db_name} is running!!!")


# STEP
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
    context.sbr_name = apply_yaml(context, user)["name"]


# STEP
@then(u'application should be re-deployed')
def then_application_redeployed(context):
    application = context.application
    if context.application_type == "nodejs":
        application_pod_name = application.get_redeployed_pod_name(context.application_original_pod_name)
        assert application_pod_name is not None, "There is no running application pod different from the original one before re-deployment."
    elif context.application_type == "knative":
        assert context.application_original_generation is not None, "application is never deployed"
        application_rev_name = application.get_rev_name_redeployed_by_generation(context.application_original_generation)
        assert application_rev_name is not None, "application is not redeployed"
    else:
        assert False, f"Invalid application type in context.application_type={context.application_type}, valid are 'nodejs', 'knative'"


# STEP
@then(u'application should be connected to the DB "{db_name}"')
def then_app_is_connected_to_db(context, db_name):
    db_endpoint = "/api/status/dbNameCM"
    polling2.poll(lambda: context.application.get_response_from_api(endpoint=db_endpoint) == db_name, step=5, timeout=600)


@when(u'Service binding "{sb_name}" is deleted')
def service_binding_is_deleted(context, sb_name):
    openshift = Openshift()
    openshift.delete_service_binding(sb_name, context.namespace.name)


# STEP
@given(u'jsonpath "{json_path}" of Service Binding "{sbr_name}" should be changed to "{json_value_regex}"')
@then(u'jsonpath "{json_path}" of Service Binding "{sbr_name}" should be changed to "{json_value_regex}"')
def then_sbo_jsonpath_is(context, json_path, sbr_name, json_value_regex):
    openshift = Openshift()
    assert openshift.search_resource_in_namespace(
        "servicebindings", sbr_name, context.namespace.name) is not None, f"Service Binding '{sbr_name}' does not exist in namespace '{context.namespace.name}'"
    result = openshift.get_resource_info_by_jsonpath("sbr", sbr_name, context.namespace.name, json_path)
    assert result is not None, f"Invalid result for SBO jsonpath: {result}."
    assert re.fullmatch(json_value_regex, result) is not None, f"SBO jsonpath result \"{result}\" does not match \"{json_value_regex}\""


# STEP
@step(u'jq "{jq_expression}" of Service Binding "{sbr_name}" should be changed to "{json_value}"')
def sbo_jq_is(context, jq_expression, sbr_name, json_value):
    openshift = Openshift()
    polling2.poll(lambda: json.loads(
        openshift.get_resource_info_by_jq("servicebinding", sbr_name, context.namespace.name, jq_expression,
                                          wait=False)) == json_value, step=5, timeout=600,
                  ignore_exceptions=(json.JSONDecodeError,))


@step(u'Service Binding {condition}.{field} is "{field_value}"')
def check_sb_condition_field_value(context, condition, field, field_value):
    sbo_jq_is(context, f'.status.conditions[] | select(.type=="{condition}").{field}', context.sbr_name, field_value)


@step(u'Service Binding "{sbr_name}" is ready')
def sbo_is_ready(context, sbr_name):
    sbo_jq_is(context, '.status.conditions[] | select(.type=="CollectionReady").status', sbr_name, 'True')
    sbo_jq_is(context, '.status.conditions[] | select(.type=="InjectionReady").status', sbr_name, 'True')
    sbo_jq_is(context, '.status.conditions[] | select(.type=="Ready").status', sbr_name, 'True')


@given(u'Openshift Serverless Operator is running')
def given_serverless_operator_is_running(context):
    """
    Checks if the serverless operator is up and running
    """
    serverless_operator = ServerlessOperator()
    if not serverless_operator.is_running():
        print("Serverless operator is not installed, installing...")
        assert serverless_operator.install_operator_subscription() is True, "serverless operator subscription is not installed"
        assert serverless_operator.is_running(wait=True) is True, "serverless operator not installed"
    context.serverless_operator = serverless_operator


@given(u'Knative serving is running')
def given_knative_serving_is_running(context):
    """
    creates a KnativeServing object to install Knative Serving using the OpenShift Serverless Operator.
    """
    knative_namespace = Namespace("knative-serving")
    assert knative_namespace.create() is True, "Knative serving namespace not created"
    knative_serving = KnativeServing(namespace=knative_namespace.name)
    if not knative_serving.is_present():
        print("knative serving is not present, create knative serving")
        assert knative_serving.create() is True, "Knative serving is not created"
        assert knative_serving.is_present() is True, "Knative serving is not present"


# STEP
quarkus_app_is_imported_step = u'Quarkus application "{application_name}" is imported as Knative service'


@given(quarkus_app_is_imported_step)
@when(quarkus_app_is_imported_step)
def quarkus_app_is_imported_as_knative_service(context, application_name):
    namespace = context.namespace
    application = QuarkusApplication(application_name, namespace.name)
    if not application.is_imported():
        print("application is not imported, trying to import it")
        assert application.install() is True, "Quarkus application is not installed"
        assert application.is_imported(wait=True) is True, "Quarkus application is not imported"
    context.application = application
    context.application_type = "knative"


# STEP


@then(u'"{app_name}" deployment must contain SBR name "{sbr_name}"')
def then_envFrom_contains(context, app_name, sbr_name):
    time.sleep(60)
    openshift = Openshift()
    result = openshift.get_deployment_envFrom_info(app_name, context.namespace.name)
    # Expected result from 'oc' (openshift client) v4.5
    expected_result_oc_45 = f'secretRef:map[name:{sbr_name}]'
    # Expected result from 'oc' (openshift client) v4.6+
    expected_result_oc_46 = f'{{"secretRef":{{"name":"{sbr_name}"}}}}'
    assert re.search(re.escape(expected_result_oc_45), result) is not None or re.search(re.escape(expected_result_oc_46), result) is not None, \
        f'\n{app_name} deployment should contain secretRef: {sbr_name} \nActual secretRef: {result}'


# STEP
@then(u'deployment must contain intermediate secret "{intermediate_secret_name}"')
def envFrom_contains_intermediate_secret_name(context, intermediate_secret_name):
    assert context.application.get_deployment_with_intermediate_secret(
        intermediate_secret_name) is not None, f"There is no deployment with intermediate secret {intermediate_secret_name}"


# STEP
@given(u'OLM Operator "{backend_service}" is running')
def operator_manifest_installed(context, backend_service):
    openshift = Openshift()
    if "namespace" in context:
        ns = context.namespace.name
    else:
        ns = None
    _ = openshift.apply_yaml_file(os.path.join(os.getcwd(), "test/acceptance/resources/", backend_service + ".operator.manifest.yaml"), namespace=ns)


@parse.with_pattern(r'.*')
def parse_nullable_string(text):
    return text


register_type(NullableString=parse_nullable_string)


# STEP
@step(u'Secret "{secret_name}" contains "{secret_key}" key with value "{secret_value:NullableString}"')
def check_secret_key_value(context, secret_name, secret_key, secret_value):
    openshift = Openshift()
    json_path = f'{{.data.{secret_key}}}'
    polling2.poll(lambda: openshift.get_resource_info_by_jsonpath("secrets", secret_name, context.namespace.name,
                                                                  json_path) == secret_value,
                  step=5, timeout=120, ignore_exceptions=(binascii.Error,))


# STEP
@then(u'Secret "{secret_name}" contains "{secret_key}" key with dynamic IP addess as the value')
def check_secret_key_with_ip_value(context, secret_name, secret_key):
    openshift = Openshift()
    json_path = f'{{.data.{secret_key}}}'
    polling2.poll(lambda: ipaddress.ip_address(
        openshift.get_resource_info_by_jsonpath("secrets", secret_name, context.namespace.name, json_path)),
        step=5, timeout=120, ignore_exceptions=(ValueError,))


# STEP
@given(u'The openshift route is present')
@given(u'Namespace is present')
@given(u'Backend service CSV is installed')
@given(u'The Custom Resource Definition is present')
@given(u'The Custom Resource is present')
@when(u'The Custom Resource is present')
@given(u'The ConfigMap is present')
@given(u'The Secret is present')
@when(u'The Secret is present')
def apply_yaml(context, user=None):
    openshift = Openshift()
    resource = Template(context.text).substitute(scenario_id=scenario_id(context))
    metadata = yaml.full_load(resource)["metadata"]
    metadata_name = metadata["name"]
    if "namespace" in metadata:
        ns = metadata["namespace"]
    else:
        if "namespace" in context:
            ns = context.namespace.name
        else:
            ns = None
    output = openshift.apply(resource, ns, user)
    result = re.search(rf'.*{metadata_name}.*(created|unchanged|configured)', output)
    assert result is not None, f"Unable to apply YAML for CR '{metadata_name}': {output}"
    return metadata


# STEP
@given(u'BackingService is deleted')
@when(u'BackingService is deleted')
def delete_yaml(context):
    openshift = Openshift()
    metadata = yaml.full_load(context.text)["metadata"]
    metadata_name = metadata["name"]
    if "namespace" in metadata:
        ns = metadata["namespace"]
    else:
        if "namespace" in context:
            ns = context.namespace.name
        else:
            ns = None
    output = openshift.delete(context.text, ns)
    result = re.search(rf'.*{metadata_name}.*(deleted)', output)
    assert result is not None, f"Unable to delete CR '{metadata_name}': {output}"


@then(u'Secret "{secret_ref}" has been injected in to CR "{cr_name}" of kind "{crd_name}" at path "{json_path}"')
def verify_injected_secretRef(context, secret_ref, cr_name, crd_name, json_path):
    openshift = Openshift()
    polling2.poll(lambda: openshift.get_resource_info_by_jsonpath(crd_name, cr_name, context.namespace.name, json_path) == secret_ref,
                  step=5, timeout=400)


@given(u'Etcd operator running')
def etcd_operator_is_running(context):
    """
    Checks if the etcd operator is up and running
    """
    etcd_operator = EtcdOperator()
    if not etcd_operator.is_running():
        print("Etcd operator is not installed, installing...")
        assert etcd_operator.install_operator_subscription() is True, "etcd operator subscription is not installed"
        assert etcd_operator.is_running(wait=True) is True, "etcd operator not installed"
    context.etcd_operator = etcd_operator


@given(u'Etcd cluster "{etcd_name}" is running')
def etc_cluster_is_running(context, etcd_name):
    """
    Checks if the etcd cluster is created
    """
    namespace = context.namespace
    etcd_cluster = EtcdCluster(etcd_name, namespace.name)
    if not etcd_cluster.is_present():
        print("etcd cluster not present, creating etcd cluster")
        assert etcd_cluster.create() is True, "etcd cluster is not created"
        assert etcd_cluster.is_present() is True, "etcd cluster is not present"


@when(u'Invalid Service Binding is applied')
def invalid_sbr_is_applied(context):
    sbr = ServiceBinding()
    # Get resource version of sbr if sbr is available
    if "sbr_name" in context:
        json_path = "{.metadata.resourceVersion}"
        rv = sbr.get_servicebinding_info_by_jsonpath(context.sbr_name, context.namespace.name, json_path)
        context.resource_version = rv
    context.expected_error = sbr.attempt_to_create_invalid(context.text, context.namespace.name)


@then(u'Error message is thrown')
@then(u'Error message "{err_msg}" is thrown')
def validate_error(context, err_msg=None):
    if err_msg is None:
        assert context.expected_error is not None, "An error message should happen"
    else:
        search = re.search(rf'.*{err_msg}.*', context.expected_error)
        assert search is not None, f"Actual error: '{context.expected_error}', Expected error: '{err_msg}'"


@then(u'Service Binding "{sb_name}" is not persistent in the cluster')
def validate_absent_sb(context, sb_name):
    openshift = Openshift()
    output = openshift.search_resource_in_namespace("servicebindings", sb_name, context.namespace.name)
    assert output is None, f"Service Binding {sb_name} is present in namespace '{context.namespace.name}'"


@then(u'Service Binding "{sb_name}" is not updated')
def validate_persistent_sb(context, sb_name):
    openshift = Openshift()
    json_path = "{.metadata.resourceVersion}"
    output = openshift.get_resource_info_by_jsonpath("servicebindings", sb_name, context.namespace.name, json_path)
    if output == context.resource_version:
        assert True
    else:
        assert False, "Service Binding got updated"


@given(u'Secret "{secret_name}" does not contain "{key}"')
def check_secret_key(context, secret_name, key):
    openshift = Openshift()
    json_path = f'{{.data.{key}}}'
    polling2.poll(lambda: openshift.get_resource_info_by_jsonpath("secrets", secret_name, context.namespace.name,
                                                                  json_path) == "",
                  step=5, timeout=120, ignore_exceptions=(binascii.Error,))


@step(u'Secret "{secret_name}" is empty')
def validate_secret_empty(context, secret_name):
    openshift = Openshift()
    try:
        polling2.poll(lambda: json.loads(
            openshift.get_resource_info_by_jq("secrets", secret_name, context.namespace.name, ".data",
                                              wait=False)) == "null", step=5, timeout=20,
                      ignore_exceptions=(json.JSONDecodeError,))
    except polling2.TimeoutException:
        pass


@given(u'"{app_name}" is deployed from image "{image_ref}"')
def create_deployment(context, app_name, image_ref):
    app = App(app_name, context.namespace.name, image_ref)
    if not app.is_running():
        assert app.install() is True, "Failed to create deployment."


@step(u'User {user} cannot read resource {resource} in namespace {namespace}')
@step(u'User {user} cannot read resource {resource} in test namespace')
def check_resource_unreadable(context, user, resource, namespace=None):
    openshift = Openshift()
    data = resource.split("/")
    res_type = data[0]
    res_name = Template(data[1]).substitute(scenario_id=scenario_id(context))
    ns = namespace if namespace is not None else context.namespace.name
    res = openshift.get_resource_info_by_jsonpath(resource_type=res_type, name=res_name, namespace=ns, user=user)
    assert res is None, f"User {user} should not be able to read {resource} in {namespace} namespace"


@step(u"User {user} has '{role_name}' role in test namespace")
def user_role_maybe_create(context, user, role_name):
    openshift = Openshift()
    openshift.cli(f"create rolebinding {user}-{role_name} --clusterrole={role_name} --user={user}", context.namespace.name)


@step(u'No user has access to the namespace')
def remove_user_rolebindings(context):
    openshift = Openshift()
    rbs = openshift.get_resource_lst("rolebindings", context.namespace.name)
    for rb in rbs:
        if rb != "service-binding-operator":
            openshift.cli(f"delete rolebinding {rb}", context.namespace.name)
