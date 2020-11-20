from olm import Operator


class DbOperator(Operator):

    def __init__(self, name="postgresql-operator"):
        self.name = name
        self.operator_catalog_source_name = "sample-db-operators"
        self.operator_catalog_image = "quay.io/redhat-developer/sample-db-operators-olm:v1"
        self.operator_catalog_channel = "beta"
        self.package_name = "db-operators"
