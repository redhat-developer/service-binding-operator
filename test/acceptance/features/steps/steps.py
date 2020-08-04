# @mark.steps
# ----------------------------------------------------------------------------
# STEPS:
# ----------------------------------------------------------------------------
import os
import re

from behave import given, then, when
from pyshould import should, should_not

from servicebindingoperator import Servicebindingoperator
from dboperator import DbOperator
from openshift import Openshift
from postgres_db import PostgresDB
from namespace import Namespace
from nodejs_application import NodeJSApp
from service_binding_request import ServiceBindingRequest
import time


# STEP
@given(u'Namespace "{namespace_name}" is used')
def given_namespace_is_used(context, namespace_name):
    namespace = Namespace(namespace_name)
    if not namespace.is_present():
        print("Namespace is not present, creating namespace: {}...".format(namespace_name))
        namespace.create() | should.be_truthy.desc("Namespace {} is created".format(namespace_name))
    print("Namespace {} is created!!!".format(namespace_name))
    context.namespace = namespace


# STEP
@given(u'Namespace [{namespace_env}] is used')
def given_namespace_from_env_is_used(context, namespace_env):
    env = os.getenv(namespace_env, "")
    env | should_not.be_none.desc(f"{namespace_env} env variable is set")
    print(f"{namespace_env} = {env}")
    given_namespace_is_used(context, env)


# STEP
sbo_is_running_in_namespace_step = u'Service Binding Operator is running in "{operator_namespace}" namespace'


@given(sbo_is_running_in_namespace_step)
@when(sbo_is_running_in_namespace_step)
def sbo_is_running_in_namespace(context, operator_namespace):
    """
    Checks if the SBO is up and running in the given namespace
    """
    sb_operator = Servicebindingoperator(namespace=operator_namespace)
    sb_operator.is_running() | should.be_truthy.desc("Service Binding Operator is running")
    print("Service binding operator is running!!!")


# STEP
sbo_is_running_in_namespace_from_env_step = u'Service Binding Operator is running in [{operator_namespace_env}] namespace'


@given(sbo_is_running_in_namespace_from_env_step)
@when(sbo_is_running_in_namespace_from_env_step)
def sbo_is_running_in_namespace_from_env(context, operator_namespace_env):
    env = os.getenv(operator_namespace_env, "")
    env | should_not.be_none.desc(f"{operator_namespace_env} env variable is set")
    print(f"{operator_namespace_env} = {env}")
    sbo_is_running_in_namespace(context, env)


# STEP
sbo_is_running_step = u'Service Binding Operator is running'


@given(sbo_is_running_step)
@when(sbo_is_running_step)
def sbo_is_running(context):
    context.namespace | should_not.be_none.desc("Namespace set in context")
    sbo_is_running_in_namespace(context, context.namespace.name)


# STEP
@given(u'PostgreSQL DB operator is installed')
def given_db_operator_is_installed(context):
    db_operator = DbOperator()
    if not db_operator.is_running():
        print("DB operator is not installed, installing...")
        db_operator.install_catalog_source() | should.be_truthy.desc("DB catalog source installed")
        db_operator.install_operator_subscription() | should.be_truthy.desc("DB operator subscription installed")
        db_operator.is_running(wait=True) | should.be_truthy.desc("DB operator installed")
    print("PostgresSQL DB operator is running!!!")


# STEP
imported_nodejs_app_is_running_step = u'Imported Nodejs application "{application_name}" is running'


@given(imported_nodejs_app_is_running_step)
@when(imported_nodejs_app_is_running_step)
def imported_nodejs_app_is_running(context, application_name):
    namespace = context.namespace
    application = NodeJSApp(application_name, namespace.name)
    if not application.is_running():
        print("application is not running, trying to import it")
        application.install() | should.be_truthy.desc("Application is installed")
        application.is_running(wait=True) | should.be_truthy.desc("Application is running")
    print("Nodejs application is running!!!")
    application.get_db_name_from_api() | should_not.be_none
    context.nodejs_app_original_generation = application.get_observed_generation()
    context.nodejs_app_original_pod_name = application.get_running_pod_name()
    context.nodejs_app = application


