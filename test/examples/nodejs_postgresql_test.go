package examples_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/redhat-developer/service-binding-operator/test/examples/util"
	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

var (
	exampleName = "nodejs_postgresql"
	operatorsNS = "openshift-operators"
	oprName     = "service-binding-operator"
	appNS       = "service-binding-demo"

	nodeJsApp = "https://github.com/pmacik/nodejs-rest-http-crud"
	appName   = "nodejs-rest-http-crud"

	clusterAvailable = false

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
	t.Run("create-backing-db-instance", CreateBackingDbInstance)
	t.Run("createservice-binding-request", CreateServiceBindingRequest)

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
	util.GetOutput(util.Run("oc", "status"))
	clusterAvailable = true
	t.Log(" *** Connected to cluster *** ")
}

//Logs out the output of command make install-service-binding-operator
func MakeInstallServiceBindingOperator(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("Installing serivice binding operator into the cluster...")
	res := util.GetOutput(util.Run("make", "install-service-binding-operator-master"))
	resExp := strings.TrimSpace(strings.Split(res, "subscription.operators.coreos.com/service-binding-operator")[1])
	fmt.Printf("subscription output is %s \n", resExp)
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while installing service binding operator", resExp)

	// with openshift-operators namespace, capture the install plan
	t.Log(" Fetching the status of install plan name ")
	ipName := util.GetCmdResult("", "oc", "get", "subscription", oprName, "-n", operatorsNS, "-o", `jsonpath={.status.installplan.name}`)
	t.Logf("-> Install plan-ip name is %s \n", ipName)

	//// with openshift-operators namespace, capture the install plan status of db-operators
	t.Log(" Fetching the status of install plan status ")
	ipStatus := util.GetCmdResult("Complete", "oc", "get", "ip", "-n", operatorsNS, ipName, "-o", `jsonpath={.status.phase}`)
	t.Logf("-> Install plan-ip name is %s \n", ipStatus)

	podName := util.GetPodNameFromListOfPods(operatorsNS, oprName)
	require.Containsf(t, podName, oprName, "list does not contain %s pod from the list of pods running service binding operator in the cluster", resExp)
	t.Logf("-> Pod name is %s \n", podName)

	t.Log("Fetching the status of running pod")
	podStatus := util.GetCmdResult("Running", "oc", "get", "pod", podName, "-n", operatorsNS, "-o", `jsonpath={.status.phase}`)
	t.Logf("-> Pod name is %s \n", podStatus)
	require.Equal(t, "Running", podStatus, "'pod plan status' is %d \n", podStatus)

}

func MakeInstallBackingServiceOperator(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Installing backing service operator source into the cluster...")
	res := util.GetOutput(util.Run("make", "install-backing-db-operator-source"))
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
	dbOprRes := util.GetOutput(util.Run("make", "install-backing-db-operator-subscription"))
	subRes := strings.TrimSpace(strings.Split(dbOprRes, "subscription.operators.coreos.com/db-operators")[1])
	t.Logf("subscription output is %s \n", subRes)

	// with openshift-operators namespace, capture the install plan
	ipName := util.GetCmdResult("", "oc", "get", "subscription", pkgManifest, "-n", operatorsNS, "-o", `jsonpath={.status.installplan.name}`)
	t.Logf("-> Pod name is %s \n", ipName)

	//// with openshift-operators namespace, capture the install plan status of db-operators
	t.Log(" Fetching the status of install plan ")
	ipStatus := util.GetCmdResult("Complete", "oc", "get", "ip", "-n", operatorsNS, ipName, "-o", `jsonpath={.status.phase}`)
	t.Logf("-> Pod name is %s \n", ipStatus)
	require.Equal(t, ipStatus, "Complete", "'install plan status' is %d \n", ipStatus)

	podName := util.GetPodNameFromListOfPods(operatorsNS, bckSvc)
	require.Containsf(t, podName, bckSvc, "list does not contain %s pod from the list of pods running backing service operator in the cluster", bckSvc)
	t.Logf("-> Pod name is %s \n", podName)

	t.Log("Fetching the status of running pod")
	podStatus := util.GetCmdResult("Running", "oc", "get", "pod", podName, "-n", operatorsNS, "-o", `jsonpath={.status.phase}`)
	require.Equal(t, "Running", podStatus, "pod status is %d \n", podStatus)

	t.Log("Checking the backing service's CRD is installed")
	crd := util.GetCmdResult("databases.postgresql.baiju.dev", "oc", "get", "crd", "databases.postgresql.baiju.dev")
	require.NotEmpty(t, crd, "packing service CRD not found")
}

