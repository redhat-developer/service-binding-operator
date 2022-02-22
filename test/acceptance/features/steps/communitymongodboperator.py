from olm import Operator
from behave import given

class CommunityMongoDBOperator(Operator):

    def __init__(self, name="mongodb-kubernetes-operator"):
        self.name = name
        self.package_name = name

@given(u'MongoDB Community operator is running')
def step_to_install_mongodb_community_operator(context):
    operator = CommunityMongoDBOperator()
    if not operator.is_running():
        #import os
        #import time
        #if(not os.path.isdir("mongodb-kubernetes-operator")):
        #    print("Directory not found")
        #    os.popen("git clone https://github.com/mongodb/mongodb-kubernetes-operator.git")
        #    time.sleep(20)
        #else:
        #    print("Directory found")
        #os.chdir("mongodb-kubernetes-operator")
        #newPath = str(os.popen("pwd").read()).replace("\n", "")
        #print("New Path: ", newPath)
        mongonamespace=operator.openshift.operators_namespace
        operator.openshift.apply_yaml_file("https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/master/config/crd/bases/mongodbcommunity.mongodb.com_mongodbcommunity.yaml",mongonamespace)
        #operator.openshift.apply_yaml_file(newPath + "/config/crd/bases/mongodbcommunity.mongodb.com_mongodbcommunity.yaml")
        operator.openshift.apply_yaml_file("https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/master/config/rbac/role_binding_database.yaml",mongonamespace)
        operator.openshift.apply_yaml_file("https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/master/config/rbac/role_binding.yaml",mongonamespace)
        operator.openshift.apply_yaml_file("https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/master/config/rbac/role_database.yaml",mongonamespace)
        operator.openshift.apply_yaml_file("https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/master/config/rbac/role.yaml",mongonamespace)
        operator.openshift.apply_yaml_file("https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/master/config/rbac/service_account_database.yaml",mongonamespace)
        operator.openshift.apply_yaml_file("https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/master/config/rbac/service_account.yaml",mongonamespace) 
        operator.openshift.apply_yaml_file("https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/master/config/manager/manager.yaml",mongonamespace)
        #operator.openshift.apply_yaml_file(newPath + "/config/samples/mongodb.com_v1_mongodbcommunity_cr.yaml")
        operator.is_running(wait=True)
        print("MongoDB Community operator is running")