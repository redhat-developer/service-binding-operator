from olm import Operator
from environment import ctx
from behave import step


class CloudNativePostgresOperator(Operator):

    def __init__(self, name="cloud-native-postgresql"):
        super().__init__(
            name=name,
            pod_name_pattern="postgresql-operator-controller-manager.*",
            operator_catalog_source_name="operatorhubio-catalog",
            operator_catalog_channel="stable",
            operator_catalog_image="quay.io/operatorhubio/catalog:latest",
            package_name=name)


@step(u'Cloud Native Postgres operator is running')
def install(_context):
    operator = CloudNativePostgresOperator()
    if not operator.is_running():
        if ctx.cli == "oc":
            operator.install_catalog_source()
        operator.install_operator_subscription()
        operator.is_running(wait=True)
    print("Cloud Native Postgres operator is running")


@step(u'Cloud Native Postgres operator is removed')
def uninstall(_context):
    operator = CloudNativePostgresOperator()
    if operator.is_running():
        operator.uninstall_operator_subscription(wait=True)
