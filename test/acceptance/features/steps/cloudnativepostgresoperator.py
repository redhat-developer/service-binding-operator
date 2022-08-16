from olm import Operator
from environment import ctx
from behave import given


class CloudNativePostgresOperator(Operator):

    def __init__(self, name="cloud-native-postgresql"):
        super().__init__(
            name=name,
            pod_name_pattern="postgresql-operator-controller-manager.*",
            operator_catalog_source_name="operatorhubio-catalog",
            operator_catalog_channel="stable",
            operator_catalog_image="quay.io/operatorhubio/catalog:latest",
            package_name=name)


@given(u'Cloud Native Postgres operator is running')
def install(_context):
    operator = CloudNativePostgresOperator()
    if not operator.is_running():
        if ctx.cli == "oc":
            operator.install_catalog_source()
        operator.install_operator_subscription()
        operator.is_running(wait=True)
    print("Cloud Native Postgres operator is running")
