import re

from openshift import Openshift


class ServiceBindingRequest(object):

    openshift = Openshift()

    def create(self, yaml):
        sbr_create_output = self.openshift.oc_apply(yaml)
        return re.search(r'.*servicebindingrequest.apps.openshift.io/.*(created|unchanged)', sbr_create_output)
