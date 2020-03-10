package service-binding-operator

import (
    "testing"
    "os"
    "github.com/redhat-developer/service-binding-operator/util"   
)

func TestMain(m *testing.M) {
    util.SetKubeConfig("config")	
    exitVal := m.Run()
    os.Exit(exitVal)
}