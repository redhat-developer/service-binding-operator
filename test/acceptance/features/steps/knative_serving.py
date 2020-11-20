import re
from environment import ctx
from command import Command
from openshift import Openshift


class KnativeServing(object):

    openshift = Openshift()
    cmd = Command()

    name = ""
    namespace = ""

    def __init__(self, name="knative-serving", namespace="knative-serving"):
        self.name = name
        self.namespace = namespace
        self.knative_serving_yaml_template = '''
apiVersion: operator.knative.dev/v1alpha1
kind: KnativeServing
metadata:
  name: {name}
  namespace: {namespace}
'''

    def is_present(self):
        cmd = f'{ctx.cli} get knativeserving.operator.knative.dev {self.name} -n {self.namespace}'
        output, exit_code = self.cmd.run(cmd)
        if exit_code != 0:
            print(f"cmd-{cmd} result for getting available knative serving is {output} with the exit code {exit_code}")
            return False
        return True

    def create(self):
        serving_output = self.openshift.apply(self.knative_serving_yaml_template.format(name=self.name, namespace=self.namespace))
        pattern = f'knativeserving.(serving|operator).knative.dev/{self.name}\\screated'
        if re.search(pattern, serving_output):
            return True
        print(f"Pattern {pattern} did not match as creating knative serving yaml output is {serving_output}")
        return False
