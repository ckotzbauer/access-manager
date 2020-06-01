package rbacdefinition

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconcileDefinition interface {
	Client() client.Client
	Scheme() *runtime.Scheme
	Logger() logr.Logger
}

type ReconcileRbacDefinition struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	reqLogger logr.Logger
}

func (r ReconcileRbacDefinition) Client() client.Client {
	return r.client
}

func (r ReconcileRbacDefinition) Scheme() *runtime.Scheme {
	return r.scheme
}

func (r ReconcileRbacDefinition) Logger() logr.Logger {
	return r.reqLogger
}
