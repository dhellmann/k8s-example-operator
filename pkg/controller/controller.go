package controller

import (
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

var log = logf.Log.WithName("controller/controller")

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	log.Info("AddToManager")
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}
