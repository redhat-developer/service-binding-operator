package controller

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebinding"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, servicebinding.Add)
}
