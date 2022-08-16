from olm import Operator
from behave import given


class RabbitMqOperator(Operator):

    def __init__(self, name="rabbitmq-cluster-operator"):
        super().__init__(name=name)


@given(u'RabbitMQ operator is running')
def install(_context):
    operator = RabbitMqOperator()
    operator.operator_namespace = "rabbitmq-system"
    if not operator.is_running():
        operator.openshift.apply_yaml_file("https://github.com/rabbitmq/cluster-operator/releases/download/v1.9.0/cluster-operator.yml")
        operator.is_running(wait=True)
    print("RabbitMQ operator is running")
