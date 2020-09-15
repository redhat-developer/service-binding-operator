import re

from openshift import Openshift


class ServiceBinding(object):

    openshift = Openshift()

    def create(self, yaml):
        sbr_create_output = self.openshift.oc_apply(yaml)
        return re.search(r'.*servicebinding.operators.coreos.com/.*(created|unchanged)', sbr_create_output)
