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

    def check_resources(self):
        crd_name = "servicebinding"
        assert self.openshift.is_resource_in(crd_name), f"CRD '{crd_name}' does not exist"

        name_pattern = self.get_name_pattern()

        output = self.openshift.search_resource_in_namespace("rolebindings", name_pattern, self.namespace)
        assert output is not None, f"Unable to find role binding with name that maches '{name_pattern}' in namespace '{self.namespace}': {output}"

        output = self.openshift.search_resource_in_namespace("roles", name_pattern, self.namespace)
        assert output is not None, f"Unable to find role with name that maches '{name_pattern}' in namespace '{self.namespace}': {output}"

        output = self.openshift.search_resource_in_namespace("serviceaccounts", name_pattern, self.namespace)
        assert output is not None,  f"Unable to find service account with name that maches '{name_pattern}' in namespace '{self.namespace}': {output}"

        return True

    def is_running(self):
        start_sbo = os.getenv("TEST_ACCEPTANCE_START_SBO")
        assert start_sbo is not None, "TEST_ACCEPTANCE_START_SBO is not set. It should be one of local, remote or operator-hub"
        assert start_sbo in {"local", "remote", "operator-hub"}, "TEST_ACCEPTANCE_START_SBO should be one of local, remote or operator-hub"

        if start_sbo == "local":
            assert not os.getenv("TEST_ACCEPTANCE_SBO_STARTED").startswith("FAILED"), "TEST_ACCEPTANCE_SBO_STARTED shoud not be FAILED."
        elif start_sbo == "operator-hub":
            return False

        return self.check_resources()

    def get_name_pattern(self):
        return self.name_pattern.format(name=self.name)
