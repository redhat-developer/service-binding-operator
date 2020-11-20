from olm import Operator
from environment import ctx


class EtcdOperator(Operator):

    operator_catalog_source_name = "operatorhubio-catalog" if ctx.cli == "kubectl" else "community-operators"
    operator_catalog_channel = "clusterwide-alpha"

    def __init__(self, name="etcd"):
        self.name = name
        self.operator_catalog_channel = "clusterwide-alpha"
        self.package_name = "etcd"
