package controller

import (
	"context"

	"access-manager/pkg/reconciler"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var logNamespace = logf.Log.WithName("controller_namespace")

// AddNamespace creates a new Namespace Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func AddNamespace(mgr manager.Manager) error {
	return addNamespace(mgr, newNamespaceReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newNamespaceReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNamespace{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Config: mgr.GetConfig()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func addNamespace(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("namespace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Namespace
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// ReconcileNamespace is a wrapper for needed runtime-objects
type ReconcileNamespace struct {
	Client client.Client
	Config *rest.Config
	Scheme *runtime.Scheme
	Logger logr.Logger
}

// blank assignment to verify that ReconcileNamespace implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNamespace{}

// Reconcile reads that state of the cluster for a Namespace object and makes changes based on the state
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNamespace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.Logger = logNamespace.WithValues("Request.Name", request.Name)
	r.Logger.Info("Reconciling Namespace")

	// Fetch the Namespace instance
	instance := &corev1.Namespace{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
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
	return rec.ReconcileNamespace(instance)
}
