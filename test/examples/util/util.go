package util

//This will set the corresponding example directory

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/icmd"
)

/*/ Defining a structure of string array to pass command line arguments
type FooCmd struct {
	Command  []string
}*/

const (
	//Timeout defines the amount of time we should spend waiting for the resource when condition is true
	Timeout = 10 * time.Minute
)

var workingDirOp icmd.CmdOp

//Run runs a command with timeout
func Run(cmd ...string) *icmd.Result {
	return icmd.RunCmd(icmd.Cmd{Command: cmd, Timeout: Timeout}, workingDirOp)
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
func SetDir(t *testing.T, path string) {
	workingDirOp = icmd.Dir(path)
}

//SetKubeConfig sets the KUBECONFIG to the cluster
func SetKubeConfig(t *testing.T, configPath string) {
	var defaultKubeconfig string
	if os.Getenv("KUBECONFIG") != "" {
		defaultKubeconfig = os.Getenv("KUBECONFIG")
	} else if usr, err := user.Current(); err == nil {
		defaultKubeconfig = path.Join(usr.HomeDir, configPath)
	}
	t.Logf("export KUBECONFIG=%s", defaultKubeconfig)
	Run("export", "KUBECONFIG=%s", defaultKubeconfig)
}

//GetExamplesDir returns a root path to the examples
func GetExamplesDir(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err, "Error getting the working directory.")
	return path.Clean(fmt.Sprintf("%s/../../examples", wd))
}

/*func Check(e error) {
    if e != nil {
        panic(e)
    }
}*/

/*func WriteFile(data,path string){

	err := ioutil.WriteFile(path, data, 0644)

 check(err)
}*/
