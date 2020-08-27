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
from serverless_operator import ServerlessOperator
from quarkus_application import QuarkusApplication
from quarkus_s2i_builder_image import QuarkusS2IBuilderImage
from knative_serving import KnativeServing
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
    env = os.getenv(namespace_env)
    assert env is not None, f"{namespace_env} environment variable needs to be set"
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
    env = os.getenv(operator_namespace_env)
    assert env is not None, f"{operator_namespace_env} environment variable needs to be set"
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
    context.application = application
    context.application_type = "nodejs"


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
    if context.__contains__("application") and context.__contains__("application_type"):
        application = context.application
        if context.application_type == "nodejs":
            context.application_original_generation = application.get_observed_generation()
            context.application_original_pod_name = application.get_running_pod_name()
        elif context.application_type == "knative":
            context.application_original_generation = context.application.get_generation()
        else:
            assert False, f"Invalid application type in context.application_type={context.application_type}, valid are 'nodejs', 'knative'"
    assert sbr.create(sbr_yaml) is not None, "Service binding request not created"


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
    application = context.application
    app_db_name = application.get_db_name_from_api(wait=True)
    app_db_name | should.be_equal_to(db_name)


# STEP
@then(u'jsonpath "{json_path}" of Service Binding Request "{sbr_name}" should be changed to "{json_value_regex}"')
def then_sbo_jsonpath_is(context, json_path, sbr_name, json_value_regex):
    openshift = Openshift()
    openshift.search_resource_in_namespace("servicebindingrequests", sbr_name, context.namespace.name) | should_not.be_none.desc("SBR {sbr_name} exists")
    result = openshift.get_resource_info_by_jsonpath("sbr", sbr_name, context.namespace.name, json_path, wait=True, timeout=600)
    result | should_not.be_none.desc("jsonpath result")
    re.fullmatch(json_value_regex, result) | should_not.be_none.desc("SBO jsonpath result \"{result}\" should match \"{json_value_regex}\"")


# STEP
@then(u'jq "{jq_expression}" of Service Binding Request "{sbr_name}" should be changed to "{json_value_regex}"')
def then_sbo_jq_is(context, jq_expression, sbr_name, json_value_regex):
    openshift = Openshift()
    openshift.search_resource_in_namespace("servicebindingrequests", sbr_name, context.namespace.name) | should_not.be_none.desc("SBR {sbr_name} exists")
    result = openshift.get_resource_info_by_jq("sbr", sbr_name, context.namespace.name, jq_expression, wait=True, timeout=600)
    result | should_not.be_none.desc("jq result")
    re.fullmatch(json_value_regex, result) | should_not.be_none.desc("SBO jq result \"{result}\" should match \"{json_value_regex}\"")


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


@given(u'Quarkus s2i builder image is present')
def given_quarkus_builder_image_is_present(context):
    """
    Checks if quarkus s2i builder image is present
    """
    builder_image = QuarkusS2IBuilderImage()
    if not builder_image.is_present():
        print("Builder image is not present, importing and patching...")
        assert builder_image.import_and_patch() is True, "Quarkus image import from image registry and patch failed"
        assert builder_image.is_present() is True, "Quarkus image is not present"


@given(u'Knative serving is running')
def given_knative_serving_is_running(context):
    """
    creates a KnativeServing object to install Knative Serving using the OpenShift Serverless Operator.
    """
    knative_namespace = Namespace("knative-serving")
    assert knative_namespace.create() is True, "Knative serving namespace not created"
    assert Namespace(context.namespace.name).switch_to() is True, "Unable to switch to the context namespace"
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
        assert application.is_imported() is True, "Quarkus application is not imported"
    context.application = application
    context.application_type = "knative"


# STEP
@then(u'"{app_name}" deployment must contain SBR name "{sbr_name1}" and "{sbr_name2}"')
def then_envFrom_contains(context, app_name, sbr_name1, sbr_name2):
    time.sleep(60)
    openshift = Openshift()
    result = openshift.get_deployment_envFrom_info(app_name, context.namespace.name)
    result | should.be_equal_to("[map[secretRef:map[name:binding-request-1]] map[secretRef:map[name:binding-request-2]]]")\
        .desc(f'{app_name} deployment should contain secretRef: {sbr_name1} and {sbr_name2}')


# STEP
@then(u'deployment must contain intermediate secret "{intermediate_secret_name}"')
def then_envFrom_contains_intermediate_secret_name(context, intermediate_secret_name):
    assert context.application.get_deployment_with_intermediate_secret(
        intermediate_secret_name, wait=True, timeout=120) is not None, f"There is no deployment with intermediate secret {intermediate_secret_name}"


@given(u'The CRD "{crd_name}" is present')
def create_crd(context, crd_name):
    openshift = Openshift()
    yaml = context.text
    output = openshift.oc_apply(yaml)
    result = re.search(rf'.*customresourcedefinition.apiextensions.k8s.io/{crd_name}.*(created|unchanged)', output)
    result | should_not.be_none.desc("CRD {crd_name} Created")


# STEP
@given(u'The application CR "{cr_name}" is present')
def create_cr(context, cr_name):
    openshift = Openshift()
    yaml = context.text
    output = openshift.oc_apply(yaml)
    result = re.search(rf'.*{cr_name}.*(created|unchanged)', output)
    result | should_not.be_none.desc("CRD {cr_name} Created")


@then(u'Secret "{secret_ref}" has been injected in to CR "{cr_name}" of kind "{crd_name}" at path "{json_path}"')
def verify_injected_secretRef(context, secret_ref, cr_name, crd_name, json_path):
    time.sleep(60)
    openshift = Openshift()
    result = openshift.get_resource_info_by_jsonpath(crd_name, cr_name, context.namespace.name, json_path, wait=True, timeout=180)
    result | should.be_equal_to(secret_ref).desc(f'Failed to inject secretRef "{secret_ref}" in "{cr_name}" at path "{json_path}"')
