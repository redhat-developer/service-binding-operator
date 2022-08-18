from subscription_install_mode import InstallMode
from olm import Operator
from environment import ctx
from behave import given


class CrunchyPostgresOperator(Operator):

    def __init__(self, name="pgo"):
        if ctx.cli == "oc":
            package_name = "crunchy-postgres-operator"
            catalog_source_name = "certified-operators"
        else:
            package_name = "postgresql"
            catalog_source_name = "operatorhubio-catalog"
        csv = "postgresoperator.v5.1.2"
        channel = "v5"

        super().__init__(
            name=name,
            package_name=package_name,
            operator_catalog_source_name=catalog_source_name,
            operator_subscription_csv_version=csv,
            operator_catalog_channel=channel)


@given(u'Crunchy Data Postgres operator is running')
def install(_context):
    operator = CrunchyPostgresOperator()
    if not operator.is_running():
        operator.install_operator_subscription(install_mode=InstallMode.Manual)
        operator.is_running(wait=True)
    print("Crunchy Data Postgres operator is running")
