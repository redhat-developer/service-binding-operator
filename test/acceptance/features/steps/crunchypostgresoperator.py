from subscription_install_mode import InstallMode
from olm import Operator
from environment import ctx
from behave import given


class CrunchyPostgresOperator(Operator):

    def __init__(self, name="pgo"):
        self.name = name
        if ctx.cli == "oc":
            self.operator_catalog_source_name = "certified-operators"
            self.package_name = "crunchy-postgres-operator"
        else:
            self.operator_catalog_source_name = "operatorhubio-catalog"
            self.package_name = "postgresql"
            self.operator_subscription_csv_version = "postgresoperator.v5.0.5"
        self.operator_catalog_channel = "v5"


@given(u'Crunchy Data Postgres operator is running')
def install(_context):
    operator = CrunchyPostgresOperator()
    if not operator.is_running():
        operator.install_operator_subscription(install_mode=InstallMode.Manual)
        operator.is_running(wait=True)
    print("Crunchy Data Postgres operator is running")
