package util

//This will set the corresponding example directory

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"gotest.tools/v3/icmd"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	//CmdTimeout defines the amount of time we should spend waiting for the resource when condition is true
	CmdTimeout    = 10 * time.Minute
	retryInterval = 5 * time.Second
	retryTimeout  = 3 * time.Minute
)

var (
	workingDirPath, _                                             = os.Getwd()
	workingDirOp                                                  = icmd.Dir(workingDirPath)
	kubeConfig                                                    = os.Getenv("KUBECONFIG")
	environment                                                   = []string{fmt.Sprintf("KUBECONFIG=%s", kubeConfig)}
	checkFlag                                                     = false
	cntr                                                          int
	ipName, dbOprRes, output, subRes, pjtRes, ipStatus, podStatus string
)

//Run runs a command with timeout
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

//GetPodNameFromLst returns specific name of the pod from the pod list
func GetPodNameFromLst(pods, srchItem string) (bool, string) {
	item := ""
	checkFlag := false
	lstArr := strings.Split(pods, " ")
	for _, item := range lstArr {
		if strings.Contains(item, srchItem) {
			if strings.Contains(srchItem, "-build") {
				fmt.Printf("item matched as %s \n", item)
				checkFlag = true
				return checkFlag, item
			}
			return true, item
		}
	}
	return checkFlag, item
}

//GetOutput returns the output using Stdout()
func GetOutput(res *icmd.Result, cmd string) string {
	exitCode := res.ExitCode
	if exitCode == 0 || exitCode == 127 {
		output = res.Stdout()
	} else {
		output = res.Stderr()
	}
	//fmt.Printf("Executed CMD: %s \n", cmd)
	fmt.Printf("OUTPUT: %s \n", output)
	return output
}

//GetPjtCreationRes returns specific name of the pod from the pod list
func GetPjtCreationRes(pjtRes string, pjt string) string {
	item := ""
	lstArr := strings.Split(pjtRes, "\n")
	for _, item := range lstArr {
		if strings.Contains(item, pjt) {
			fmt.Printf("item matched as %s \n", item)
			return item
		}
	}
	return item
}

//GetCmdResult retrieves the info about build
func GetCmdResult(status string, item ...string) string {
	var res string
	cntr = 0
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
	if err != nil {
		return ""
	}
	return res
}

//execCmd returns boolean result if ip name is available, with openshift-operators namespace, capture the install plan
func execCmd(item ...string) (bool, string) {
	cntr++
	var cmdRes string
	checkFlag := false
	fmt.Printf("Get result of the command...iteration %v \n", cntr)
	//fmt.Printf("Command executed: %v \n", item)
	cmdRes = GetOutput(Run(item...), "")
	if cmdRes != "" {
		checkFlag = true
	}
	return checkFlag, cmdRes
}

//GetExecCmdResult retrieves the info about build
func GetExecCmdResult(status string, item ...string) string {
	var res string
	cntr = 0
	wait.PollImmediate(retryInterval, retryTimeout, func() (bool, error) {
		checkFlag, res = executeExecCmd(item...)
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
	return res
}

//executeExecCmd returns boolean result if ip name is available, with openshift-operators namespace, capture the install plan
func executeExecCmd(item ...string) (bool, string) {
	cntr++
	var cmdRes string
	checkFlag := false
	fmt.Printf("Get result of the command...iteration %v \n", cntr)
	//fmt.Printf("Command executed: %v \n", item)
	out, err := exec.Command("curlCMD", item...).Output()
	if err != nil {
		fmt.Println("Error!")
		log.Fatal(err)
	}

	if string(out) != "" {
		checkFlag = true
	}
	return checkFlag, cmdRes
}
