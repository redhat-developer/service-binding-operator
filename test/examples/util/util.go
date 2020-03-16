package util

//This will set the corresponding example directory

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"gotest.tools/v3/icmd"
)

const (
	//Timeout defines the amount of time we should spend waiting for the resource when condition is true
	Timeout = 10 * time.Minute
)

var (
	workingDirPath, _ = os.Getwd()
	workingDirOp      = icmd.Dir(workingDirPath)
	kubeConfig        = os.Getenv("KUBECONFIG")
	environment       = []string{fmt.Sprintf("KUBECONFIG=%s", kubeConfig)}
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

/*/ MustSucceed asserts that the command ran with 0 exit code
func MustSucceed(args ...string) *icmd.Result {
	return Assert(args...)
}

// Assert runs a command and verifies exit code (0)
func Assert(args ...string) *icmd.Result {
	res := Run(args...)
	//t := &testsuitAdaptor{}
	//res.Assert(t, exp)
	return res
}*/

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
