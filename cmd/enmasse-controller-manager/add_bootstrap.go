package main

import (
	"github.com/enmasseproject/enmasse/pkg/controller/bootstrap"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, bootstrap.Add)
}