# STEP
imported_nodejs_app_is_not_running_step = u'Imported Nodejs application "{application_name}" is not running'


@given(imported_nodejs_app_is_not_running_step)
@when(imported_nodejs_app_is_not_running_step)
def imported_nodejs_app_is_not_running(context, application_name):
    namespace = context.namespace
    application = NodeJSApp(application_name, namespace.name)
    application.is_running() | should.be_falsy.desc("Aplication not running")


# STEP
db_instance_is_running_step = u'DB "{db_name}" is running'


@given(db_instance_is_running_step)
@when(db_instance_is_running_step)
def db_instance_is_running(context, db_name):
    namespace = context.namespace

    db = PostgresDB(db_name, namespace.name)
    if not db.is_running():
        db.create() | should.be_truthy.desc("Postgres DB created")
        db.is_running(wait=True) | should.be_truthy.desc("Postgres DB is running")
    print(f"DB {db_name} is running!!!")


# STEP
sbr_is_applied_step = u'Service Binding Request is applied to connect the database and the application'


@given(sbr_is_applied_step)
@when(sbr_is_applied_step)
def sbr_is_applied(context):
    sbr_yaml = context.text
    sbr = ServiceBindingRequest()
    if context.__contains__("nodejs_app"):
        application = context.nodejs_app
        context.nodejs_app_original_generation = application.get_observed_generation()
        context.nodejs_app_original_pod_name = application.get_running_pod_name()
    sbr.create(sbr_yaml) | should.be_truthy.desc("Service Binding Request Created")


# STEP
@then(u'application should be re-deployed')
def then_application_redeployed(context):
    application = context.nodejs_app
    application.get_redeployed_pod_name(context.nodejs_app_original_pod_name) | should_not.be_none.desc(
        "There is a running pod of the application different from the original one before redeployment.")


# STEP
@then(u'application should be connected to the DB "{db_name}"')
def then_app_is_connected_to_db(context, db_name):
    application = context.nodejs_app
    app_db_name = application.get_db_name_from_api()
    app_db_name | should.be_equal_to(db_name)


# STEP
@then(u'jsonpath "{json_path}" of Service Binding Request "{sbr_name}" should be changed to "{json_value_regex}"')
def then_sbo_jsonpath_is(context, json_path, sbr_name, json_value_regex):
    openshift = Openshift()
    openshift.search_resource_in_namespace("servicebindingrequests", sbr_name, context.namespace.name) | should_not.be_none.desc("SBR {sbr_name} exists")
    result = openshift.get_resource_info_by_jsonpath("sbr", sbr_name, context.namespace.name, json_path, wait=True)
    result | should_not.be_none.desc("jsonpath result")
    re.fullmatch(json_value_regex, result) | should_not.be_none.desc("SBO jsonpath result \"{result}\" should match \"{json_value_regex}\"")


# STEP
@then(u'jq "{jq_expression}" of Service Binding Request "{sbr_name}" should be changed to "{json_value_regex}"')
def then_sbo_jq_is(context, jq_expression, sbr_name, json_value_regex):
    openshift = Openshift()
    openshift.search_resource_in_namespace("servicebindingrequests", sbr_name, context.namespace.name) | should_not.be_none.desc("SBR {sbr_name} exists")
    result = openshift.get_resource_info_by_jq("sbr", sbr_name, context.namespace.name, jq_expression, wait=True)
    result | should_not.be_none.desc("jq result")
    re.fullmatch(json_value_regex, result) | should_not.be_none.desc("SBO jq result \"{result}\" should match \"{json_value_regex}\"")


# STEP
@then(u'"{app_name}" deployment must contain SBR name "{sbr_name1}" and "{sbr_name2}"')
def then_envFrom_contains(context, app_name, sbr_name1, sbr_name2):
    time.sleep(60)
    openshift = Openshift()
    result = openshift.get_deployment_envFrom_info(app_name, context.namespace.name)
    result | should.be_equal_to("[map[secretRef:map[name:binding-request-1]] map[secretRef:map[name:binding-request-2]]]")\
        .desc(f'{app_name} deployment should contain secretRef: {sbr_name1} and {sbr_name2}')
