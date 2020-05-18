package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"gotest.tools/v3/icmd"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	//CmdTimeout defines the amount of time we should spend waiting for the resource when condition is true
	CmdTimeout           = 10 * time.Minute
	defaultRetryInterval = 5 * time.Second
	defaultRetryTimeout  = 3 * time.Minute
)

var (
	workingDirPath, _ = os.Getwd()
	workingDirOp      = icmd.Dir(workingDirPath)
	kubeConfig        = os.Getenv("KUBECONFIG")
	environment       = []string{fmt.Sprintf("KUBECONFIG=%s", kubeConfig)}
	cntr              int
)

//Run function executes a command with timeout
func Run(cmd ...string) *icmd.Result {
	currentCmd := icmd.Cmd{
		Command: cmd,
		Timeout: CmdTimeout,
		Env:     environment,
	}
	fmt.Printf("=> Command to execute: %v \n", currentCmd)
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

//GetOutput returns the output using Stdout()
func GetOutput(res *icmd.Result) string {
	var output string
	exitCode := res.ExitCode
	if exitCode == 0 || exitCode == 127 {
		output = res.Stdout()
	} else {
		output = res.Stderr()
	}
	fmt.Printf("OUTPUT: %s \n", output)
	return output
}

//GetCreateDeleteProjectResult returns the result of make create-project
func GetCreateDeleteProjectResult(projectRes string, r *regexp.Regexp) (string, bool) {
	var item string
	lstArr := strings.Split(projectRes, "\n")
	for _, item := range lstArr {
		if r.MatchString(item) {
			return item, true
		}
	}
	return item, false
}

//GetRegExMatch returns the result of make create-project
func GetRegExMatch(srchData string, r *regexp.Regexp) (string, bool) {
	if r.MatchString(srchData) {
		return r.FindString(srchData), true
	}
	return r.FindString(srchData), false
}

//GetCmdResult executes execCmd function indefinitely till the default timeout occurs if there is no response of a command, returns the result(res) immediately
func GetCmdResult(status string, item ...string) (string, error) {
	result, err := GetCmdResultWithTimeout(status, defaultRetryInterval, defaultRetryTimeout, item...)
	return result, err
}

//GetCmdResultWithTimeout executes execCmd function indefinitely till the timeout occurs if there is no response of a command, returns the result(res) immediately
func GetCmdResultWithTimeout(status string, retryInterval time.Duration, retryTimeout time.Duration, item ...string) (string, error) {
	var res string
	cntr = 0
	checkFlag := false
	err := wait.PollImmediate(retryInterval, retryTimeout, func() (bool, error) {
		checkFlag, res = execCmd(item...)
		if checkFlag {
			if status != "" {
				if strings.Contains(res, status) {
					return true, nil
				}
				return false, nil
			}
			return true, nil
		}
		return false, nil
	})
	return res, err
}

//execCmd returns a boolean result and the result of the command (cmdRes) executed
func execCmd(item ...string) (bool, string) {
	cntr++
	var cmdRes string
	checkFlag := false
	fmt.Printf("CMD Result [%v]...iteration %v \n", item, cntr)
	cmdRes = GetOutput(Run(item...))
	if cmdRes != "" {
		checkFlag = true
	}
	return checkFlag, cmdRes
}

//GetPodLst fetches the list of pods
func GetPodLst(operatorsNS string) string {
	log.Print("Fetching the list of running pods")
	pods, err := GetCmdResult("", "oc", "get", "pods", "-n", operatorsNS, "-o", `jsonpath={.items[*].metadata.name}`)
	if pods == "" && err != nil {
		log.Fatalf("No pods are running...")
	}
	log.Printf("The list of running pods -- %v", pods)
	return pods
}

//GetPodNameFromListOfPods function returns required pod name from the list of pods running
func GetPodNameFromListOfPods(operatorsNS string, expPodName string) (string, error) {
	cntr = 0
	var podName string
	checkFlag := false
	pods := GetPodLst(operatorsNS)
	err := wait.PollImmediate(defaultRetryInterval, defaultRetryTimeout, func() (bool, error) {
		checkFlag, podName = SrchItemFromLst(pods, expPodName)
		if checkFlag {
			return true, nil
		}
		return false, nil
	})
	return podName, err
}

//SrchItemFromLst returns specific search item (srchItem) from the list (lst)
func SrchItemFromLst(lst, srchItem string) (bool, string) {
	var item string
	cntr++
	fmt.Printf("Get result of the command...iteration %v \n", cntr)
	lstArr := strings.Split(lst, " ")
	for _, item := range lstArr {
		if strings.Contains(item, srchItem) {
			if strings.Contains(srchItem, "-build") {
				fmt.Printf("item matched as %s \n", item)
				return true, item
			}
			return true, item
		}
	}
	return false, item
}

//UnmarshalJSONData unmarshall the data in form of json to a struct
func UnmarshalJSONData(jsonData string, obj *v1alpha1.ServiceBindingRequest) error {
	if strings.Contains(jsonData, "'") {
		jsonData = strings.Trim(jsonData, "'")
	}
	return json.Unmarshal([]byte(jsonData), obj)
}
