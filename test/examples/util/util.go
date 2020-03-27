package util

//This will set the corresponding example directory

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"gotest.tools/v3/icmd"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	//Timeout defines the amount of time we should spend waiting for the resource when condition is true
	Timeout       = 10 * time.Minute
	retryInterval = 1 * time.Second
	retryTimeout  = 30 * time.Second
)

var (
	workingDirPath, _                = os.Getwd()
	workingDirOp                     = icmd.Dir(workingDirPath)
	kubeConfig                       = os.Getenv("KUBECONFIG")
	environment                      = []string{fmt.Sprintf("KUBECONFIG=%s", kubeConfig)}
	checkFlag                        = false
	cntr                             int
	ipName, dbOprRes, output, subRes string
)

//Run runs a command with timeout
func Run(cmd ...string) *icmd.Result {
	currentCmd := icmd.Cmd{
		Command: cmd,
		Timeout: Timeout,
		Env:     environment,
	}

	if workingDirOp != nil {
		return icmd.RunCmd(currentCmd, workingDirOp)
	}
	return icmd.RunCmd(currentCmd)
}

//SetDir sets the working directory
func SetDir(path string) {
	workingDirOp = icmd.Dir(path)
}

//SetKubeConfig sets the KUBECONFIG to the cluster
func SetKubeConfig(kc string) {
	kubeConfig = kc
}

//GetExamplesDir returns a root path to the examples
func GetExamplesDir() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get the working dir.")
	}
	return path.Clean(fmt.Sprintf("%s/../../examples", wd))
}

//GetPodNameFromLst returns specific name of the pod from the pod list
func GetPodNameFromLst(pods, oprName string) string {
	item := ""
	lstArr := strings.Split(pods, " ")
	for _, item := range lstArr {
		if strings.Contains(item, oprName) {
			fmt.Printf("item matched as %s \n", item)
			return item
			//break
		}
	}
	return item
}

//GetOutput returns the output using Stdout()
func GetOutput(res *icmd.Result, cmd string) string {

	exitCode := res.ExitCode
	if exitCode == 0 {
		output = res.Stdout()
	} else {
		output = res.Stderr()
	}
	fmt.Printf("Executed CMD: %s \n", cmd)
	fmt.Printf("OUTPUT: %s \n", output)

	return output

}

//WaitForIPNameAvailability returns boolean result if ip name is available, with openshift-operators namespace, capture the install plan
func WaitForIPNameAvailability(oprName string, ns string) string {
	cntr = 0
	wait.PollImmediate(retryInterval, retryTimeout, func() (bool, error) {
		checkFlag, ipName = checkIPNameAvailability(oprName, ns)
		if ipName != "" {
			return true, nil
		}
		return false, nil
	})

	return ipName
}

//WaitForDbOprAvailability returns boolean result if ip name is available, with openshift-operators namespace, capture the install plan
func WaitForDbOprAvailability(manifest string) string {
	cntr = 0
	wait.PollImmediate(retryInterval, retryTimeout, func() (bool, error) {
		checkFlag, dbOprRes = checkdbOprAvailability(manifest)
		if dbOprRes != "" {
			return true, nil
		}
		return false, nil
	})

	return dbOprRes
}

//checkIPNameAvailability returns boolean result if ip name is available, with openshift-operators namespace, capture the install plan
func checkIPNameAvailability(oprName string, ns string) (bool, string) {
	cntr++
	fmt.Printf("Get install plan name from the cluster...iteration %v \n", cntr)
	ipName := GetOutput(Run("oc", "get", "subscription", oprName, "-n", ns, "-o", `jsonpath={.status.installplan.name}`), "oc get subscription service-binding-operator -n openshift-operators -o jsonpath='{.status.installplan.name}'")
	if ipName != "" {
		checkFlag = true
	}
	return checkFlag, ipName
}

//checkIPNameAvailability returns boolean result if ip name is available, with openshift-operators namespace, capture the install plan
func checkdbOprAvailability(manifest string) (bool, string) {

	cntr++
	fmt.Printf("Get db operator subscription from the cluster...iteration %v \n", cntr)
	dbOprRes := GetOutput(Run("make", "install-backing-db-operator-subscription"), "make install-backing-db-operator-subscription")
	if dbOprRes != "" && strings.Contains(dbOprRes, manifest) {
		checkFlag = true		
	}
	return checkFlag, dbOprRes
}
