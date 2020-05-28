package examples_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/examples/util"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	exampleName = "nodejs_postgresql"
	operatorsNS = "openshift-operators"
	oprName     = "service-binding-operator"
	appNS       = "service-binding-demo"

	nodeJsApp = "https://github.com/pmacik/nodejs-rest-http-crud"
	appName   = "nodejs-rest-http-crud"

	clusterAvailable = false

	pkgManifest                      = "db-operators"
	bckSvc                           = "postgresql-operator"
	dbName                           = "db-demo"
	sbr                              = ""
	podNameSBO, podNameBackingSvcOpr string
)

func TestNodeJSPostgreSQL(t *testing.T) {

	t.Run("set-example-dir", SetExampleDir)
	t.Run("get-oc-status", GetOCStatus)

	t.Run("install-service-binding-operator", MakeInstallServiceBindingOperator)
	t.Run("install-backing-service-operator", MakeInstallBackingServiceOperator)
	t.Run("create-project", CreateProject)
	t.Run("import-nodejs-app", ImportNodeJSApp)
	t.Run("create-backing-db-instance", CreateBackingDbInstance)
	t.Run("createservice-binding-request", CreateServiceBindingRequest)
	t.Run("delete-project", DeleteProject)
	t.Run("uninstall-backing-db-instance", UninstallBackingServiceOperator)
	t.Run("uninstall-service-binding-operator", UninstallServiceBindingOperator)
}

//SetExampleDir tests that the current example directory was set as a working directory for running the commands.
func SetExampleDir(t *testing.T) {
	examplePath := fmt.Sprintf("%s/%s", util.GetExamplesDir(), exampleName)
	util.SetDir(examplePath)
	res := strings.TrimSpace(util.Run("pwd").Stdout())
	require.Equal(t, examplePath, res)
}

//Logs the oc status
func GetOCStatus(t *testing.T) {
	t.Log("--- Getting OC Status ---")
	result := util.Run("oc", "status")
	require.Equal(t, 0, result.ExitCode)
	clusterAvailable = true
	t.Log(" *** Connected to cluster *** ")
}

