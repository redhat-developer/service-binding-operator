from olm import Operator
from environment import ctx
from behave import given


class PerconaMongoDBOperator(Operator):

    def __init__(self, name="percona-server-mongodb-operator"):
        self.name = name
        if ctx.cli == "oc":
            self.operator_catalog_source_name = "community-operators"
        else:
            self.operator_catalog_source_name = "operatorhubio-catalog"
        self.operator_catalog_channel = "stable"
        self.package_name = name


@given(u'Percona MongoDB operator is running')
def install_percona_mongodb_operator(context):
    operator = PerconaMongoDBOperator()
    operator.openshift.operators_namespace = context.namespace.name
    if not operator.is_running():
        subscription = f'''
---
apiVersion: v1
kind: Namespace
metadata:
  name: percona
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: operatorgroup
  namespace: percona
spec:
  targetNamespaces:
  - {operator.openshift.operators_namespace}
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: '{operator.name}'
  namespace: percona
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
    print("Percona MongoDB operator is running")
