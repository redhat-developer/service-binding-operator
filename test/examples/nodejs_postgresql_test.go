package examples_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/redhat-developer/service-binding-operator/test/examples/util"
	"github.com/stretchr/testify/require"
)

var (
	exampleName = "nodejs_postgresql"
	ns          = "openshift-operators"
	oprName     = "service-binding-operator"
	expStatus   = "Complete"

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

//TestSetExampleDir tests that the corrent example directory was set as a working directory for running the commands.
func TestSetExampleDir(t *testing.T) {
	examplePath := fmt.Sprintf("%s/%s", util.GetExamplesDir(), exampleName)
	util.SetDir(examplePath)
	res := strings.TrimSpace(util.Run("pwd").Stdout())
	require.Equal(t, examplePath, res)
}

//Logs the oc status
func TestGetOCStatus(t *testing.T) {
	t.Log("--- Getting OC Status ---")
	util.GetOutput(util.Run("oc", "status"), "oc status")
	clusterAvailable = true
	t.Log(" *** Connected to cluster *** ")
}

//Logs out the output of command make install-service-binding-operator
func TestMakeInstallServiceBindingOperator(t *testing.T) {

	checkClusterAvailable(t)

	t.Log("Installing serivice binding operator into the cluster...")
	res := util.GetOutput(util.Run("make", "install-service-binding-operator"), "make install-service-binding-operator")
	resExp := strings.TrimSpace(strings.Split(res, "subscription.operators.coreos.com/service-binding-operator")[1])
	fmt.Printf("subscription output is %s \n", resExp)
	require.Containsf(t, []string{"created", "unchanged", "configured"}, resExp, "list does not contain %s while installing service binding operator", resExp)

	// with openshift-operators namespace, capture the install plan
	ipName = util.WaitForIPNameAvailability(oprName, ns)
	t.Logf("-> Install plan-ip name is %s \n", ipName)

	//// with openshift-operators namespace, capture the install plan status
	t.Log(" Fetching the status of install plan ")
	ipStatus := util.GetOutput(util.Run("oc", "get", "ip", "-n", ns, ipName, "-o", `jsonpath={.status.phase}`), "oc get ip -n openshift-operators <<Name>> -o jsonpath='{.status.phase}")
	require.Equal(t, ipStatus, expStatus, "'install plan status' is %d \n", ipStatus)

	//oc get pods -n openshift-operator
	t.Log("Fetching the pod name of the running pod")
	pods := util.GetOutput(util.Run("oc", "get", "pods", "-n", ns, "-o", "jsonpath={.items[*].metadata.name}"), "oc get pods -n openshift-operators -o jsonpath={.items[*]}.metadata.name")
	require.NotEmptyf(t, pods, "", "There are number of pods listed in a cluster %s \n", pods)

	podName := util.GetPodNameFromLst(pods, oprName)
	require.Containsf(t, podName, oprName, "list does not contain %s pod from the list of pods running service binding operator in the cluster", resExp)
	t.Logf("-> Pod name is %s \n", podName)

	//oc get pod <<Name of pod(from step 4)>> -n openshift-operators -o jsonpath='{.status.phase}'
	t.Log("Fetching the status of running pod")
	podStatus := util.GetOutput(util.Run("oc", "get", "pod", podName, "-n", ns, "-o", `jsonpath={.status.phase}`), "oc get pods $podName -n $ns -o jsonpath='{.status.phase}'")
	require.Equal(t, podStatus, "Running", "pod status is %d \n", podStatus)

}

func TestMakeInstallBackingServiceOperator(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Installing backing service operator into the cluster...")
	res := util.GetOutput(util.Run("make", "install-backing-db-operator-source"), "make install-backing-db-operator-source")
	resExp := strings.TrimSpace(strings.Split(res, "operatorsource.operators.coreos.com/db-operators")[1])
	fmt.Printf("db operator installation output is %s \n", resExp)
	require.Containsf(t, []string{"created", "unchanged"}, resExp, "list does not contain %s while installing db service operator", resExp)

	//with this command 'oc get packagemanifest | grep db-operators' make sure there is an entry to the package manifest
	t.Log("Get install plan name from the cluster...\n")
	manifest := util.GetOutput(util.Run("oc", "get", "packagemanifest", pkgManifest, "-o", `jsonpath={.metadata.name}`), "CMD: oc get packagemanifest db-operators -o jsonpath='{.metadata.name}'")
	t.Logf("-> %s has an entry in the package manifest \n", manifest)
	require.NotEmptyf(t, manifest, "", "There are number of manifest listed in a cluster %s \n", manifest)

	//Install the subscription using this command: make install-backing-db-operator-subscription
	t.Log("Installing backing service db subscription into the cluster...")

	//Get db-operatos name from the package manifest
	dbOprRes = util.WaitForDbOprAvailability(manifest)
	subRes := strings.TrimSpace(strings.Split(dbOprRes, "subscription.operators.coreos.com/db-operators")[1])
	t.Logf("subscription output is %s \n", subRes)
	require.Containsf(t, []string{"created", "unchanged"}, subRes, "list does not contain %s while installing backing service db operator", subRes)

	//pods := util.GetLstOfPods(ns)

	//oc get pods -n openshift-operator
	t.Log("Fetching the pod name of the running pod")
	pods := util.GetOutput(util.Run("oc", "get", "pods", "-n", ns, "-o", "jsonpath={.items[*].metadata.name}"), "oc get pods -n openshift-operators -o jsonpath={.items[*]}.metadata.name")
	require.NotEmptyf(t, pods, "", "There are number of pods listed in a cluster %s \n", pods)

	podName := util.GetPodNameFromLst(pods, bckSvc)
	require.Containsf(t, podName, bckSvc, "list does not contain %s pod from the list of pods running service binding operator in the cluster", bckSvc)
	t.Logf("-> Pod name is %s \n", podName)

	//oc get pod <<Name of pod(from step 4)>> -n openshift-operators -o jsonpath='{.status.phase}'
	t.Log("Fetching the status of running pod")
	podStatus := util.GetOutput(util.Run("oc", "get", "pod", podName, "-n", ns, "-o", `jsonpath={.status.phase}`), "oc get pods $podName -n $ns -o jsonpath='{.status.phase}'")
	require.Equal(t, podStatus, "Running", "pod status is %d \n", podStatus)

}

func checkClusterAvailable(t *testing.T) {
	if !clusterAvailable {
		t.Skip("Cluster is not available, skipping")
	}
}
