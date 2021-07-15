from openshift import Openshift


class Servicebindingoperator():
    openshift = Openshift()
    name = ""
    namespace = ""

    name_pattern = f"{name}.*"

    def __init__(self,  name="service-binding-operator", namespace="openshift-operators"):
        self.namespace = namespace
        self.name = name

    def check_crd(self):
        crd_name = "servicebinding"
        assert self.openshift.is_resource_in(crd_name), f"CRD '{crd_name}' does not exist"
        return True

    def check_deployment(self):
        sbo_namespace = self.openshift.lookup_namespace_for_resource("deployments", "service-binding-operator")
        return sbo_namespace is not None

    def is_running(self):
        return self.check_crd() and self.check_deployment()

    def get_name_pattern(self):
        return self.name_pattern.format(name=self.name)
