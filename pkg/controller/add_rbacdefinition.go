package controller

import (
	"access-manager/pkg/controller/rbacdefinition"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, rbacdefinition.Add)
}
