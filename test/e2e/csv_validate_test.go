package e2e

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"
)

// TestCsvRole tests whether deploy/roles and roles in the CSV matches.
func TestCsvRole(t *testing.T) {
	curdir, err := os.Getwd()
	require.NoError(t, err)
	rootdir := filepath.Dir(filepath.Dir(curdir))
	csv := filepath.Join(rootdir, "manifests-upstream", "0.0.20", "service-binding-operator.v0.0.20.clusterserviceversion.yaml")
	roles := filepath.Join(rootdir, "deploy", "role.yaml")

	var mapCSV interface{}
	var mapRoles interface{}
	filenameCSV, _ := filepath.Abs(csv)
	yamlFileCSV, err := ioutil.ReadFile(filenameCSV)
	check(err)
	map1 := make(map[interface{}]interface{})
	err = yaml.Unmarshal(yamlFileCSV, &map1)
	check(err)

	if spec, ok := map1["spec"].(map[interface{}]interface{}); ok {
		if install, ok := spec["install"].(map[interface{}]interface{}); ok {
			if spec, ok := install["spec"].(map[interface{}]interface{}); ok {
				if permissions, ok := spec["permissions"].([]interface{}); ok {
					if rules, ok := permissions[0].(map[interface{}]interface{}); ok {
						for i, j := range rules {
							if i == "rules" {
								mapCSV = j
							}
						}
					}
				}
			}
		}
	}

	filenameRoles, _ := filepath.Abs(roles)
	yamlFileRoles, err := ioutil.ReadFile(filenameRoles)
	check(err)
	map2 := make(map[interface{}]interface{})
	err = yaml.Unmarshal(yamlFileRoles, &map2)
	check(err)
	for i, j := range map2 {
		if i == "rules" {
			mapRoles = j
		}
	}

	result := reflect.DeepEqual(mapCSV, mapRoles)
	if result {
		t.Log("No Error- Validation Succeeded. The roles in deploy/roles.yaml and csv are equal.")
	} else {
		t.Error("Error- Validation failed. The roles in deploy/roles.yaml and csv are not equal.")
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
