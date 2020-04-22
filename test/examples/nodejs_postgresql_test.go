package examples_test

import (
	"fmt"
	"log"

	"strings"
	"testing"
	"time"

	"github.com/redhat-developer/service-binding-operator/test/examples/util"
	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"

	"io/ioutil"

	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	exampleName = "nodejs_postgresql"
	ns          = "openshift-operators"
	oprName     = "service-binding-operator"
	pjt         = "service-binding-demo"

	nodeJsApp = "https://github.com/pmacik/nodejs-rest-http-crud"
	appName   = "nodejs-rest-http-crud"

	ipName, ipStatus, podName, podStatus, dbOprRes string
	expBuildPodName, dc, bc                        string
	clusterAvailable                               = false
	checkFlag                                      bool

	pkgManifest = "db-operators"
	bckSvc      = "postgresql-operator"
	dbName      = "db-demo"
)

const (
	retryInterval = 1 * time.Second
	retryTimeout  = 10 * time.Second
)

func TestNodeJSPostgreSQL(t *testing.T) {

	t.Run("set-example-dir", SetExampleDir)
	t.Run("get-oc-status", GetOCStatus)

	//t.Run("install-service-binding-operator", MakeInstallServiceBindingOperator)
	//t.Run("install-backing-service-operator", MakeInstallBackingServiceOperator)
	t.Run("create-project", CreatePorject)
	t.Run("import-nodejs-app", ImportNodeJSApp)

	//Comment this once https://github.com/openshift/oc/pull/355 is fixed
	t.Run("use-deployment", UseDeployment)
	t.Run("create-backing-db-instance", CreateBackingDbInstance)
	t.Run("createservice-binding-request", CreateServiceBindingRequest)

	//t.Run("test-yaml", TestYaml)

}

