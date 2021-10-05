from olm import Operator
from environment import ctx
from behave import given


class CrunchyPostgresOperator(Operator):

    def __init__(self, name="pgo"):
        self.name = name
        if ctx.cli == "oc":
            self.operator_catalog_source_name = "community-operators"
        else:
            self.operator_catalog_source_name = "operatorhubio-catalog"
        self.operator_catalog_channel = "v5"
        self.package_name = "postgresql"


@given(u'Crunchy Data Postgres operator is running')
def install(_context):
    operator = CrunchyPostgresOperator()
    if not operator.is_running():
        operator.install_operator_subscription()
        operator.is_running(wait=True)
    print("Crunchy Data Postgres operator is running")
