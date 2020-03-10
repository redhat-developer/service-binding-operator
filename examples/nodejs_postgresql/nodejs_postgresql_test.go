package nodejs_postgresql

import (
	"strings"
	"testing"

	"github.com/redhat-developer/service-binding-operator/util"
)

var Path = "$GOPATH/src/github.com/redhat-developer/service-binding-operator/examples/nodejs_postgresql"
var ns = "openshift-operators"
var ipName, ipStatus, podName, podStatus string

func TestSetDir(t *testing.T) {
	util.SetDir(Path)
}

//Logs the oc status
func TestGetOCStatus(t *testing.T) {
	t.Log("--- Getting OC Status ---")
	ocStatus := util.MustSucceed("oc", "status").Stdout()
	t.Logf(" ---> OC Status is '%v' ", ocStatus)
	//ocStatus:=util.GetOCStatus()
	if strings.Contains(ocStatus, "svc/openshift - kubernetes.default.svc.cluster.local") != true {
		//t.Logf("Unable to connect to the cluster because of the following message \n %s", ocStatus)
		t.Fatalf("Unable to connect to the cluster because of the following message \n %s", ocStatus)
		//t.FailNow()
	}

	t.Logf(" *** Connected to cluster *** '%s'", ocStatus)
}

//Logs out the output of command make install-service-binding-operator
func TestMakeInstallServiceBindingOperator(t *testing.T) {
	//	log.Printf("--- CMD executed is make install-service-binding-operator --- %s \n", util.MustSucceed(
	//"make", "install-service-binding-operator").Stdout())

	t.Log("Installing serivice binding operator into the cluster...")
	res := util.MustSucceed("make", "install-service-binding-operator")
	//log.Printf("--- CMD executed is make install-service-binding-operator --- %s \n", res)
	t.Log("--- CMD executed is make install-service-binding-operator --- ")
	t.Logf("-> Status of install-service-binding-operator is %s", res)

	// with openshift-operators namespace, capture the install plan
	t.Log("Get install plan name from the cluster...")
	ipName := util.MustSucceed("oc", "get", "ip", "-n", ns)
	t.Log("--- CMD executed is  oc get ip -n openshift-operators --- ")
	t.Logf("Install plan-ip name is %s", ipName)

	//oc get ip -n openshift-operators <<Name>> -o jsonpath='{.status.phase}'
	t.Log(" Fetching the status of install plan ")
	ipStatus := util.MustSucceed("oc", "get", "ip", "-n", ns, "-o", "jsonpath='{.status.phase}'")
	t.Log("--- CMD executed is  oc get ip -n openshift-operators <<Name>> -o jsonpath='{.status.phase}' --- ")
	t.Logf("-> Status of Install plan-ip is %s", ipStatus)

	//oc get pods -n openshift-operator
	t.Log("Fetching the pod name of the running pod")
	podName := util.MustSucceed("oc", "get", "pods", "-n", ns)
	t.Log("--- CMD executed is  oc get pods -n openshift-operator --- ")
	t.Logf("-> Pod name is %s", podName)

	//oc get pod <<Name of pod(from step 4)>> -n openshift-operators -o jsonpath='{.status.phase}'
	t.Log("Fetching the status of running pod")
	podStatus := util.MustSucceed("oc", "get", "pods", "-n", ns)
	t.Log("--- CMD executed is  oc get pod $pods -n $ns -o jsonpath='{.status.phase}' --- ")
	t.Logf("Pod status is %s", podStatus)

}
