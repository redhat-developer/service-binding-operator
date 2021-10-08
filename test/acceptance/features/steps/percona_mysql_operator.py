from olm import Operator
from environment import ctx
from behave import given


class PerconaMysqlOperator(Operator):

    def __init__(self, name="percona-xtradb-cluster-operator"):
        self.name = name
        if ctx.cli == "oc":
            self.operator_catalog_source_name = "community-operators"
        else:
            self.operator_catalog_source_name = "operatorhubio-catalog"
        self.operator_catalog_channel = "stable"
        self.package_name = "percona-xtradb-cluster-operator"


@given(u'Percona Mysql operator is running')
def install(context):
    operator = PerconaMysqlOperator()
    operator.openshift.operators_namespace = context.namespace.name
    if not operator.is_running():
        subscription = f'''
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: operatorgroup
  namespace: {operator.openshift.operators_namespace}
spec:
  targetNamespaces:
  - {operator.openshift.operators_namespace}
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: '{operator.name}'
  namespace: {operator.openshift.operators_namespace}
spec:
  channel: '{operator.operator_catalog_channel}'
  installPlanApproval: Automatic
  name: '{operator.package_name}'
  source: '{operator.operator_catalog_source_name}'
  sourceNamespace: {operator.openshift.olm_namespace}
        '''
        print(subscription)
        operator.openshift.apply(subscription)
        operator.is_running(wait=True)
    print("Percona Mysql operator is running")
