package examples_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/redhat-developer/service-binding-operator/test/examples/util"
	"github.com/stretchr/testify/require"
)

var exampleName = "nodejs_postgresql"
var ns = "openshift-operators"
var ipName, ipStatus, podName, podStatus string

var clusterAvailable = false

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
	result := util.Run("oc", "status")
	ocStatus := result.ExitCode

	util.GetOutput(ocStatus, result, "OC Status")

	require.Equal(t, ocStatus, 1, "'oc status' is %d", ocStatus)
	clusterAvailable = true

	t.Log(" *** Connected to cluster *** ")
}

//Logs out the output of command make install-service-binding-operator
func TestMakeInstallServiceBindingOperator(t *testing.T) {
	//	log.Printf("--- CMD executed is make install-service-binding-operator --- %s \n", util.MustSucceed(
	//"make", "install-service-binding-operator").Stdout())

	checkClusterAvailable(t)

	t.Log("Installing serivice binding operator into the cluster...")
	res := util.Run("make", "install-service-binding-operator")
	//log.Printf("--- CMD executed is make install-service-binding-operator --- %s \n", res)
	t.Log("--- CMD executed is make install-service-binding-operator --- ")

	t.Logf("-> Status of install-service-binding-operator is %s", res.Stdout())

	// with openshift-operators namespace, capture the install plan
	t.Log("Get install plan name from the cluster...")
	ipName := util.Run("oc", "get", "ip", "-n", ns)
	t.Log("--- CMD executed is  oc get ip -n openshift-operators --- ")
	t.Logf("Install plan-ip name is %s", ipName)

	//oc get ip -n openshift-operators <<Name>> -o jsonpath='{.status.phase}'
	t.Log(" Fetching the status of install plan ")
	ipStatus := util.Run("oc", "get", "ip", "-n", ns, "-o", "jsonpath='{.status.phase}'")
	t.Log("--- CMD executed is  oc get ip -n openshift-operators <<Name>> -o jsonpath='{.status.phase}' --- ")
	t.Logf("-> Status of Install plan-ip is %s", ipStatus)

	//oc get pods -n openshift-operator
	t.Log("Fetching the pod name of the running pod")
	podName := util.Run("oc", "get", "pods", "-n", ns)
	t.Log("--- CMD executed is  oc get pods -n openshift-operator --- ")
	t.Logf("-> Pod name is %s", podName)

	//oc get pod <<Name of pod(from step 4)>> -n openshift-operators -o jsonpath='{.status.phase}'
	t.Log("Fetching the status of running pod")
	podStatus := util.Run("oc", "get", "pods", "-n", ns)
	t.Log("--- CMD executed is  oc get pod $pods -n $ns -o jsonpath='{.status.phase}' --- ")
	t.Logf("Pod status is %s", podStatus)

}

func checkClusterAvailable(t *testing.T) {
	if !clusterAvailable {
		t.Skip("Cluster is not available, skipping")
	}
}
