from olm import Operator


class EtcdOperator(Operator):

    operator_catalog_source_name = "operatorhubio-catalog"
    operator_catalog_channel = "clusterwide-alpha"

    def __init__(self, name="etcd"):
        self.name = name
        self.package_name = "etcd"
