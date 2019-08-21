package controller

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/catchall"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, catchall.Add)
}
