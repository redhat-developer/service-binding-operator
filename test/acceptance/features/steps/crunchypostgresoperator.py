from subscription_install_mode import InstallMode
from olm import Operator
from environment import ctx
from behave import given


class CrunchyPostgresOperator(Operator):

    def __init__(self, name="pgo"):
        if ctx.cli == "oc":
            super().__init__(
                name=name,
                package_name="crunchy-postgres-operator",
                operator_catalog_source_name="certified-operators",
                operator_catalog_channel="v5")
        else:
            super().__init__(
                name=name,
                package_name="postgresql",
                operator_catalog_source_name="operatorhubio-catalog",
                operator_subscription_csv_version="postgresoperator.v5.0.5",
                operator_catalog_channel="v5")


@given(u'Crunchy Data Postgres operator is running')
def install(_context):
    operator = CrunchyPostgresOperator()
    if not operator.is_running():
        operator.install_operator_subscription(install_mode=InstallMode.Manual)
        operator.is_running(wait=True)
    print("Crunchy Data Postgres operator is running")
