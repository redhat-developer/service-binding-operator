import re
from environment import ctx

from command import Command
from openshift import Openshift


class PostgresDB(object):

    cmd = Command()
    openshift = Openshift()

    pod_name_pattern = "{name}.*"

    name = ""
    namespace = ""

    database_yaml_template = """---
apiVersion: postgresql.baiju.dev/v1alpha1
kind: Database
metadata:
  name: {name}
  namespace: {namespace}
spec:
  image: docker.io/postgres
  imageName: postgres
  dbName: {name}
"""

    def __init__(self, name, namespace):
        self.name = name
        self.namespace = namespace

    def create(self):
        db_create_output = self.openshift.apply(self.database_yaml_template.format(name=self.name, namespace=self.namespace))
        return re.search(r'.*database.postgresql.baiju.dev/%s\s(created|unchanged)' % self.name, db_create_output)

    def is_running(self, wait=False):
        if wait:
            pod_name = self.openshift.wait_for_pod(self.pod_name_pattern.format(name=self.name), self.namespace, timeout=120)
        else:
            pod_name = self.openshift.search_pod_in_namespace(self.pod_name_pattern.format(name=self.name), self.namespace)
        if pod_name is not None:
            pod_status = self.openshift.check_pod_status(pod_name, self.namespace)
            print("The pod {} is running: {}".format(self.name, pod_status))
            output, exit_code = self.cmd.run(f'{ctx.cli} get db {self.name} -n {self.namespace} -o jsonpath="{{.status.dbConnectionIP}}"')
            if exit_code == 0 and re.search(r'\d+\.\d+\.\d+\.\d+', output):
                print(f"The DB {self.name} is up and listening at {output}.")
                return True
        return False

    def get_connection_ip(self):
        cmd = f'{ctx.cli} get db {self.name} -n {self.namespace} -o jsonpath="{{.status.dbConnectionIP}}"'
        return self.cmd.run(cmd)

    def check_pod_status(self, status="Running"):
        return self.openshift.check_pod_status(self.name, self.namespace, wait_for_status=status)
