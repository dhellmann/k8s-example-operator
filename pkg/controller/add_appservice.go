package controller

import (
	"github.com/dhellmann/k8s-example-operator/pkg/controller/appservice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, appservice.Add)
}
