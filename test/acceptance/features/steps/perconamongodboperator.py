from olm import Operator
from environment import ctx
from behave import given


class PerconaMongoDBOperator(Operator):

    def __init__(self, name="percona-server-mongodb-operator"):
        super().__init__(
            name=name,
            operator_catalog_source_name="operatorhubio-catalog",
            operator_catalog_channel="stable",
            operator_catalog_image="quay.io/operatorhubio/catalog:latest",
            package_name=name
        )


@given(u'Percona MongoDB operator is running')
def install_percona_mongodb_operator(context):
    operator = PerconaMongoDBOperator()
    operator.operator_namespace = context.namespace.name
    if not operator.is_running():
        if ctx.cli == "oc":
            operator.install_catalog_source()
        subscription = f'''
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: operatorgroup
  namespace: {operator.operator_namespace}
spec:
  targetNamespaces:
  - {operator.operator_namespace}
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: '{operator.name}'
  namespace: {operator.operator_namespace}
spec:
  channel: '{operator.operator_catalog_channel}'
  installPlanApproval: Automatic
  name: '{operator.package_name}'
  source: '{operator.operator_catalog_source_name}'
  sourceNamespace: {operator.operator_catalog_namespace}
        '''
        print(subscription)
        operator.openshift.apply(subscription)
        operator.openshift.approve_operator_subscription_in_namespace(operator.name, operator.operator_namespace)
        operator.is_running(wait=True)
    print("Percona MongoDB operator is running")