func CreatePorject(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("Creating a project into the cluster...")
	res := util.GetOutput(util.Run("make", "create-project"))
	require.NotEmptyf(t, res, "", "Pjt not created because of this - %s \n", res)
	createPjtRes := util.GetPjtCreationRes(res, appNS)
	require.Containsf(t, createPjtRes, appNS, "Namespace - %s is not created because of %s", appNS, res)

	t.Logf("-> Project created - %s \n", createPjtRes)
	t.Log("Get the name of project added to the cluster...\n")
	getPjt := util.GetCmdResult("", "oc", "get", "project", appNS, "-o", `jsonpath={.metadata.name}`)

	t.Logf("-> Project created - %s \n", getPjt)
	require.Equal(t, getPjt, appNS, "Namespace created is %d \n", getPjt)

	t.Log("Get the status of the project added to the cluster...\n")
	pjtStatus := util.GetCmdResult("Active", "oc", "get", "project", appNS, "-o", `jsonpath={.status.phase}`)

	t.Logf("-> Project created %s has the status %s \n", appNS, pjtStatus)
	require.Equal(t, "Active", pjtStatus, "Namespace status is %d \n", pjtStatus)

	t.Log(" Setting the project to service-binding-demo ")
	pjtRes := util.GetCmdResult("", "oc", "project", appNS)
	t.Logf("Project/Namespace set is %s", pjtRes)
}

func ImportNodeJSApp(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("import an nodejs app to bind with backing service...")

	nodeJsAppArg := "nodejs~" + nodeJsApp
	t.Log(" Fetching the build config name of the app ")
	app := util.GetCmdResult("", "oc", "new-app", nodeJsAppArg, "--name", appName, "-n", appNS)
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
	buildPodName := util.GetPodNameFromListOfPods(appNS, expBuildPodName)
	require.Containsf(t, expBuildPodName, buildPodName, "list does not contain %s build pod from the list of pods running builds in the cluster", buildPodName)
	t.Logf("-> Pod name is %s \n", buildPodName)

	t.Log("Fetching the status of running build pod")
	buildPodStatus := util.GetCmdResult("Succeeded", "oc", "get", "pods", buildPodName, "-n", appNS, "-o", `jsonpath={.status.phase}`)
	require.Equal(t, "Succeeded", buildPodStatus, "Build pod status is %d \n", buildPodStatus)

	t.Log("Fetching the name of deployment config")
	dc := util.GetCmdResult("", "oc", "get", "dc", "-n", appNS, "-o", `jsonpath={.items[*].metadata.name}`)
	require.Equal(t, bc, dc, "DeploymentConfig name is %d \n", dc)
	t.Logf("-> Deployment Config name is %s \n", dc)

	UseDeployment(t, dc)

	t.Log(" Exposing an app ")
	svc := "svc/" + bc
	name := "--name=" + bc
	exposeRes := util.GetCmdResult("", "oc", "expose", svc, name)
	t.Logf("app exposed result is %s", exposeRes)

	t.Log(" Fetching the route of an app ")
	route := util.GetCmdResult("", "oc", "get", "route", bc, "-n", appNS, "-o", `jsonpath={.status.ingress[0].host}`)
	t.Logf("-> ROUTE - %s \n", route)

	host := "http://" + route
	checkNodeJSAppFrontend(t, host, "(DB: N/A)")
}

