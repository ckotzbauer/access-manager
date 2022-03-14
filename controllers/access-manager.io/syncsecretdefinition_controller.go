package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/ckotzbauer/access-manager/apis/access-manager.io/v1"
	"github.com/ckotzbauer/access-manager/pkg/reconciler"
)

// SyncSecretDefinitionReconciler reconciles a SyncSecretDefinition object
type SyncSecretDefinitionReconciler struct {
	client.Client
	Config *rest.Config
	Logger logr.Logger
	Scheme *runtime.Scheme
}

func (r *SyncSecretDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Logger.WithValues("syncsecretdefinition", req.NamespacedName)

	// Fetch the SecretSyncDefinition instance
	instance := &v1.SyncSecretDefinition{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.Logger.Error(err, "Unexpected error occurred!")
		return reconcile.Result{}, err
	}

	if instance.Spec.Paused {
		return reconcile.Result{}, nil
	}

	r.Logger.Info("Reconciling SecretSyncDefinition", "Name", req.Name)
	rec := reconciler.Reconciler{Client: *kubernetes.NewForConfigOrDie(r.Config), Logger: r.Logger, Scheme: r.Scheme}
	return rec.ReconcileSyncSecretDefinition(instance)
}

func (r *SyncSecretDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.SyncSecretDefinition{}).
		Complete(r)
}
