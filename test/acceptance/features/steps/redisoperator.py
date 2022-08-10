from olm import Operator
from environment import ctx
from behave import given


class RedisOperator(Operator):

    def __init__(self, name="redis-operator"):
        super().__init__(
            name=name,
            operator_catalog_source_name="community-operators" if ctx.cli == "oc" else "operatorhubio-catalog",
            operator_catalog_channel="stable",
            package_name=name
        )


@given(u'Opstree Redis operator is running')
def install_redis_operator(_context):
    operator = RedisOperator()
    if not operator.is_running():
        operator.install_operator_subscription()
        operator.is_running(wait=True)
    print("Opstree Redis operator is running")
