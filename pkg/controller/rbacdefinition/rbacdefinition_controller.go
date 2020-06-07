package rbacdefinition

import (
	"context"

	accessmanagerv1beta1 "access-manager/pkg/apis/accessmanager/v1beta1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_rbacdefinition")

// Add creates a new RbacDefinition Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRbacDefinition{Client: *kubernetes.NewForConfigOrDie(mgr.GetConfig()), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rbacdefinition-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RbacDefinition
	err = c.Watch(&source.Kind{Type: &accessmanagerv1beta1.RbacDefinition{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner RbacDefinition
	watchTypes := []runtime.Object{
		&corev1.Namespace{},
		&rbacv1.ClusterRoleBinding{},
		&rbacv1.RoleBinding{},
	}

	for i := range watchTypes {
		err := c.Watch(&source.Kind{Type: watchTypes[i]}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &accessmanagerv1beta1.RbacDefinition{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileRbacDefinition implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRbacDefinition{}

// Reconcile reads that state of the cluster for a RbacDefinition object and makes changes based on the state read
// and what is in the RbacDefinition.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRbacDefinition) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.Logger = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.Logger.Info("Reconciling RbacDefinition")

	// Fetch the RbacDefinition instance
	instance := &accessmanagerv1beta1.RbacDefinition{}
	err := r.Client.RESTClient().Get().Resource("rbacdefinitions").Do(context.TODO()).Into(instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	return doReconcilation(instance, *r)
}
