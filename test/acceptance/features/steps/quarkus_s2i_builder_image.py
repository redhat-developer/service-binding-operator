from command import Command
import re
from openshift import Openshift


class QuarkusS2IBuilderImage(object):

    tag = "19.1.1"
    namespace = "openshift"
    image = ""

    openshift = Openshift()

    def __init__(self, image="ubi-quarkus-native-s2i"):
        self.cmd = Command()
        self.image = image

    def is_present(self):
        get_name_output = self.get_name()
        if (get_name_output is not None):
            pattern = f'.*{self.image}:{self.tag}'
            if re.search(pattern, get_name_output) is not None:
                return True
        return False

    def get_name(self):
        output, exit_code = self.cmd.run(
            f'oc get is {self.image} -n {self.namespace} -o json | jq -rc \'.spec.tags[] | select(.annotations.tags == "builder").from.name\'')
        if self.image in output:
            if re.search("not found", output):
                return None
            else:
                return output.rstrip('\n')
        return None

    def import_and_patch(self):
        cmd_import = f"oc import-image quay.io/quarkus/{self.image}:{self.tag} -n {self.namespace} --confirm"
        print(f'===> Import CMD: {cmd_import}')
        (import_output, exit_code) = self.cmd.run(cmd_import)
        if exit_code != 0:
            return False
        if re.search(f'.*{self.image}\\simported', import_output) is None:
            return False
        spec = '{"spec": {"tags": [{"name" : \"%s\", "annotations": {"tags": "builder"}}]}}' % self.tag
        cmd_patch = f'oc patch is {self.image} -n {self.namespace} -p \'{spec}\''
        print(f'===> Patch CMD: {cmd_patch}')
        (patch_output, exit_code) = self.cmd.run(cmd_patch)
        if exit_code != 0:
            print(f"Result for patching image stream is {patch_output} with the exit code {exit_code}")
            return False
        if re.search(f'.*{self.image}\\spatched', patch_output) is None:
            return False
        return True
