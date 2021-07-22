import os

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
        crd_type = "crd"
        crd_name = "servicebindings.service.binding"
        assert self.openshift.is_resource_in(crd_type, crd_name), f"CRD '{crd_name}' does not exist"
        return True

    def check_deployment(self):
        sbo_namespace = self.openshift.lookup_namespace_for_resource("deployments", "service-binding-operator")
        return sbo_namespace is not None

    def is_running(self):
        start_sbo = os.getenv("TEST_ACCEPTANCE_START_SBO")
        assert start_sbo is not None, "TEST_ACCEPTANCE_START_SBO is not set. It should be one of local, remote or operator-hub"
        assert start_sbo in {"local", "remote", "operator-hub"}, "TEST_ACCEPTANCE_START_SBO should be one of local, remote or operator-hub"

        if start_sbo == "local":
            assert not os.getenv("TEST_ACCEPTANCE_SBO_STARTED").startswith("FAILED"), "TEST_ACCEPTANCE_SBO_STARTED shoud not be FAILED."
            return self.check_crd()
        elif start_sbo == "remote":
            return self.check_crd() and self.check_deployment()
        elif start_sbo == "operator-hub":
            return False

    def get_name_pattern(self):
        return self.name_pattern.format(name=self.name)