func MakeInstallServiceBindingOperator(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("Installing service binding operator into the cluster...")
	res := util.GetOutput(util.Run("make", "install-service-binding-operator-master"))
	resExp := strings.TrimSpace(strings.Split(res, "subscription.operators.coreos.com/service-binding-operator")[1])
	fmt.Printf("subscription output is %s \n", resExp)
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while installing service binding operator", resExp)

	// with openshift-operators namespace, capture the install plan
	t.Log(" Fetching the status of install plan name ")
	ipName, err := util.GetCmdResult("", "oc", "get", "subscription", oprName, "-n", operatorsNS, "-o", `jsonpath={.status.installplan.name}`)
	require.NoErrorf(t, err, "<<Error while getting the name of the install plan as the result is - %s >>", ipName)
	t.Logf("-> Install plan-ip name is %s \n", ipName)

	//// with openshift-operators namespace, capture the install plan status of db-operators
	t.Log(" Fetching the status of install plan status ")
	ipStatus, err := util.GetCmdResult("Complete", "oc", "get", "ip", "-n", operatorsNS, ipName, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Error while getting the status of install plan as the result is - %s >>", ipStatus)
	t.Logf("-> Install plan-ip name is %s \n", ipStatus)

	podNameSBO, err = util.GetPodNameFromListOfPods(operatorsNS, oprName)

	require.NoErrorf(t, err, "<<There are no pods running service binding operator in the cluster as the result is - %s >>", podNameSBO)
	require.Containsf(t, podNameSBO, oprName, "list does not contain %s pod from the list of pods running service binding operator in the cluster", resExp)
	t.Logf("-> Pod name is %s \n", podNameSBO)

	t.Log("Fetching the status of running pod")
	podStatus, err := util.GetCmdResult("Running", "oc", "get", "pod", podNameSBO, "-n", operatorsNS, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Error while getting the status of the pod running SBO as the result is - %s >>", podStatus)
	t.Logf("-> Pod name is %s \n", podStatus)
	require.Equal(t, "Running", podStatus, "'pod plan status' is %d \n", podStatus)

}

func MakeInstallBackingServiceOperator(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Installing backing service operator source into the cluster...")
	res := util.GetOutput(util.Run("make", "install-backing-db-operator-source"))
	resExp := strings.TrimSpace(strings.Split(res, "operatorsource.operators.coreos.com/db-operators")[1])
	fmt.Printf("db operator installation source output is %s \n", resExp)
	require.True(t, strings.Contains(resExp, "created") || strings.Contains(resExp, "unchanged"), "Output %s does not contain created or unchanged while installing db service operator source", resExp)

	//with this command 'oc get packagemanifest | grep db-operators' make sure there is an entry to the package manifest
	t.Log("get the package manifest for db-operators...\n")
	manifest, err := util.GetCmdResult(pkgManifest, "oc", "get", "packagemanifest", pkgManifest, "-o", `jsonpath={.metadata.name}`)
	require.NoErrorf(t, err, "<<Error while getting the name of the manifset as the result is - %s >>", manifest)
	t.Logf("-> %s has an entry in the package manifest \n", manifest)
	require.NotEmptyf(t, manifest, "There are number of manifest listed in a cluster %s \n", manifest)

	//Install the subscription using this command: make install-backing-db-operator-subscription
	t.Log("Installing backing service operator subscription into the cluster...")
	dbOprRes := util.GetOutput(util.Run("make", "install-backing-db-operator-subscription"))
	subRes := strings.TrimSpace(strings.Split(dbOprRes, "subscription.operators.coreos.com/db-operators")[1])
	t.Logf("subscription output is %s \n", subRes)

	// with openshift-operators namespace, capture the install plan
	ipName, err := util.GetCmdResult("", "oc", "get", "subscription", pkgManifest, "-n", operatorsNS, "-o", `jsonpath={.status.installplan.name}`)
	require.NoErrorf(t, err, "<<Error while getting the install plan name as the result is - %s >>", ipName)
	t.Logf("-> Pod name is %s \n", ipName)

	//// with openshift-operators namespace, capture the install plan status of db-operators
	t.Log(" Fetching the status of install plan ")
	ipStatus, err := util.GetCmdResult("Complete", "oc", "get", "ip", "-n", operatorsNS, ipName, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Error while getting the install plan status as the result is - %s >>", ipStatus)
	t.Logf("-> Pod name is %s \n", ipStatus)
	require.Equal(t, ipStatus, "Complete", "'install plan status' is %d \n", ipStatus)

	podNameBackingSvcOpr, err = util.GetPodNameFromListOfPods(operatorsNS, bckSvc)
	require.NoErrorf(t, err, "<<There are no pods running backing service operator in the cluster as the result is - %s >>", podNameBackingSvcOpr)
	require.Containsf(t, podNameBackingSvcOpr, bckSvc, "list does not contain %s pod from the list of pods running backing service operator in the cluster", bckSvc)
	t.Logf("-> Pod name is %s \n", podNameBackingSvcOpr)

	t.Log("Fetching the status of running pod")
	podStatus, err := util.GetCmdResult("Running", "oc", "get", "pod", podNameBackingSvcOpr, "-n", operatorsNS, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Error while getting the status of the pod running service binding operator as the result is - %s >>", podStatus)
	require.Equal(t, "Running", podStatus, "pod status is %d \n", podStatus)

	t.Log("Checking the backing service's CRD is installed")
	crd, err := util.GetCmdResult("databases.postgresql.baiju.dev", "oc", "get", "crd", "databases.postgresql.baiju.dev")
	require.NoErrorf(t, err, "<<Error while getting the crd as the result is - %s >>", crd)
	require.NotEmpty(t, crd, "packing service CRD not found")
}

func CreateProject(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("Creating a project into the cluster...")
	res := util.GetOutput(util.Run("make", "create-project"))
	require.NotEmptyf(t, res, "Project not created because of - %s \n", res)

	expectedCreateProjectResult, err := regexp.Compile(`project\s.` + appNS + `.\son\sserver.*`)
	require.NoErrorf(t, err, "<<Error while applying regular expression to create project result - %s >>", expectedCreateProjectResult)
	createProjectRes, checkFlag := util.GetCreateDeleteProjectResult(res, expectedCreateProjectResult)

	require.Equal(t, true, checkFlag, "Unable to create the project as the value is %s", createProjectRes)
	require.Containsf(t, createProjectRes, appNS, "Namespace - %s is not created because of %s", appNS, res)

	t.Logf("-> Project created - %s \n", createProjectRes)
	t.Log("Get the name of project added to the cluster...\n")
	getProject, err := util.GetCmdResult("", "oc", "get", "project", appNS, "-o", `jsonpath={.metadata.name}`)
	require.NoErrorf(t, err, "<<Error while getting the name of the project as the result is - %s >>", getProject)
	t.Logf("-> Project created - %s \n", getProject)
	require.Equal(t, getProject, appNS, "Namespace created is %d \n", getProject)

	t.Log("Get the status of the project added to the cluster...\n")
	projectStatus, err := util.GetCmdResult("Active", "oc", "get", "project", appNS, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Error while getting the status of the project creation as the result is - %s >>", projectStatus)
	t.Logf("-> Project created %s has the status %s \n", appNS, projectStatus)
	require.Equal(t, "Active", projectStatus, "Namespace status is %d \n", projectStatus)

	t.Log(" Setting the project to service-binding-demo ")
	projectRes, err := util.GetCmdResult("", "oc", "project", appNS)
	require.NoErrorf(t, err, "<<Error while setting the project as the result is - %s >>", projectRes)
	t.Logf("Project/Namespace set is %s", projectRes)
}

func ImportNodeJSApp(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("import an nodejs app to bind with backing service...")

	nodeJsAppArg := "nodejs~" + nodeJsApp
	t.Log(" Fetching the build config name of the app ")
	app, err := util.GetCmdResult("", "oc", "new-app", nodeJsAppArg, "--name", appName, "-n", appNS)
	require.NoErrorf(t, err, "<<Error while creating an nodejs app as the result is - %s >>", app)
	t.Logf("-> App info - %s \n", app)

	t.Log(" Fetching the build config name of the app ")
	bc, err := util.GetCmdResult("", "oc", "get", "bc", "-o", `jsonpath={.items[*].metadata.name}`)
	require.NoErrorf(t, err, "<<Error while getting the build config as the result is - %s >>", bc)
	require.Contains(t, appName, bc, "Build config name is %d \n", bc)

	t.Log(" Fetching the buid info - name of the app ")
	buildName, err := util.GetCmdResult("", "oc", "get", "build", "-o", `jsonpath={.items[*].metadata.name}`)
	require.NoErrorf(t, err, "<<Error while getting the name of the build as the result is - %s >>", buildName)
	require.Contains(t, buildName, bc, "Build config name is %d \n", buildName)
	t.Logf(" Buid name of an app %s", buildName)

	t.Log(" Fetching the buid status of the app ")
	buildStatus, err := util.GetCmdResultWithTimeout("Complete", 5*time.Second, 5*time.Minute, "oc", "get", "build", buildName, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Build status is %s >>", buildStatus)
	require.Equal(t, "Complete", buildStatus, "Build status of config name %d is %d \n", buildName, buildStatus)
	t.Logf(" Buid status of an app %s", buildStatus)

	t.Log(" Fetching the list of pod for the build of an app ")
	expBuildPodName := buildName + "-build"
	buildPodName, err := util.GetPodNameFromListOfPods(appNS, expBuildPodName)
	require.NoErrorf(t, err, "<<There are no pods running nodejs app in the cluster using deployment config as the result is - %s >>", buildPodName)
	require.Containsf(t, expBuildPodName, buildPodName, "list does not contain %s build pod from the list of pods running builds in the cluster", buildPodName)
	t.Logf("-> Pod name is %s \n", buildPodName)

	t.Log("Fetching the status of running build pod")
	buildPodStatus, err := util.GetCmdResult("Succeeded", "oc", "get", "pods", buildPodName, "-n", appNS, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Error while getting the pod status of build as the result is - %s >>", buildPodStatus)
	require.Equal(t, "Succeeded", buildPodStatus, "Build pod status is %d \n", buildPodStatus)

	t.Log("Fetching the name of deployment config")
	dc, err := util.GetCmdResult("", "oc", "get", "dc", "-n", appNS, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NoErrorf(t, err, "<<Error while getting deployment config name as the result is - %s >>", dc)
	require.Equal(t, bc, dc, "DeploymentConfig name is %d \n", dc)
	t.Logf("-> Deployment Config name is %s \n", dc)

	UseDeployment(t, dc)

	t.Log(" Exposing an app ")
	svc := "svc/" + bc
	name := "--name=" + bc
	exposeRes, err := util.GetCmdResult("", "oc", "expose", svc, name)
	require.NoErrorf(t, err, "<<Error while exposing the nodejs app as the result is - %s >>", exposeRes)
	t.Logf("app exposed result is %s", exposeRes)

	t.Log(" Fetching the route of an app ")
	route, err := util.GetCmdResult("", "oc", "get", "route", bc, "-n", appNS, "-o", `jsonpath={.status.ingress[0].host}`)
	require.NoErrorf(t, err, "<<Error while getting the route as the result is - %s >>", route)
	t.Logf("-> ROUTE - %s \n", route)

	appStatusEndpoint := fmt.Sprintf("http://%s/api/status/dbNameCM", route)
	checkNodeJSAppFrontend(t, appStatusEndpoint, "N/A")
}

func UseDeployment(t *testing.T, dc string) {

	t.Log(" Delete the deployment config ")
	deletedStatus, err := util.GetCmdResult("", "oc", "delete", "dc", dc, "-n", appNS, "--wait=true")
	require.NoErrorf(t, err, "<<Error while deleting the deployment config as the result is - %s >>", deletedStatus)
	require.Containsf(t, deletedStatus, "deleted", "Deployment config is deleted with the message %d \n", deletedStatus)
	t.Logf("-> Deployment config is deleted with the message %s \n", deletedStatus)
	dcPod := dc + "-1-deploy"

	pods := util.GetPodLst(operatorsNS)
	checkFlag, podName := util.SrchItemFromLst(pods, dcPod)

	require.Equal(t, false, checkFlag, "list contains deployment config pod from the list of pods running in the cluster")
	require.Equal(t, "", podName, "list contains %s deployment config pod from the list of pods running in the cluster", podName)
	t.Logf("-> List does not contain deployment config pod running in the cluster")

	deploymentData, err := util.GetCmdResult("", "oc", "apply", "-f", "deployment.yaml")
	require.NoErrorf(t, err, "<<Error while applying the deployment yaml as the result is - %s >>", deploymentData)
	t.Logf("-> Deployment config is deleted with the message %s \n", deploymentData)

	deploymentPodName, err := util.GetPodNameFromListOfPods(appNS, appName)
	require.NoErrorf(t, err, "<<There are no pods running nodejs app using deployment in the cluster as the result is - %s >>", deploymentPodName)
	require.Containsf(t, deploymentPodName, appName, "list does not contain %s build pod from the list of pods running builds in the cluster", deploymentPodName)
	t.Logf("-> deployment build Pod name is %s \n", deploymentPodName)

	t.Log("Fetching the name of deployment")
	deployment, err := util.GetCmdResult("", "oc", "get", "deployment", "-n", appNS, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NoErrorf(t, err, "<<Unable to get the deployment name as the result is - %s >>", deployment)
	require.Equal(t, appName, deployment, "Deployment name is %d \n", deployment)
	t.Logf("-> Deployment name is %s \n", deployment)

	t.Log("Fetching the status of deployment ")
	deploymentStatus, err := util.GetCmdResult("True", "oc", "get", "deployment", deployment, "-n", appNS, "-o", `jsonpath={.status.conditions[*].status}`)
	require.NoErrorf(t, err, "<<Unable to get the deployment status as the result is - %s >>", deploymentStatus)
	require.Contains(t, deploymentStatus, "True", "Deployment status is %d \n", deploymentStatus)
	t.Logf("-> Deployment status is %s \n", deploymentStatus)

	env := util.GetOutput(util.Run("oc", "get", "deploy", appName, "-n", appNS, "-o", `jsonpath={.spec.template.spec.containers[0].env}`))
	require.Equalf(t, "", env, "Env details are %d \n", env)

	envFrom := util.GetOutput(util.Run("oc", "get", "deploy", appName, "-n", appNS, "-o", `jsonpath={.spec.template.spec.containers[0].envFrom}`))
	require.Equalf(t, "", envFrom, "EnvFrom details are %d \n", envFrom)
}

func CreateBackingDbInstance(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Creating Backing DB instance for backing service to connect with the app...")
	res := util.GetOutput(util.Run("make", "create-backing-db-instance"))
	resExp := strings.TrimSpace(strings.Split(res, dbName)[1])
	fmt.Printf("Result of created backing DB instance is %s \n", resExp)
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while installing db instance", resExp)

	t.Log("Fetching the name of running db instance")
	dbInstanceName, err := util.GetCmdResult("", "oc", "get", "db", "-n", appNS, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NoErrorf(t, err, "<<Unable to get the db instance name as the result is - %s >>", dbInstanceName)
	require.Equal(t, dbName, dbInstanceName, "db instance name is %d \n", dbInstanceName)
	t.Logf("-> DB instance name is %s \n", dbInstanceName)

	connectionIP, err := util.GetCmdResult("", "oc", "get", "db", dbInstanceName, "-o", "jsonpath={.status.dbConnectionIP}")
	require.NoErrorf(t, err, "<<Unable to get the db connection IP as the result is - %s >>", connectionIP)
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	require.Equalf(t, true, re.MatchString(connectionIP), "IP address of the pod which runs the database instance is %d", connectionIP)
	t.Logf("-> DB Operation Result - %s \n", connectionIP)

	dbPodName, err := util.GetPodNameFromListOfPods(appNS, dbName)
	require.NoErrorf(t, err, "<<There are no pods running backing db instance in the cluster as the result is - %s >>", dbPodName)
	require.Containsf(t, dbPodName, dbName, "list does not contain %s db pod from the list of pods running db instance in the cluster", dbPodName)
	t.Logf("-> Pod name is %s \n", dbPodName)

}

func CreateServiceBindingRequest(t *testing.T) {
	var (
		sbr string
		err error
	)
	checkClusterAvailable(t)

	t.Log("Creating Backing DB instance for backing service to connect with the app...")
	resCreateSBR := util.GetOutput(util.Run("make", "create-service-binding-request"))
	resExp := strings.TrimSpace(strings.Split(resCreateSBR, "servicebindingrequest.apps.openshift.io/binding-request")[1])
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while creating service binding request", resExp)
	fmt.Printf("Result of creating service binding request is %s \n", resExp)

	t.Log("Creating service binding request to bind nodejs app to connect with the backing service which is database...")
	sbr, err = util.GetCmdResult("", "oc", "get", "sbr", "-n", appNS, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NoErrorf(t, err, "<<Unable to get the sbr name as the result is - %s >>", sbr)
	require.NotEmptyf(t, sbr, "There are number of service binding request listed in a cluster %s \n", sbr)
	require.Contains(t, resCreateSBR, sbr, "Service binding request name is %d \n", sbr)
	t.Logf("-> Service binding request name is %s \n", sbr)

	t.Log("Fetching the status of binding")
	sbrStatus, err := util.GetCmdResult("True", "oc", "get", "sbr", sbr, "-n", appNS, "-o", `jsonpath={.status.conditions[*].status}`)
	require.NoErrorf(t, err, "<<Unable to get the sbr status as the result is - %s >>", sbrStatus)
	require.Contains(t, sbrStatus, "True", "service binding status is %d \n", sbrStatus)
	t.Logf("-> service binding status is %s \n", sbrStatus)

	annotation, err := util.GetCmdResult("", "oc", "get", "sbr", sbr, "-n", appNS, "-o", `jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io\/last-applied-configuration}'`)
	require.NoErrorf(t, err, "<<Unable to get the annotation as the result is - %s >>", annotation)
	t.Logf("-> Annotation is %s \n", annotation)
	actSBRResponse := &v1alpha1.ServiceBindingRequest{}
	err = util.UnmarshalJSONData(annotation, actSBRResponse)
	require.NoError(t, err, "Unable to unmarshall SBR from JSON")

	expResource := "deployments"
	expKind := "Database"

	require.Equal(t, appNS, actSBRResponse.ObjectMeta.Namespace, "SBR Namespace is not matched, As expected namespace is %d and actual namespace is %d\n", appNS, actSBRResponse.ObjectMeta.Namespace)
	require.Equal(t, sbr, actSBRResponse.ObjectMeta.Name, "SBR Name is not matched, As expected name is %d and actual name is %d\n", sbr, actSBRResponse.ObjectMeta.Name)
	require.Equal(t, expResource, actSBRResponse.Spec.ApplicationSelector.Resource, "SBR application resource is not matched, As expected application resource is %d and actual application resource is %d\n", expResource, actSBRResponse.Spec.ApplicationSelector.Resource)
	require.Equal(t, appName, actSBRResponse.Spec.ApplicationSelector.ResourceRef, "SBR application resource ref is not matched, As expected application resource ref is %d and actual application resource ref  is %d\n", appName, actSBRResponse.Spec.ApplicationSelector.ResourceRef)
	require.Equal(t, expKind, actSBRResponse.Spec.BackingServiceSelector.Kind, "SBR BackingServiceSelector kind is not matched, As expected BackingServiceSelector kind is %d and actual BackingServiceSelector kind  is %d\n", expKind, actSBRResponse.Spec.BackingServiceSelector.Kind)
	require.Equal(t, dbName, actSBRResponse.Spec.BackingServiceSelector.ResourceRef, "SBR BackingServiceSelector ResourceRef is not matched, As expected BackingServiceSelector ResourceRef is %d and actual BackingServiceSelector ResourceRef  is %d\n", dbName, actSBRResponse.Spec.BackingServiceSelector.ResourceRef)

	t.Log(" Fetching the route of an app ")
	route, err := util.GetCmdResult("", "oc", "get", "route", appName, "-n", appNS, "-o", `jsonpath={.status.ingress[0].host}`)
	require.NoErrorf(t, err, "<<Unable to get the route as the result is - %s >>", route)
	t.Logf("-> ROUTE - %s \n", route)

	appStatusEndpoint := fmt.Sprintf("http://%s/api/status/dbNameCM", route)
	checkNodeJSAppFrontend(t, appStatusEndpoint, dbName)

	t.Log("Fetching the details of secret")
	secret, err := util.GetCmdResult("", "oc", "get", "sbr", sbr, "-n", appNS, "-o", `jsonpath={.status.secret}`)
	require.NoErrorf(t, err, "<<Unable to get the secret as the result is - %s >>", secret)
	require.Contains(t, secret, sbr, "service binding secret detail -> %d \n", secret)
	t.Logf("-> service binding secret detail -> %s \n", secret)

	t.Log("Fetching the env details of binding")
	env, err := util.GetCmdResult("", "oc", "get", "deploy", appName, "-n", appNS, "-o", `jsonpath={.spec.template.spec.containers[0].env}`)
	require.NoErrorf(t, err, "<<Unable to get the env result as the result is - %s >>", env)
	require.Contains(t, env, "ServiceBindingOperator", "service binding env detail -> %d \n", env)
	t.Logf("-> service binding env detail ->%s \n", env)

	t.Log("Fetching the envFrom details of binding")
	envFrom, err := util.GetCmdResult("", "oc", "get", "deploy", appName, "-n", appNS, "-o", `jsonpath={.spec.template.spec.containers[0].envFrom}`)
	require.NoErrorf(t, err, "<<Unable to get the envFrom result as the result is - %s >>", envFrom)
	require.Contains(t, envFrom, sbr, "service binding envFrom detail -> %d \n", envFrom)
	t.Logf("-> service binding envFrom detail -> %s \n", envFrom)
}

func checkClusterAvailable(t *testing.T) {
	if !clusterAvailable {
		t.Skip("Cluster is not available, skipping")
	}
}

func checkNodeJSAppFrontend(t *testing.T, startURL string, expectedResponse string) {
	bodyStr := ""
	err := wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		resp, err := http.Get(startURL)
		require.NoErrorf(t, err, "Unable to get the application status at %s.")
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		bodyStr = string(body)
		require.NoError(t, err, "Failed to get the response body.")
		return strings.Contains(bodyStr, expectedResponse), err
	})
	require.NoErrorf(t, err, "Unexpected application status response: '%s'", bodyStr)
}

func DeleteProject(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("Deleting a project from the cluster...")
	res := util.GetOutput(util.Run("make", "delete-project"))
	require.NotEmptyf(t, res, "Project unable to delete because of - %s \n", res)

	expectedDeleteProjectResult, err := regexp.Compile(`.` + appNS + `.\sdeleted.*`)
	require.NoErrorf(t, err, "<<Unable to compile regular expression to delete project result - %s >>", expectedDeleteProjectResult)
	deleteProjectRes, checkFlag := util.GetCreateDeleteProjectResult(res, expectedDeleteProjectResult)
	require.Equal(t, true, checkFlag, "Unable to delete the project as the value is %s", deleteProjectRes)
	deleteStatus := appNS + `" deleted`
	require.Containsf(t, deleteProjectRes, deleteStatus, "Unable to match the project delete status because of %s", deleteProjectRes)
	t.Logf("-> Project deleted - %s \n", deleteProjectRes)
	fmt.Printf("-> Project deleted - %s \n", deleteProjectRes)

	expectedOCGetProjectMsg := appNS + `" not found`
	deleteProjectOcGetResult, err := util.GetCmdResultWithTimeout(expectedOCGetProjectMsg, 5*time.Second, 5*time.Minute, "oc", "get", "project", appNS, "-o", `jsonpath={.metadata.name}`)
	require.NoErrorf(t, err, "<<Error while getting the name of the project while deleting the project as the result is - %s >>", deleteProjectOcGetResult)
	t.Logf("-> OC get Project delete status - %s \n", deleteProjectOcGetResult)
	require.Containsf(t, expectedOCGetProjectMsg, deleteProjectOcGetResult, "OC get project after project delete status is %d \n", deleteProjectOcGetResult)
	fmt.Printf("-> After Project deleted, oc get project gives this message - %s \n", deleteProjectOcGetResult)
}

func UninstallBackingServiceOperator(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Uninstalling backing service operator from the cluster...")
	resultMakeUninstallBackingSvcOpr := util.GetOutput(util.Run("make", "uninstall-backing-db-operator"))

	expBackingSvcSubDelete, err := regexp.Compile(`.*subscription.operators.coreos.com\s.` + pkgManifest + `.\sdeleted.*`)
	require.NoErrorf(t, err, "<<Error while applying regular expression to uninstall backing svc subscription - %s >>", expBackingSvcSubDelete)

	expBackingSvcSrcDelete, err := regexp.Compile(`.*operatorsource.operators.coreos.com\s.` + pkgManifest + `.\sdeleted.*`)
	require.NoErrorf(t, err, "<<Error while applying regular expression to uninstall backing svc - %s >>", expBackingSvcSrcDelete)

	matchedBackingSvcSubDelete, checkFlag := util.GetRegExMatch(resultMakeUninstallBackingSvcOpr, expBackingSvcSubDelete)
	require.Equal(t, true, checkFlag, "Unable to match the uninstall status for backing svc subscription as the value is %s", matchedBackingSvcSubDelete)
	require.Containsf(t, resultMakeUninstallBackingSvcOpr, matchedBackingSvcSubDelete, "Unable to uninstall the backing svc subscription because of %s", matchedBackingSvcSubDelete)

	matchedBackingSvcDelete, checkFlag := util.GetRegExMatch(resultMakeUninstallBackingSvcOpr, expBackingSvcSrcDelete)
	require.Equal(t, true, checkFlag, "Unable to match the uninstall status for backing svc source as the value is %s", matchedBackingSvcDelete)
	require.Containsf(t, resultMakeUninstallBackingSvcOpr, matchedBackingSvcDelete, "Unable to uninstall the backing svc source because of %s", matchedBackingSvcDelete)

	t.Log("get the package manifest for db-operators after uninstalling backingservice operator...\n")
	manifest, err := util.GetCmdResult(pkgManifest, "oc", "get", "packagemanifest", pkgManifest, "-n", operatorsNS, "-o", `jsonpath={.metadata.name}`)
	require.NoErrorf(t, err, "<<Error while getting the name of the manifest after backing svc uninstallation as the result is - %s >>", manifest)
	expectedOCGetManifestCMD := pkgManifest + `" not found`
	require.Containsf(t, manifest, expectedOCGetManifestCMD, "OC get packagemanifest after backing svc uninstall status is %d \n", manifest)
	fmt.Printf("-> After Uninstall backing svc, oc get packagemanifest gives this message - %s \n", manifest)

	t.Log("Fetching the subscription name after uninstalling backing service operator")
	deleteBackingSvcOcGetManifestResult, err := util.GetCmdResult("", "oc", "get", "subscription", pkgManifest, "-n", operatorsNS, "-o", `jsonpath={.status.installplan.name}`)
	require.NoErrorf(t, err, "<<Error while getting the name of the install plan from subscription after backing svc Subscription uninstallation as the result is - %s >>", deleteBackingSvcOcGetManifestResult)
	require.Containsf(t, deleteBackingSvcOcGetManifestResult, expectedOCGetManifestCMD, "OC get subscription after backing svc subscription uninstall status is %d \n", deleteBackingSvcOcGetManifestResult)
	fmt.Printf("-> After Uninstall backing svc, oc get subscription gives this message - %s \n", deleteBackingSvcOcGetManifestResult)

	t.Log("Fetching the status of running pod after uninstalling backing service operator")
	expectedOCGetPodMsg := podNameBackingSvcOpr + `" not found`
	deleteBackingSvcOprGetPodResult, err := util.GetCmdResultWithTimeout(expectedOCGetPodMsg, 5*time.Second, 5*time.Minute, "oc", "get", "pod", podNameBackingSvcOpr, "-n", operatorsNS, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Error while getting the pod after uninstalling backing service operator as the result is - %s >>", deleteBackingSvcOprGetPodResult)
	require.Containsf(t, deleteBackingSvcOprGetPodResult, expectedOCGetPodMsg, "OC get pod after backing service operator uninstall status is %d \n", deleteBackingSvcOprGetPodResult)
	fmt.Printf("-> After Uninstall backing service operator, oc get pod gives this message - %s \n", deleteBackingSvcOprGetPodResult)

	/*i := 1
	for {
		fmt.Printf("-> Iteration %d , looking for message other than Running in oc get pod status for backing svc opr", i)
		deleteBackingSvcOprGetPodResult, _ := util.GetCmdResult("", "oc", "get", "pod", podNameBackingSvcOpr, "-n", operatorsNS, "-o", `jsonpath={.status.phase}`)
		if deleteBackingSvcOprGetPodResult != "Running" || i >= 10 {
			break
		}
		i++
	}*/

}

func UninstallServiceBindingOperator(t *testing.T) {

	checkClusterAvailable(t)

	oprSrc := "redhat-developer-operators"

	t.Log("Uninstalling serivice binding operator from the cluster...")
	resultMakeUninstallSBO := util.GetOutput(util.Run("make", "uninstall-service-binding-operator-master"))

	expSBOSubDelete, err := regexp.Compile(`.*subscription.operators.coreos.com\s.` + oprName + `.\sdeleted.*`)
	require.NoErrorf(t, err, "<<Error while applying regular expression to uninstall SBO subscription - %s >>", expSBOSubDelete)

	expSBODelete, err := regexp.Compile(`.*operatorsource.operators.coreos.com\s.` + oprSrc + `.\sdeleted.*`)
	require.NoErrorf(t, err, "<<Error while applying regular expression to uninstall SBO - %s >>", expSBODelete)

	matchedSBOSubDelete, checkFlag := util.GetRegExMatch(resultMakeUninstallSBO, expSBOSubDelete)
	require.Equal(t, true, checkFlag, "Unable to match the uninstall status for SBO subscription as the value is %s", matchedSBOSubDelete)
	require.Containsf(t, resultMakeUninstallSBO, matchedSBOSubDelete, "Unable to uninstall the SBO subscription because of %s", matchedSBOSubDelete)

	matchedSBODelete, checkFlag := util.GetRegExMatch(resultMakeUninstallSBO, expSBODelete)
	require.Equal(t, true, checkFlag, "Unable to match the uninstall status for SBO source as the value is %s", matchedSBODelete)
	require.Containsf(t, resultMakeUninstallSBO, matchedSBODelete, "Unable to uninstall the SBO source because of %s", matchedSBODelete)

	deleteSBOOcGetResult, err := util.GetCmdResult("", "oc", "get", "subscription", oprName, "-n", operatorsNS, "-o", `jsonpath={.status.installplan.name}`)
	require.NoErrorf(t, err, "<<Error while getting the name of the install plan from subscription after SBO Subscription uninstallation as the result is - %s >>", deleteSBOOcGetResult)
	expectedOCGetOprCMD := oprName + `" not found`
	require.Containsf(t, deleteSBOOcGetResult, expectedOCGetOprCMD, "OC get subscription after uninstall SBO subscription status is %d \n", deleteSBOOcGetResult)
	fmt.Printf("-> After Uninstall SBO, oc get subscription gives this message - %s \n", deleteSBOOcGetResult)

	t.Log("Fetching the status of running pod after uninstall")
	expectedOCGetPodMsg := podNameSBO + `" not found`
	deleteSBOGetPodResult, err := util.GetCmdResultWithTimeout(expectedOCGetPodMsg, 5*time.Second, 5*time.Minute, "oc", "get", "pod", podNameSBO, "-n", operatorsNS, "-o", `jsonpath={.status.phase}`)
	require.NoErrorf(t, err, "<<Error while getting the pod running SBO after uninstalling SBO as the result is - %s >>", deleteSBOGetPodResult)
	require.Containsf(t, deleteSBOGetPodResult, expectedOCGetPodMsg, "OC get pod after SBO uninstall status is %d \n", deleteSBOGetPodResult)
	fmt.Printf("-> After Uninstall SBO, oc get pod gives this message - %s \n", deleteSBOGetPodResult)

	/*
		i := 1
		for {
			fmt.Printf("-> Iteration %d , looking for message other than Running in oc get pod status for sbo ", i)
			deleteSBOGetPodResult, _ := util.GetCmdResult("", "oc", "get", "pod", podNameSBO, "-n", operatorsNS, "-o", `jsonpath={.status.phase}`)
			if deleteSBOGetPodResult != "Running" || i >= 10 {
				break
			}
			i++
		}*/

}
