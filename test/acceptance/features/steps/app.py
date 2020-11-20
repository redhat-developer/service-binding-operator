from openshift import Openshift
from command import Command
from environment import ctx
import polling2


class App(object):
    openshift = Openshift()
    cmd = Command()
    name = ""
    namespace = ""
    app_image = ""
    route_url = ""
    port = ""
    bindingRoot = ""
    remote_repo_repository = ""

    def __init__(self, name, namespace, app_image=None, port="", remote_repo_repository=None):
        self.name = name
        self.namespace = namespace
        self.app_image = app_image
        self.port = port
        self.remote_repo_repository = remote_repo_repository

    def is_running(self, wait=False):
        output, exit_code = self.cmd.run(
            f"{ctx.cli} wait --for=condition=Available=True deployment/{self.name} -n {self.namespace} --timeout={300 if wait else 0}s")
        running = exit_code == 0
        if running:
            self.route_url = polling2.poll(lambda: self.base_url(),
                                           check_success=lambda v: v != "", step=1, timeout=100)
        return running

    def install(self, bindingRoot=None):
        self.openshift.new_app(self.name, self.app_image, self.namespace, bindingRoot)
        self.openshift.expose_service_route(self.name, self.namespace, self.port)
        return self.is_running(wait=True)

    def base_url(self):
        return self.openshift.get_route_host(self.name, self.namespace)

    def install_from_remote_repository(self):
        self.openshift.new_app_from_remote_repository(self.name, self.remote_repo_repository, self.namespace)
        self.openshift.expose_service_route(self.name, self.namespace, self.port)
        return self.is_running(wait=True)