func UseDeployment(t *testing.T) {

	t.Log(" Delete the deployment config ")
	deletedStatus := util.GetCmdResult("", "oc", "delete", "dc", dc, "-n", pjt)
	require.Containsf(t, deletedStatus, "deleted", "Deployment config is deleted with the message %d \n", deletedStatus)
	t.Logf("-> Deployment config is deleted with the message %s \n", deletedStatus)

	buildPods := util.GetCmdResult("", "oc", "get", "pods", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	t.Logf(" List of pods running in the cluster - %s", buildPods)
	t.Log(" Fetching the build pod name from the list of pods ")

	checkFlag, deploymentBuildPodName := util.GetPodNameFromLst(buildPods, expBuildPodName)
	require.NotEqual(t, true, checkFlag, "List does not contain the pod")
	require.NotContainsf(t, deploymentBuildPodName, expBuildPodName, "list does not contain %s build pod from the list of pods running builds in the cluster", expBuildPodName)
	t.Logf("-> list does not contain %s build pod from the list of pods running builds in the cluster \n", expBuildPodName)

	deploymentData := util.GetCmdResult("", "oc", "apply", "-f", "deployment.yaml")
	t.Logf("-> Deployment config is deleted with the message %s \n", deploymentData)

	buildPods = util.GetCmdResult("", "oc", "get", "pods", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NotEmptyf(t, buildPods, "", "There are number of build pods listed in a cluster %s \n", buildPods)
	t.Logf(" List of pods running in the cluster - %s", buildPods)
	t.Log(" Fetching the build pod name from the list of pods ")

	checkFlag, deploymentPodName := util.GetPodNameFromLst(buildPods, bc)
	require.Equal(t, true, checkFlag, "List does not contain the pod")
	require.Containsf(t, deploymentPodName, bc, "list does not contain %s build pod from the list of pods running builds in the cluster", deploymentPodName)
	t.Logf("-> deployment build Pod name is %s \n", deploymentPodName)

	t.Log("Fetching the name of deployment")
	deployment := util.GetCmdResult("", "oc", "get", "deployment", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	require.Equal(t, bc, deployment, "Deployment name is %d \n", deployment)
	t.Logf("-> Deployment name is %s \n", deployment)

	t.Log("Fetching the status of deployment ")
	deploymentStatus := util.GetCmdResult("True", "oc", "get", "deployment", deployment, "-n", pjt, "-o", `jsonpath={.status.conditions[*].status}`)
	require.Contains(t, deploymentStatus, "True", "Deployment status is %d \n", deploymentStatus)
	t.Logf("-> Deployment status is %s \n", deploymentStatus)

	//envStatus := util.GetCmdResult("", "oc", "get", "deployment", deployment, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].env}`)
	//t.Logf("-> Deployment status is %s \n", envStatus)
	//envFromStatus := util.GetCmdResult("", "oc", "get", "deployment", deployment, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].envFrom}`)
	//t.Logf("-> Deployment status is %s \n", envFromStatus)
}

func TestYaml(t *testing.T) {

	deployment := appsv1.Deployment{}
	data, err := ioutil.ReadFile("/home/shobith/rh/pjts/serviceBindingOperator/src/github.com/redhat-developer/service-binding-operator/examples/nodejs_postgresql/deployment.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal(data, &deployment)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%v", deployment)

}

//SetExampleDir tests that the corrent example directory was set as a working directory for running the commands.
func SetExampleDir(t *testing.T) {
	examplePath := fmt.Sprintf("%s/%s", util.GetExamplesDir(), exampleName)
	util.SetDir(examplePath)
	res := strings.TrimSpace(util.Run("pwd").Stdout())
	require.Equal(t, examplePath, res)
}

//Logs the oc status
func GetOCStatus(t *testing.T) {
	t.Log("--- Getting OC Status ---")
	util.GetOutput(util.Run("oc", "status"), "oc status")
	clusterAvailable = true
	t.Log(" *** Connected to cluster *** ")
}

//Logs out the output of command make install-service-binding-operator
func MakeInstallServiceBindingOperator(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("Installing serivice binding operator into the cluster...")
	res := util.GetOutput(util.Run("make", "install-service-binding-operator-master"), "make install-service-binding-operator-master")
	resExp := strings.TrimSpace(strings.Split(res, "subscription.operators.coreos.com/service-binding-operator")[1])
	fmt.Printf("subscription output is %s \n", resExp)
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while installing service binding operator", resExp)

	// with openshift-operators namespace, capture the install plan
	t.Log(" Fetching the status of install plan name ")
	ipName = util.GetCmdResult("", "oc", "get", "subscription", oprName, "-n", ns, "-o", `jsonpath={.status.installplan.name}`)
	t.Logf("-> Install plan-ip name is %s \n", ipName)

	//// with openshift-operators namespace, capture the install plan status of db-operators
	t.Log(" Fetching the status of install plan status ")
	ipStatus = util.GetCmdResult("Complete", "oc", "get", "ip", "-n", ns, ipName, "-o", `jsonpath={.status.phase}`)
	t.Logf("-> Install plan-ip name is %s \n", ipStatus)

	//oc get pods -n openshift-operator
	t.Log("Fetching the pod name of the running pod")
	pods := util.GetCmdResult("", "oc", "get", "pods", "-n", ns, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NotEmptyf(t, pods, "", "There are number of pods listed in a cluster %s \n", pods)

	checkFlag, podName := util.GetPodNameFromLst(pods, oprName)
	if checkFlag {
		require.Containsf(t, podName, oprName, "list does not contain %s pod from the list of pods running service binding operator in the cluster", resExp)
		t.Logf("-> Pod name is %s \n", podName)
	}

	t.Log("Fetching the status of running pod")
	podStatus = util.GetCmdResult("Running", "oc", "get", "pod", podName, "-n", ns, "-o", `jsonpath={.status.phase}`)
	t.Logf("-> Pod name is %s \n", podStatus)
	require.Equal(t, "Running", podStatus, "'pod plan status' is %d \n", podStatus)

}

func MakeInstallBackingServiceOperator(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Installing backing service operator source into the cluster...")
	res := util.GetOutput(util.Run("make", "install-backing-db-operator-source"), "make install-backing-db-operator-source")
	//resExp := strings.TrimSpace(strings.Split(strings.Split(res, "operatorsource.operators.coreos.com/db-operators")[1], "Waiting")[0])
	resExp := strings.TrimSpace(strings.Split(res, "operatorsource.operators.coreos.com/db-operators")[1])
	fmt.Printf("db operator installation source output is %s \n", resExp)
	require.True(t, strings.Contains(resExp, "created") || strings.Contains(resExp, "unchanged"), "Output %s does not contain created or unchanged while installing db service operator source", resExp)

	//with this command 'oc get packagemanifest | grep db-operators' make sure there is an entry to the package manifest
	t.Log("get the package manifest for db-operators...\n")
	manifest := util.GetCmdResult(pkgManifest, "oc", "get", "packagemanifest", pkgManifest, "-o", `jsonpath={.metadata.name}`)
	t.Logf("-> %s has an entry in the package manifest \n", manifest)
	require.NotEmptyf(t, manifest, "", "There are number of manifest listed in a cluster %s \n", manifest)

	//Install the subscription using this command: make install-backing-db-operator-subscription
	t.Log("Installing backing service operator subscription into the cluster...")
	dbOprRes := util.GetOutput(util.Run("make", "install-backing-db-operator-subscription"), "make install-backing-db-operator-subscription")
	subRes := strings.TrimSpace(strings.Split(dbOprRes, "subscription.operators.coreos.com/db-operators")[1])
	t.Logf("subscription output is %s \n", subRes)

	// with openshift-operators namespace, capture the install plan
	ipName = util.GetCmdResult("", "oc", "get", "subscription", pkgManifest, "-n", ns, "-o", `jsonpath={.status.installplan.name}`)
	t.Logf("-> Pod name is %s \n", ipName)

	//// with openshift-operators namespace, capture the install plan status of db-operators
	t.Log(" Fetching the status of install plan ")
	ipStatus = util.GetCmdResult("Complete", "oc", "get", "ip", "-n", ns, ipName, "-o", `jsonpath={.status.phase}`)
	t.Logf("-> Pod name is %s \n", ipStatus)
	require.Equal(t, ipStatus, "Complete", "'install plan status' is %d \n", ipStatus)

	//oc get pods -n openshift-operator
	t.Log("Fetching the pod name of the running pod")
	pods := util.GetCmdResult("", "oc", "get", "pods", "-n", ns, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NotEmptyf(t, pods, "", "There are number of pods listed in a cluster %s \n", pods)

	checkFlag, podName := util.GetPodNameFromLst(pods, bckSvc)
	if checkFlag {
		require.Containsf(t, podName, bckSvc, "list does not contain %s pod from the list of pods running service binding operator in the cluster", bckSvc)
		t.Logf("-> Pod name is %s \n", podName)
	}
	//oc get pod <<Name of pod(from step 4)>> -n openshift-operators -o jsonpath='{.status.phase}'
	t.Log("Fetching the status of running pod")
	podStatus := util.GetCmdResult("Running", "oc", "get", "pod", podName, "-n", ns, "-o", `jsonpath={.status.phase}`)
	require.Equal(t, "Running", podStatus, "pod status is %d \n", podStatus)

	//oc get crd | grep database
	t.Log("Checking the backing service's CRD is installed")
	crd := util.GetCmdResult("databases.postgresql.baiju.dev", "oc", "get", "crd", "databases.postgresql.baiju.dev")
	require.NotEmpty(t, crd, "packing service CRD not found")
}

func CreatePorject(t *testing.T) {

	var createPjtRes string

	checkClusterAvailable(t)

	t.Log("Creating a project into the cluster...")
	res := util.GetOutput(util.Run("make", "create-project"), "make create-project")
	require.NotEmptyf(t, res, "", "Pjt not created because of this - %s \n", res)
	createPjtRes = util.GetPjtCreationRes(res, pjt)
	require.Containsf(t, createPjtRes, pjt, "Pjt - %s is not created because of %s", pjt, res)

	t.Logf("-> Project created - %s \n", createPjtRes)
	//with this command 'oc get project service-binding-demo -o jsonpath='{.metadata.name}' creates an new project
	t.Log("Get the name of project added to the cluster...\n")
	getPjt := util.GetCmdResult("", "oc", "get", "project", pjt, "-o", `jsonpath={.metadata.name}`)

	t.Logf("-> Project created - %s \n", getPjt)
	require.Equal(t, getPjt, pjt, "Pjt created is %d \n", getPjt)

	//with this command 'oc get project service-binding-demo -o jsonpath='{.status.phase}' creates an new project
	t.Log("Get the status of the project added to the cluster...\n")
	pjtStatus := util.GetCmdResult("Active", "oc", "get", "project", pjt, "-o", `jsonpath={.status.phase}`)

	t.Logf("-> Project created %s has the status %s \n", pjt, pjtStatus)
	require.Equal(t, "Active", pjtStatus, "Pjt status is %d \n", pjtStatus)

	t.Log(" Setting the project to service-binding-demo ")
	pjtRes := util.GetCmdResult("", "oc", "project", pjt)
	t.Logf("Project/Namespace set is %s", pjtRes)
}

func ImportNodeJSApp(t *testing.T) {

	checkClusterAvailable(t)

	var name, svc string

	t.Log("import an nodejs app to bind with backing service...")

	nodeJsAppArg := "nodejs~" + nodeJsApp
	t.Log(" Fetching the build config name of the app ")
	app := util.GetCmdResult("", "oc", "new-app", nodeJsAppArg, "--name", appName, "-n", pjt)
	//require.Contains(t, appName, bc, "Build config name is %d \n", bc)
	t.Logf("-> App info - %s \n", app)

	t.Log(" Fetching the build config name of the app ")
	bc = util.GetCmdResult("", "oc", "get", "bc", "-o", `jsonpath={.items[*].metadata.name}`)
	require.Contains(t, appName, bc, "Build config name is %d \n", bc)

	t.Log(" Fetching the buid info - name of the app ")
	buildName := util.GetCmdResult("", "oc", "get", "build", "-o", `jsonpath={.items[*].metadata.name}`)
	require.Contains(t, buildName, bc, "Build config name is %d \n", buildName)
	t.Logf(" Buid name of an app %s", buildName)

	t.Log(" Fetching the buid status of the app ")
	buildStatus := util.GetCmdResult("Complete", "oc", "get", "build", buildName, "-o", `jsonpath={.status.phase}`)
	require.Equal(t, "Complete", buildStatus, "Build status of config name %d is %d \n", buildName, buildStatus)
	t.Logf(" Buid status of an app %s", buildStatus)
	t.Log(" Fetching the list of pod for the build of an app ")
	expBuildPodName = buildName + "-build"
	buildPods := util.GetCmdResult("", "oc", "get", "pods", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NotEmptyf(t, buildPods, "", "There are number of build pods listed in a cluster %s \n", buildPods)
	t.Logf(" List of pods running in the cluster - %s", buildPods)
	t.Log(" Fetching the build pod name from the list of pods ")
	checkFlag, buildPodName := util.GetPodNameFromLst(buildPods, expBuildPodName)
	require.Equal(t, true, checkFlag, "List does not contain the pod")
	require.Containsf(t, expBuildPodName, buildPodName, "list does not contain %s build pod from the list of pods running builds in the cluster", buildPodName)
	t.Logf("-> Pod name is %s \n", buildPodName)

	t.Log("Fetching the status of running build pod")
	buildPodStatus := util.GetCmdResult("Succeeded", "oc", "get", "pods", buildPodName, "-n", pjt, "-o", `jsonpath={.status.phase}`)
	require.Equal(t, "Succeeded", buildPodStatus, "Build pod status is %d \n", buildPodStatus)

	t.Log("Fetching the deployment resource pod name from the list of pods")
	expBuildPodName = bc + "-1-deploy"
	checkFlag, deploymentBuildPodName := util.GetPodNameFromLst(buildPods, expBuildPodName)
	require.Equal(t, true, checkFlag, "List does not contain the pod")
	require.Containsf(t, deploymentBuildPodName, expBuildPodName, "list does not contain %s build pod from the list of pods running builds in the cluster", deploymentBuildPodName)
	t.Logf("-> deployment build Pod name is %s \n", deploymentBuildPodName)

	t.Log("Fetching the status of running deployment resource pod")
	deploymentPodStatus := util.GetCmdResult("Succeeded", "oc", "get", "pods", deploymentBuildPodName, "-n", pjt, "-o", `jsonpath={.status.phase}`)
	require.Equal(t, "Succeeded", deploymentPodStatus, "Build pod status is %d \n", deploymentPodStatus)
	t.Logf("-> Deployment pods status is %s \n", deploymentPodStatus)

	t.Log("Fetching the name of deployment config")
	dc = util.GetCmdResult("", "oc", "get", "dc", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	require.Equal(t, bc, dc, "DeploymentConfig name is %d \n", dc)
	t.Logf("-> Deployment Config name is %s \n", dc)

	t.Log("Fetching the status of deployment config")
	dcStatus := util.GetCmdResult("True", "oc", "get", "dc", dc, "-n", pjt, "-o", `jsonpath={.status.conditions[*].status}`)
	require.Equal(t, "True True", dcStatus, "DeploymentConfig status is %d \n", dcStatus)
	t.Logf("-> Deployment Config status is %s \n", dcStatus)

	//oc expose svc/nodejs-rest-http-crud --name=nodejs-rest-http-crud
	t.Log(" Exposing an app ")
	svc = "svc/" + bc
	name = "--name=" + bc
	exposeRes := util.GetCmdResult("", "oc", "expose", svc, name)
	t.Logf("app exposed result is %s", exposeRes)

	t.Log(" Fetching the route of an app ")
	route := util.GetCmdResult("", "oc", "get", "route", bc, "-n", pjt, "-o", `jsonpath={.status.ingress[0].host}`)
	t.Logf("-> ROUTE - %s \n", route)

	host := "http://" + route
	checkNodeJSAppFrontend(t, host, "(DB: N/A)")

}

func CreateBackingDbInstance(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Creating Backing DB instance for backing service to connect with the app...")
	res := util.GetOutput(util.Run("make", "create-backing-db-instance"), "make create-backing-db-instance")
	resExp := strings.TrimSpace(strings.Split(res, dbName)[1])
	fmt.Printf("Result of created backing DB instance is %s \n", resExp)
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while installing db instance", resExp)

	//oc get db -o jsonpath={.items[*].metadata.name}
	t.Log("Fetching the name of running db instance")
	dbInstanceName := util.GetCmdResult("", "oc", "get", "db", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	require.Equal(t, dbName, dbInstanceName, "db instance name is %d \n", dbInstanceName)
	t.Logf("-> DB instance name is %s \n", dbInstanceName)

	//oc get db db-demo -o jsonpath='{.status.dbConnectionIP}'
	connectionIP := util.GetCmdResult("", "oc", "get", "db", dbInstanceName, "-o", "jsonpath={.status.dbConnectionIP}")
	t.Logf("-> DB Operation Result - %s \n", connectionIP)

	buildPods := util.GetCmdResult("", "oc", "get", "pods", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NotEmptyf(t, buildPods, "", "There are number of build pods listed in a cluster %s \n", buildPods)
	t.Logf(" List of pods running in the cluster - %s", buildPods)
	t.Log(" Fetching the build pod name from the list of pods ")

	checkFlag, dbPodName := util.GetPodNameFromLst(buildPods, dbName)
	require.Equal(t, true, checkFlag, "List does not contain the pod")
	require.Containsf(t, dbPodName, dbName, "list does not contain %s db pod from the list of pods running db instance in the cluster", dbPodName)
	t.Logf("-> Pod name is %s \n", dbPodName)

}

func CreateServiceBindingRequest(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Creating Backing DB instance for backing service to connect with the app...")
	resCreateSBR := util.GetOutput(util.Run("make", "create-service-binding-request"), "make create-service-binding-request")
	resExp := strings.TrimSpace(strings.Split(resCreateSBR, "servicebindingrequest.apps.openshift.io/binding-request")[1])
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while creating service binding request", resExp)
	fmt.Printf("Result of creating service binding request is %s \n", resExp)

	//oc get sbr -o jsonpath={.items[*].metadata.name}
	t.Log("Creating service binding request to bind nodejs app to connect with the backing service which is database...")
	sbr := util.GetCmdResult("", "oc", "get", "sbr", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NotEmptyf(t, sbr, "", "There are number of service binding request listed in a cluster %s \n", sbr)
	require.Contains(t, resCreateSBR, sbr, "Service binding request name is %d \n", sbr)
	t.Logf("-> Service binding request name is %s \n", sbr)

	t.Log("Fetching the status of binding")
	sbrStatus := util.GetCmdResult("True", "oc", "get", "sbr", sbr, "-n", pjt, "-o", `jsonpath={.status.conditions[*].status}`)
	require.Contains(t, sbrStatus, "True", "service binding status is %d \n", sbrStatus)
	t.Logf("-> service binding status is %s \n", sbrStatus)

	///oc get sbr binding-request -o jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io\/last-applied-configuration}'
	annotation := util.GetCmdResult("", "oc", "get", "sbr", sbr, "-n", pjt, "-o", `jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io\/last-applied-configuration}'`)
	t.Logf("-> Annotation is %s \n", annotation)
	actSBRResponse := util.UnmarshalJSONData(annotation)
	expSBRResponse := util.GetSbrResponse()
	expSBRResponse.Metadata.Name = sbr
	expSBRResponse.Metadata.Namespace = pjt
	expSBRResponse.Spec.ApplicationSelector.Resource = "deployments"
	expSBRResponse.Spec.ApplicationSelector.ResourceRef = bc
	expSBRResponse.Spec.BackingServiceSelector.Group = "postgresql.baiju.dev"
	expSBRResponse.Spec.BackingServiceSelector.Kind = "Database"
	expSBRResponse.Spec.BackingServiceSelector.ResourceRef = dbName
	//	require.Containsf(t, "true", reflect.DeepEqual(actSBRResponse, expSBRResponse), "structs not matching")

	require.Equal(t, expSBRResponse.Kind, actSBRResponse.Kind, "SBR kind is not matched, As expected kind is %d and actual kind is %d\n", expSBRResponse.Kind, actSBRResponse.Kind)
	require.Equal(t, expSBRResponse.Metadata.Namespace, actSBRResponse.Metadata.Namespace, "SBR Namespace is not matched, As expected namespace is %d and actual namespace is %d\n", expSBRResponse.Metadata.Namespace, actSBRResponse.Metadata.Namespace)
	require.Equal(t, expSBRResponse.Metadata.Name, actSBRResponse.Metadata.Name, "SBR Name is not matched, As expected name is %d and actual name is %d\n", expSBRResponse.Metadata.Name, actSBRResponse.Metadata.Name)
	require.Equal(t, expSBRResponse.Spec.ApplicationSelector.Resource, actSBRResponse.Spec.ApplicationSelector.Resource, "SBR application resource is not matched, As expected application resource is %d and actual application resource is %d\n", expSBRResponse.Spec.ApplicationSelector.Resource, actSBRResponse.Spec.ApplicationSelector.Resource)
	require.Equal(t, expSBRResponse.Spec.ApplicationSelector.ResourceRef, actSBRResponse.Spec.ApplicationSelector.ResourceRef, "SBR application resource ref is not matched, As expected application resource ref is %d and actual application resource ref  is %d\n", expSBRResponse.Spec.ApplicationSelector.ResourceRef, actSBRResponse.Spec.ApplicationSelector.ResourceRef)
	require.Equal(t, expSBRResponse.Spec.BackingServiceSelector.Kind, actSBRResponse.Spec.BackingServiceSelector.Kind, "SBR BackingServiceSelector kind is not matched, As expected BackingServiceSelector kind is %d and actual BackingServiceSelector kind  is %d\n", expSBRResponse.Spec.BackingServiceSelector.Kind, actSBRResponse.Spec.BackingServiceSelector.Kind)
	require.Equal(t, expSBRResponse.Spec.BackingServiceSelector.ResourceRef, actSBRResponse.Spec.BackingServiceSelector.ResourceRef, "SBR BackingServiceSelector ResourceRef is not matched, As expected BackingServiceSelector ResourceRef is %d and actual BackingServiceSelector ResourceRef  is %d\n", expSBRResponse.Spec.BackingServiceSelector.ResourceRef, actSBRResponse.Spec.BackingServiceSelector.ResourceRef)

	t.Log(" Fetching the route of an app ")
	route := util.GetCmdResult("", "oc", "get", "route", bc, "-n", pjt, "-o", `jsonpath={.status.ingress[0].host}`)
	t.Logf("-> ROUTE - %s \n", route)

	host := "http://" + route
	checkNodeJSAppFrontend(t, host, dbName)

	//secret := util.GetCmdResult("", "oc", "get", "sbr", servBindReq, "-n", pjt, "-o", `jsonpath={.status.secret}`)
	t.Log("Fetching the details of secret")
	secret := util.GetCmdResult("", "oc", "get", "sbr", sbr, "-n", pjt, "-o", `jsonpath={.status.secret}`)
	require.Contains(t, secret, sbr, "service binding secret detail -> %d \n", secret)
	t.Logf("-> service binding secret detail -> %s \n", secret)

	//env := util.GetCmdResult("", "oc", "get", "deploy", bc, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].env}`)
	t.Log("Fetching the status of binding")
	env := util.GetCmdResult("", "oc", "get", "deploy", bc, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].env}`)
	require.Contains(t, env, "ServiceBindingOperator", "service binding env detail -> %d \n", env)
	t.Logf("-> service binding env detail ->%s \n", env)

	//envFrom := util.GetCmdResult("", "oc", "get", "deploy", bc, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].envFrom}`)
	t.Log("Fetching the status of binding")
	envFrom := util.GetCmdResult("", "oc", "get", "deploy", bc, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].envFrom}`)
	require.Contains(t, envFrom, sbr, "service binding envFrom detail -> %d \n", envFrom)
	t.Logf("-> service binding envFrom detail -> %s \n", envFrom)
}

func checkClusterAvailable(t *testing.T) {
	if !clusterAvailable {
		t.Skip("Cluster is not available, skipping")
	}
}

func checkNodeJSAppFrontend(t *testing.T, startURL string, expectedHeader string) {
	wd, svc := initSelenium(t)

	defer svc.Stop()
	defer wd.Quit()

	err := wd.Get(startURL)
	require.NoErrorf(t, err, "Unable to open app page: %s", startURL)
	header := findElementBy(t, wd, selenium.ByTagName, "h1")

	headerText, err := header.Text()
	require.NoError(t, err, "Unable to get the page header")

	require.Contains(t, headerText, expectedHeader)
}

func initSelenium(t *testing.T) (selenium.WebDriver, *selenium.Service) {
	chromedriverPath := "chromedriver"
	chromedriverPort := 9515

	service, err := selenium.NewChromeDriverService(chromedriverPath, chromedriverPort)
	checkErr(t, err)

	chromeOptions := map[string]interface{}{
		"args": []string{
			"--no-cache",
			"--no-sandbox",
			"--headless",
			"--window-size=1920,1080",
			"--window-position=0,0",
		},
	}

	caps := selenium.Capabilities{
		"browserName":   "chrome",
		"chromeOptions": chromeOptions,
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", chromedriverPort))
	checkErr(t, err)
	return wd, service
}

// CheckErr checks for errors and logging it to log as Fatal if not nil
func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func findElementBy(t *testing.T, wd selenium.WebDriver, by string, selector string) selenium.WebElement {
	elem, err := wd.FindElement(by, selector)
	checkErr(t, err)
	return elem
}
