package examples_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/redhat-developer/service-binding-operator/test/examples/util"
	"github.com/stretchr/testify/require"
)

var (
	exampleName = "nodejs_postgresql"
	ns          = "openshift-operators"
	oprName     = "service-binding-operator"
	expStatus   = "Complete"

	ipName, ipStatus, podName, podStatus string
	clusterAvailable                     = false

	pkgManifest = "db-operators"
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
	res := util.GetOutput(util.Run("make", "install-service-binding-operator"), "CMD: make install-service-binding-operator")

	resExp := strings.TrimSpace(strings.Split(res, "subscription.operators.coreos.com/service-binding-operator")[1])
	fmt.Printf("subscription output is %s \n", resExp)

	require.Containsf(t, []string{"created", "unchanged"}, resExp, "list does not contain %s while installing service binding operator", resExp)

	// with openshift-operators namespace, capture the install plan
	t.Log("Get install plan name from the cluster...\n")
	ipName := util.GetOutput(util.Run("oc", "get", "subscription", oprName, "-n", ns, "-o", `jsonpath={.status.installplan.name}`), "CMD: oc get subscription service-binding-operator -n openshift-operators -o jsonpath='{.status.installplan.name}'")
	t.Logf("-> Install plan-ip name is %s \n", ipName)

	//// with openshift-operators namespace, capture the install plan status
	t.Log(" Fetching the status of install plan ")
	ipStatus := util.GetOutput(util.Run("oc", "get", "ip", "-n", ns, ipName, "-o", `jsonpath={.status.phase}`), "CMD: oc get ip -n openshift-operators <<Name>> -o jsonpath='{.status.phase}")
	require.Equal(t, ipStatus, expStatus, "'install plan status' is %d \n", ipStatus)

	//oc get pods -n openshift-operator
	t.Log("Fetching the pod name of the running pod")
	pods := util.GetOutput(util.Run("oc", "get", "pods", "-n", ns, "-o", "jsonpath={.items[*].metadata.name}"), "CMD: oc get pods -n openshift-operators -o jsonpath={.items[*]}.metadata.name")

	podName := util.GetPodNameFromLst(pods, oprName)
	require.Containsf(t, podName, oprName, "list does not contain %s pod from the list of pods running service binding operator in the cluster", resExp)
	t.Logf("-> Pod name is %s \n", podName)

	//oc get pod <<Name of pod(from step 4)>> -n openshift-operators -o jsonpath='{.status.phase}'
	t.Log("Fetching the status of running pod")
	podStatus := util.GetOutput(util.Run("oc", "get", "pod", podName, "-n", ns, "-o", `jsonpath={.status.phase}`), "CMD: oc get pods $podName -n $ns -o jsonpath='{.status.phase}")
	require.Equal(t, podStatus, "Running", "pod status is %d \n", podStatus)

}

func TestMakeInstallBackingServiceOperator(t *testing.T) {
	checkClusterAvailable(t)

	t.Log("Installing backing service operator into the cluster...")
	res := util.GetOutput(util.Run("make", "install-backing-db-operator-source"), "CMD: make install-backing-db-operator-source")

	resExp := strings.TrimSpace(strings.Split(res, "operatorsource.operators.coreos.com/db-operators")[1])
	fmt.Printf("db operator installation output is %s \n", resExp)

	require.Containsf(t, []string{"created", "unchanged"}, resExp, "list does not contain %s while installing db service operator", resExp)

}

func checkClusterAvailable(t *testing.T) {
	if !clusterAvailable {
		t.Skip("Cluster is not available, skipping")
	}
}
