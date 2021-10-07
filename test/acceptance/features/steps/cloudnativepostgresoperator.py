from olm import Operator
from environment import ctx
from behave import given


class CloudNativePostgresOperator(Operator):

    def __init__(self, name="cloud-native-postgresql"):
        self.name = name
        if ctx.cli == "oc":
            self.operator_catalog_source_name = "certified-operators"
        else:
            self.operator_catalog_source_name = "operatorhubio-catalog"
        self.operator_catalog_channel = "stable"
        self.package_name = name


@given(u'Cloud Native Postgres operator is running')
def install(_context):
    operator = CloudNativePostgresOperator()
    if not operator.is_running():
        operator.install_operator_subscription()
        operator.is_running(wait=True)
    print("Cloud Native Postgres operator is running")
