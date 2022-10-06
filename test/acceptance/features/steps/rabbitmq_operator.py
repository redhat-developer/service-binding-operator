from olm import Operator
from behave import step


class RabbitMqOperator(Operator):

    deploy_yaml = "https://github.com/rabbitmq/cluster-operator/releases/download/v1.9.0/cluster-operator.yml"

    def __init__(self, name="rabbitmq-cluster-operator"):
        super().__init__(name=name)


@step(u'RabbitMQ operator is running')
def install(_context):
    operator = RabbitMqOperator()
    operator.operator_namespace = "rabbitmq-system"
    if not operator.is_running():
        operator.openshift.apply_yaml_file(operator.deploy_yaml)
        operator.is_running(wait=True)
    print("RabbitMQ operator is running")


@step(u'RabbitMQ operator is removed')
def uninstall(_context):
    operator = RabbitMqOperator()
    operator.operator_namespace = "rabbitmq-system"
    if operator.is_running():
        operator.openshift.delete_from_yaml_file(operator.deploy_yaml)
        operator.is_not_running(wait=True)
