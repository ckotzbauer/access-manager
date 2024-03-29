package controllers

import (
	"context"

	"github.com/ckotzbauer/access-manager/pkg/reconciler"

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

// SecretReconciler is a wrapper for needed runtime-objects
type SecretReconciler struct {
	Client client.Client
	Config *rest.Config
	Scheme *runtime.Scheme
	Logger logr.Logger
}

// Reconcile reads that state of the cluster for a Secret object and makes changes based on the state
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *SecretReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	_ = context.Background()
	_ = r.Logger.WithValues("secret", request.Name)

	// Fetch the Secret instance
	instance := &corev1.Secret{}
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
	return rec.ReconcileSecret(instance)
}

func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}
