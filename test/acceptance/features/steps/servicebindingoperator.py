import os
from command import Command
from steps.environment import ctx
import polling2

from openshift import Openshift


class Servicebindingoperator():
    openshift = Openshift()
    cmd = Command()
    name = ""
    namespace = ""

    name_pattern = f"{name}.*"

    def __init__(self,  name="service-binding-operator", namespace="openshift-operators"):
        self.namespace = namespace
        self.name = name

    def check_crd(self, wait=False):
        crd_type = "crd"
        crd_name = "servicebindings.servicebinding.io"
        if wait:
            return polling2.poll(target=lambda: self.openshift.is_resource_in(crd_type, crd_name),
                                 step=5, timeout=400, ignore_exceptions=(ValueError,))
        else:
            return self.openshift.is_resource_in(crd_type, crd_name)

    def check_deployment(self, wait=False, csv_version=None):
        if wait:
            sbo_namespace = polling2.poll(
                target=lambda: self.openshift.lookup_namespace_for_resource("deployments", "service-binding-operator"),
                check_success=lambda o: o is not None, step=5, timeout=400, ignore_exceptions=(ValueError,))
        else:
            sbo_namespace = self.openshift.lookup_namespace_for_resource("deployments", "service-binding-operator")
        if sbo_namespace is None:
            return False

        if csv_version is not None:
            cmd = f"{ctx.cli} get deployment/service-binding-operator -o jsonpath='{{.metadata.labels.olm\\.owner}}' -n {sbo_namespace}"
            (output, code) = polling2.poll(target=lambda: tuple(self.cmd.run(cmd)),
                                           check_success=lambda o: o[1] == 0 and o[0] == csv_version, step=5, timeout=300, ignore_exceptions=(ValueError,))
            assert code == 0, f"Non-zero return code while trying to check SBO deployment is owned by {csv_version} CSV: {output}"

            cmd = f"{ctx.cli} get secrets/service-binding-operator-service-cert -o jsonpath='{{.metadata.ownerReferences[0].name}}' -n {sbo_namespace}"
            (output, code) = polling2.poll(target=lambda: tuple(self.cmd.run(cmd)),
                                           check_success=lambda o: o[1] == 0 and o[0] == csv_version, step=5, timeout=300, ignore_exceptions=(ValueError,))
            assert code == 0, f"Non-zero return code while trying to check SBO cert secret is owned by {csv_version} CSV: {output}"

        output, code = self.cmd.run(f"{ctx.cli} rollout status -w deployment/service-binding-operator -n {sbo_namespace}")
        assert code == 0, f"Non-zero return code while trying to SBO is healthy: {output}"
        return sbo_namespace is not None

    def is_running(self, wait=False, csv_version=None):
        start_sbo = os.getenv("TEST_ACCEPTANCE_START_SBO")
        assert start_sbo is not None, "TEST_ACCEPTANCE_START_SBO is not set. It should be one of local, remote or operator-hub"
        assert start_sbo in {"local", "remote", "operator-hub", "scenarios"}, "TEST_ACCEPTANCE_START_SBO should be one of local, remote or operator-hub"

        if start_sbo == "local":
            assert not os.getenv("TEST_ACCEPTANCE_SBO_STARTED").startswith("FAILED"), "TEST_ACCEPTANCE_SBO_STARTED should not be FAILED."
            return self.check_crd(wait)
        elif start_sbo == "remote" or start_sbo == "scenarios":
            return self.check_crd(wait) and self.check_deployment(wait=wait, csv_version=csv_version)
        elif start_sbo == "operator-hub":
            return False

    def get_name_pattern(self):
        return self.name_pattern.format(name=self.name)