func UseDeployment(t *testing.T, dc string) {

	t.Log(" Delete the deployment config ")
	deletedStatus := util.GetCmdResult("", "oc", "delete", "dc", dc, "-n", appNS, "--wait=true")
	require.Containsf(t, deletedStatus, "deleted", "Deployment config is deleted with the message %d \n", deletedStatus)
	t.Logf("-> Deployment config is deleted with the message %s \n", deletedStatus)
	dcPod := dc + "-1-deploy"

	pods := util.GetPodLst(operatorsNS)
	checkFlag, podName := util.SrchItemFromLst(pods, dcPod)

	require.Equal(t, false, checkFlag, "list contains deployment config pod from the list of pods running in the cluster")
	require.Equal(t, "", podName, "list contains %s deployment config pod from the list of pods running in the cluster", podName)
	t.Logf("-> List does not contain deployment config pod running in the cluster")

	deploymentData := util.GetCmdResult("", "oc", "apply", "-f", "deployment.yaml")
	t.Logf("-> Deployment config is deleted with the message %s \n", deploymentData)

	deploymentPodName := util.GetPodNameFromListOfPods(appNS, appName)
	require.Containsf(t, deploymentPodName, appName, "list does not contain %s build pod from the list of pods running builds in the cluster", deploymentPodName)
	t.Logf("-> deployment build Pod name is %s \n", deploymentPodName)

	t.Log("Fetching the name of deployment")
	deployment := util.GetCmdResult("", "oc", "get", "deployment", "-n", appNS, "-o", `jsonpath={.items[*].metadata.name}`)
	require.Equal(t, appName, deployment, "Deployment name is %d \n", deployment)
	t.Logf("-> Deployment name is %s \n", deployment)

	t.Log("Fetching the status of deployment ")
	deploymentStatus := util.GetCmdResult("True", "oc", "get", "deployment", deployment, "-n", appNS, "-o", `jsonpath={.status.conditions[*].status}`)
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
	dbInstanceName := util.GetCmdResult("", "oc", "get", "db", "-n", appNS, "-o", `jsonpath={.items[*].metadata.name}`)
	require.Equal(t, dbName, dbInstanceName, "db instance name is %d \n", dbInstanceName)
	t.Logf("-> DB instance name is %s \n", dbInstanceName)

	connectionIP := util.GetCmdResult("", "oc", "get", "db", dbInstanceName, "-o", "jsonpath={.status.dbConnectionIP}")
	re := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
	require.Equalf(t, true, re.MatchString(connectionIP), "IP address of the pod which runs the database instance is %d", connectionIP)
	t.Logf("-> DB Operation Result - %s \n", connectionIP)

	dbPodName := util.GetPodNameFromListOfPods(appNS, dbName)
	require.Containsf(t, dbPodName, dbName, "list does not contain %s db pod from the list of pods running db instance in the cluster", dbPodName)
	t.Logf("-> Pod name is %s \n", dbPodName)

}

