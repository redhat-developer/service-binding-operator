package main

import (
	bindingapi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	specv1alpha2 "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	scheme   = runtime.NewScheme()
)


func main() {
	binding := &bindingapi.ServiceBinding{}
	binding.Name = "sb1"

	specBinding := &specv1alpha2.ServiceBinding{}
	specBinding.Name = "sb2"

}