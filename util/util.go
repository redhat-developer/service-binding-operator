package util

//This will set the corresponding example directory 

import (
	"fmt"	
	"gotest.tools/v3/icmd"
	"time"
	"os"
	"path"
	"os/user"
	"log"
	)

// Defining a structure of string array to pass command line arguments
type FooCmd struct {
	Command  []string
}	

const (		
	//Timeout defines the amount of time we should spend waiting for the resource when condition is true
	Timeout = 10 * time.Minute	
)

func Run(cmd ...string) *icmd.Result {
	return icmd.RunCmd(icmd.Cmd{Command: cmd, Timeout: Timeout})
}

// MustSucceed asserts that the command ran with 0 exit code
func MustSucceed(args ...string) *icmd.Result {
	return Assert(args...)
}

// Assert runs a command and verifies exit code (0)
func Assert(args ...string) *icmd.Result {
	res := Run(args...)
	//t := &testsuitAdaptor{}
	//res.Assert(t, exp)
	return res
}

//Sets the directory
func SetDir(path string) {
	result := icmd.Dir(path)
	log.Println(fmt.Sprintf("Path set after cd %s",result))
	res := MustSucceed("pwd")
	log.Println(fmt.Sprintf("Path displayed on pwd %s",res))
} 

//Sets the KUBECONFIG to the cluster
func SetKubeConfig(configPath string){
	var defaultKubeconfig string
	if os.Getenv("KUBECONFIG") != "" {
		defaultKubeconfig = os.Getenv("KUBECONFIG")		
	} else if usr, err := user.Current(); err == nil {
		defaultKubeconfig = path.Join(usr.HomeDir, configPath)
	}
	log.Println(fmt.Sprintf("export KUBECONFIG=%s",defaultKubeconfig))
	MustSucceed("export", "KUBECONFIG=%s",defaultKubeconfig)
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