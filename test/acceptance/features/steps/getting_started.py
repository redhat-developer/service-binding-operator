import os
from behave import step
from command import Command
from openshift import Openshift
from environment import ctx

openshift = Openshift()
cmd = Command()


def yaml_is_applied(context, yaml_file=None, namespace=None):
    _ = openshift.apply_yaml_file(os.path.join(os.getcwd(), yaml_file), namespace=context.namespace.name)


@step(u'PetClinic sample application is installed')
def app_is_installed(context):
    yaml_is_applied(context, yaml_file="samples/apps/spring-petclinic/petclinic-deployment.yaml")


@step(u'PostgreSQL database is running')
def postgresql_is_running(context):
    yaml_is_applied(context, yaml_file="samples/apps/spring-petclinic/postgresql-deployment.yaml")
    output, exit_code = cmd.run(
        f"{ctx.cli} wait --for=condition=Available=True deployment/spring-petclinic-postgresql -n {context.namespace.name} --timeout=300s")
    assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned when attempting to deploy PostgreSQL database for quickstart:\n {output}"


@step(u'PostgresCluster database is running')
def pgcluster_is_running(context):
    yaml_is_applied(context, yaml_file="samples/apps/spring-petclinic/pgcluster-deployment.yaml")
    output, exit_code = cmd.run(f"{ctx.cli} wait --for=condition=PGBackRestReplicaCreate=True PostgresCluster/hippo -n {context.namespace.name} --timeout=300s")
    assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned when attempting to deploy PostgreSQL database for quickstart:\n {output}"


@step(u'PetClinic sample application is running')
def app_is_running(context):
    output, exit_code = cmd.run(f"{ctx.cli} wait --for=condition=Available=True deployment/spring-petclinic -n {context.namespace.name} --timeout=300s")
    assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned when attempting to run PetClinic application for quickstart:\n {output}"
