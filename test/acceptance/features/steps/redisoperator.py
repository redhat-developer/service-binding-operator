from olm import Operator
from environment import ctx
from behave import given


class RedisOperator(Operator):

    def __init__(self, name="redis-operator"):
        self.name = name
        if ctx.cli == "oc":
            self.operator_catalog_source_name = "community-operators"
        else:
            self.operator_catalog_source_name = "operatorhubio-catalog"
        self.operator_catalog_channel = "stable"
        self.package_name = name


@given(u'Opstree Redis operator is running')
def install_redis_operator(_context):
    operator = RedisOperator()
    if not operator.is_running():
        operator.install_operator_subscription()
        operator.is_running(wait=True)
    print("Opstree Redis operator is running")
