package rbacdefinition

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileDefinition is a wrapper for needed runtime-objects
type ReconcileDefinition interface {
	Client() client.Client
	Scheme() *runtime.Scheme
	Logger() logr.Logger
}

// ReconcileRbacDefinition is a wrapper for needed runtime-objects
type ReconcileRbacDefinition struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	K8sClient  client.Client
	UsedScheme *runtime.Scheme
	ReqLogger  logr.Logger
}

// Client returns an instance of client.Client
func (r ReconcileRbacDefinition) Client() client.Client {
	return r.K8sClient
}

// Scheme returns an instance of runtime.Scheme
func (r ReconcileRbacDefinition) Scheme() *runtime.Scheme {
	return r.UsedScheme
}

// Logger returns an instance of logr.Logger
func (r ReconcileRbacDefinition) Logger() logr.Logger {
	return r.ReqLogger
}
