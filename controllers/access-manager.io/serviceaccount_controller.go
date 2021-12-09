package controllers

import (
	"access-manager/pkg/reconciler"
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ServiceAccountReconciler is a wrapper for needed runtime-objects
type ServiceAccountReconciler struct {
	Client client.Client
	Config *rest.Config
	Scheme *runtime.Scheme
	Logger logr.Logger
}

// Reconcile reads that state of the cluster for a ServiceAccount object and makes changes based on the state
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	_ = context.Background()
	_ = r.Logger.WithValues("serviceAccount", request.Name)

	// Fetch the ServiceAccount instance
	instance := &corev1.ServiceAccount{}
	err := r.Client.Get(ctx, request.NamespacedName, instance)
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

	rec := reconciler.Reconciler{Client: *kubernetes.NewForConfigOrDie(r.Config), ControllerClient: r.Client, Logger: r.Logger, Scheme: r.Scheme}
	return rec.ReconcileServiceAccount(instance)
}

func (r *ServiceAccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ServiceAccount{}).
		Complete(r)
}
