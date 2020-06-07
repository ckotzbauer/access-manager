package rbacdefinition

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileRbacDefinition is a wrapper for needed runtime-objects
type ReconcileRbacDefinition struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client           kubernetes.Clientset
	ControllerClient client.Client
	Scheme           *runtime.Scheme
	Logger           logr.Logger
}
