from openshift import Openshift
from command import Command
import polling2


class App:
    openshift = Openshift()
    cmd = Command()
    name = ""
    namespace = ""
    app_image = ""
    route_url = ""

    def __init__(self, name, namespace, app_image):
        self.name = name
        self.namespace = namespace
        self.app_image = app_image

    def is_running(self, wait=False):
        output, exit_code = self.cmd.run(
            f"oc wait --for=condition=Available=True deployment/{self.name} -n {self.namespace} --timeout={300 if wait else 0}s")
        running = exit_code == 0
        if running:
            self.route_url = polling2.poll(lambda: self.openshift.get_route_host(self.name, self.namespace),
                                           check_success=lambda v: v != "", step=1, timeout=100)
        return running

    def install(self):
        create_new_app_output, exit_code = self.cmd.run(
            f"oc new-app --docker-image={self.app_image} --name={self.name} -n {self.namespace}")
        assert exit_code == 0, f"Non-zero exit code ({exit_code}) returned when attempting to create a new app: {create_new_app_output}"
        assert self.openshift.expose_service_route(self.name,
                                                   self.namespace) is not None, "Unable to expose service route"
        return self.is_running(wait=True)
