package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var BundleVersion string

func main() {
	type OwnedData struct {
		Kind        string `yaml:"kind"`
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		DisplayName string `yaml:"displayName"`
		Description string `yaml:"description"`
	}
	type CustomResourceDefinitions struct {
		Owned []OwnedData `yaml:"owned"`
	}

	crds2 := &CustomResourceDefinitions{
		Owned: []OwnedData{
			{
				Kind:        "ServiceBindingRequest",
				Name:        "servicebindingrequests.apps.openshift.io",
				Version:     "v1alpha1",
				DisplayName: "ServiceBindingRequest",
				Description: "Expresses intent to bind an operator-backed service with a Deployment",
			},
		},
	}
	curdir, err := os.Getwd()
	rootdir := filepath.Dir(filepath.Dir(curdir))
	csvBundle := filepath.Join(rootdir, "redhat-developer", "service-binding-operator", "deploy", "olm-catalog", "service-binding-operator", BundleVersion, "service-binding-operator.v"+BundleVersion+".clusterserviceversion.yaml")
	check(err)
	filenameCSVBundle, _ := filepath.Abs(csvBundle)
	yamlFileCSVBundle, err := ioutil.ReadFile(filenameCSVBundle)
	check(err)
	CSVMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal(yamlFileCSVBundle, &CSVMap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if spec, ok := CSVMap["spec"].(map[interface{}]interface{}); ok {
		if crds, ok := spec["customresourcedefinitions"].(map[interface{}]interface{}); ok {
			crds["owned"] = crds2.Owned
		}
	}
	d, err := yaml.Marshal(&CSVMap)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = ioutil.WriteFile(filenameCSVBundle, d, os.FileMode(777))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
