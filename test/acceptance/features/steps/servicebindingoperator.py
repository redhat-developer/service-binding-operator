import os

from openshift import Openshift
from pyshould import should, should_not


class Servicebindingoperator():
    openshift = Openshift()
    name = ""
    namespace = ""

    name_pattern = f"{name}.*"

    def __init__(self,  name="service-binding-operator", namespace="openshift-operators"):
        self.namespace = namespace
        self.name = name

    def check_resources(self):
        self.openshift.is_resource_in("servicebinding") | should.be_truthy.desc("CRD is in")
        self.openshift.search_resource_in_namespace("rolebindings", self.get_name_pattern(), self.namespace) | should_not.be_none.desc("Role binding is in")
        self.openshift.search_resource_in_namespace("roles", self.get_name_pattern(), self.namespace) | should_not.be_none.desc("Role is in")
        self.openshift.search_resource_in_namespace("serviceaccounts", self.get_name_pattern(), self.namespace) | should_not.be_none.desc("Service Account")
        return True

    def is_running(self):
        start_sbo = os.getenv("TEST_ACCEPTANCE_START_SBO")
        start_sbo | should_not.be_none.desc("TEST_ACCEPTANCE_START_SBO is set")
        start_sbo | should.be_in({"local", "remote", "operator-hub"})

        if start_sbo == "local":
            os.getenv("TEST_ACCEPTANCE_SBO_STARTED") | should_not.start_with("FAILED").desc("TEST_ACCEPTANCE_SBO_STARTED is not FAILED")
        elif start_sbo == "operator-hub":
            return False

        return self.check_resources()

    def get_name_pattern(self):
        return self.name_pattern.format(name=self.name)
