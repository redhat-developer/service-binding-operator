package examples_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/redhat-developer/service-binding-operator/test/examples/util"
	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

var (
	exampleName = "nodejs_postgresql"
	ns          = "openshift-operators"
	oprName     = "service-binding-operator"
	pjt         = "service-binding-demo"

	nodeJsApp = "https://github.com/pmacik/nodejs-rest-http-crud"
	appName   = "nodejs-rest-http-crud"

	ipName, ipStatus, podName, podStatus, dbOprRes string
	clusterAvailable                               = false
	checkFlag                                      bool

	pkgManifest = "db-operators"
	bckSvc      = "postgresql-operator"
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
	//t.Run("create-backing-db-instance", CreateBackingDbInstance)
	//t.Run("createservice-binding-request", CreateServiceBindingRequest)
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
	res := util.GetOutput(util.Run("make", "install-service-binding-operator"), "make install-service-binding-operator")
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
	bc := util.GetCmdResult("", "oc", "get", "bc", "-o", `jsonpath={.items[*].metadata.name}`)
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
	expBuildPodName := buildName + "-build"
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

	//t.Log(" Fetching the deployment name of the app ")
	//deploymentName := util.GetCmdResult("", "oc", "get", "deployment", "-n", pjt, "-o", `jsonpath={.items[*].metadata.name}`)
	//require.Equal(t, bc, deploymentName, "Deployment name is %d \n", deploymentName)
	//t.Logf("-> Deployment name is %s \n", deploymentName)

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

	/*t.Log(" Fetching the curl operation result using the route ")
	curlRes := util.GetExecCmdResult("", "curl", "-L", "--silent", "-XGET", host, "|", "grep", "h1")

	require.Containsf(t, "Node.js Crud Application (DB:", curlRes, "page does not contain %s ", curlRes)
	t.Logf("-> CURL Operation Result - %s \n", curlRes)*/

}

func CreateBackingDbInstance(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Creating Backing DB instance for backing service to connect with the app...")
	res := util.GetOutput(util.Run("make", "create-backing-db-instance"), "make create-backing-db-instance")
	resExp := strings.TrimSpace(strings.Split(res, "subscription.operators.coreos.com/service-binding-operator")[1])
	fmt.Printf("Result of created backing DB instance is %s \n", resExp)
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while installing service binding operator", resExp)

	//oc get db db-demo -o jsonpath='{.status.dbConnectionIP}'
	db := util.GetCmdResult("", "oc", "get", "db-demo", "-o", "jsonpath={.status.dbConnectionIP}")
	t.Logf("-> DB Operation Result - %s \n", db)

	t.Log("Fetching the status of running pod")
	dbPodStatus := util.GetCmdResult("Succeeded", "oc", "get", "pods", db, "-n", pjt, "-o", `jsonpath={.status.phase}`)
	require.Equal(t, "Succeeded", dbPodStatus, "Build pod status is %d \n", dbPodStatus)
	t.Logf("-> DB pods status is %s \n", dbPodStatus)

}

func CreateServiceBindingRequest(t *testing.T) {
	checkClusterAvailable(t)

	//dbPodStatus := util.GetCmdResult("", "oc", "get", "deploy", bc, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].env}`)
	//dbPodStatus := util.GetCmdResult("", "oc", "get", "deploy", bc, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].envFrom}`)

	t.Log("Creating Backing DB instance for backing service to connect with the app...")
	res := util.GetOutput(util.Run("make", "create-service-binding-request"), "make create-service-binding-request")
	resExp := strings.TrimSpace(strings.Split(res, "subscription.operators.coreos.com/service-binding-operator")[1])
	fmt.Printf("Result of created backing DB instance is %s \n", resExp)

	//sbrStatus := util.GetCmdResult("", "oc", "get", "sbr", servBindReq, "-n", pjt, "-o", `jsonpath={.status.bindingStatus}`)
	//secret := util.GetCmdResult("", "oc", "get", "sbr", servBindReq, "-n", pjt, "-o", `jsonpath={.status.secret}`)

	//env := util.GetCmdResult("", "oc", "get", "deploy", bc, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].env}`)
	//envFrom := util.GetCmdResult("", "oc", "get", "deploy", bc, "-n", pjt, "-o", `jsonpath={.spec.template.spec.containers[0].envFrom}`)

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
