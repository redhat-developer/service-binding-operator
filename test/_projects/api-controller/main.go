package main

import (
	"context"
	bindingapi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	specv1alpha2 "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme   = runtime.NewScheme()
)

func init() {
	utilruntime.Must(bindingapi.AddToScheme(scheme))
	utilruntime.Must(specv1alpha2.AddToScheme(scheme))
}

func main() {
	binding := &bindingapi.ServiceBinding{}
	binding.Name = "sb1"

	specBinding := &specv1alpha2.ServiceBinding{}
	specBinding.Name = "sb2"

	m, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		panic(err)
	}
	m.Start(context.Background())
}