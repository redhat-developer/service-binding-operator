import os

from openshift import Openshift
from pyshould import should, should_not


class Servicebindingoperator():
    openshift = Openshift()
    name = ""
    namespace = ""

    def __init__(self,  name="service-binding-operator", namespace="openshift-operators"):
        self.namespace = namespace
        self.name = name

    def check_resources(self):
        self.openshift.is_resource_in("servicebindingrequest") | should.be_truthy.desc("CRD is in")
        self.openshift.search_resource_in_namespace("rolebindings", self.name, self.namespace) | should_not.be_none.desc("Role binding is in")
        self.openshift.search_resource_in_namespace("roles", self.name, self.namespace) | should_not.be_none.desc("Role is in")
        self.openshift.search_resource_in_namespace("serviceaccounts", self.name, self.namespace) | should_not.be_none.desc("Service Account")
        return True

    def is_running(self):
        start_sbo = os.getenv("TEST_ACCEPTANCE_START_SBO")
        start_sbo | should_not.be_none.desc("TEST_ACCEPTANCE_START_SBO is set")
        start_sbo | should.be_in({"local", "operator-hub"})

        if start_sbo == "local":
            os.getenv("TEST_ACCEPTANCE_SBO_STARTED") | should_not.start_with("FAILED").desc("TEST_ACCEPTANCE_SBO_STARTED is not FAILED")
        elif start_sbo == "operator-hub":
            return False

        return self.check_resources()