func CreateServiceBindingRequest(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Creating Backing DB instance for backing service to connect with the app...")
	resCreateSBR := util.GetOutput(util.Run("make", "create-service-binding-request"))
	resExp := strings.TrimSpace(strings.Split(resCreateSBR, "servicebindingrequest.apps.openshift.io/binding-request")[1])
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while creating service binding request", resExp)
	fmt.Printf("Result of creating service binding request is %s \n", resExp)

	t.Log("Creating service binding request to bind nodejs app to connect with the backing service which is database...")
	sbr := util.GetCmdResult("", "oc", "get", "sbr", "-n", appNS, "-o", `jsonpath={.items[*].metadata.name}`)
	require.NotEmptyf(t, sbr, "", "There are number of service binding request listed in a cluster %s \n", sbr)
	require.Contains(t, resCreateSBR, sbr, "Service binding request name is %d \n", sbr)
	t.Logf("-> Service binding request name is %s \n", sbr)

	t.Log("Fetching the status of binding")
	sbrStatus := util.GetCmdResult("True", "oc", "get", "sbr", sbr, "-n", appNS, "-o", `jsonpath={.status.conditions[*].status}`)
	require.Contains(t, sbrStatus, "True", "service binding status is %d \n", sbrStatus)
	t.Logf("-> service binding status is %s \n", sbrStatus)

	annotation := util.GetCmdResult("", "oc", "get", "sbr", sbr, "-n", appNS, "-o", `jsonpath='{.metadata.annotations.kubectl\.kubernetes\.io\/last-applied-configuration}'`)
	t.Logf("-> Annotation is %s \n", annotation)
	actSBRResponse := util.UnmarshalJSONData(annotation)
	expSBRResponse := util.GetSbrResponse()
	expResource := "deployments"
	expKind := "Database"

	require.Equal(t, expSBRResponse.Kind, actSBRResponse.Kind, "SBR kind is not matched, As expected kind is %d and actual kind is %d\n", expSBRResponse.Kind, actSBRResponse.Kind)
	require.Equal(t, appNS, actSBRResponse.Metadata.Namespace, "SBR Namespace is not matched, As expected namespace is %d and actual namespace is %d\n", appNS, actSBRResponse.Metadata.Namespace)
	require.Equal(t, sbr, actSBRResponse.Metadata.Name, "SBR Name is not matched, As expected name is %d and actual name is %d\n", sbr, actSBRResponse.Metadata.Name)
	require.Equal(t, expResource, actSBRResponse.Spec.ApplicationSelector.Resource, "SBR application resource is not matched, As expected application resource is %d and actual application resource is %d\n", expResource, actSBRResponse.Spec.ApplicationSelector.Resource)
	require.Equal(t, appName, actSBRResponse.Spec.ApplicationSelector.ResourceRef, "SBR application resource ref is not matched, As expected application resource ref is %d and actual application resource ref  is %d\n", appName, actSBRResponse.Spec.ApplicationSelector.ResourceRef)
	require.Equal(t, expKind, actSBRResponse.Spec.BackingServiceSelector.Kind, "SBR BackingServiceSelector kind is not matched, As expected BackingServiceSelector kind is %d and actual BackingServiceSelector kind  is %d\n", expKind, actSBRResponse.Spec.BackingServiceSelector.Kind)
	require.Equal(t, dbName, actSBRResponse.Spec.BackingServiceSelector.ResourceRef, "SBR BackingServiceSelector ResourceRef is not matched, As expected BackingServiceSelector ResourceRef is %d and actual BackingServiceSelector ResourceRef  is %d\n", dbName, actSBRResponse.Spec.BackingServiceSelector.ResourceRef)

	t.Log(" Fetching the route of an app ")
	route := util.GetCmdResult("", "oc", "get", "route", appName, "-n", appNS, "-o", `jsonpath={.status.ingress[0].host}`)
	t.Logf("-> ROUTE - %s \n", route)

	host := "http://" + route
	checkNodeJSAppFrontend(t, host, dbName)

	t.Log("Fetching the details of secret")
	secret := util.GetCmdResult("", "oc", "get", "sbr", sbr, "-n", appNS, "-o", `jsonpath={.status.secret}`)
	require.Contains(t, secret, sbr, "service binding secret detail -> %d \n", secret)
	t.Logf("-> service binding secret detail -> %s \n", secret)

	t.Log("Fetching the env details of binding")
	env := util.GetCmdResult("", "oc", "get", "deploy", appName, "-n", appNS, "-o", `jsonpath={.spec.template.spec.containers[0].env}`)
	require.Contains(t, env, "ServiceBindingOperator", "service binding env detail -> %d \n", env)
	t.Logf("-> service binding env detail ->%s \n", env)

	t.Log("Fetching the envFrom details of binding")
	envFrom := util.GetCmdResult("", "oc", "get", "deploy", appName, "-n", appNS, "-o", `jsonpath={.spec.template.spec.containers[0].envFrom}`)
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
