package reconciler

import (
	"context"

	v1 "github.com/ckotzbauer/access-manager/apis/access-manager.io/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler runtime-object
type Reconciler struct {
	Client           kubernetes.Clientset
	ControllerClient client.Client
	Scheme           *runtime.Scheme
	Logger           logr.Logger
}

// ReconcileNamespace applies all desired changes of the Namespace
func (r *Reconciler) ReconcileNamespace(instance *corev1.Namespace) (reconcile.Result, error) {
	result1, err1 := r.processRbacDefinitions()
	result2, err2 := r.processSecretDefinitions()

	if err1 != nil {
		return result1, err1
	}

	if err2 != nil {
		return result2, err2
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) processRbacDefinitions() (reconcile.Result, error) {
	list := &v1.RbacDefinitionList{}
	err := r.ControllerClient.List(context.TODO(), list)

	if err != nil {
		r.Logger.Error(err, "Unexpected error occurred!")
		return reconcile.Result{}, err
	}

	for _, def := range list.Items {
		if def.Spec.Paused {
			continue
		}

		_, err = r.ReconcileRbacDefinition(&def)

		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) processSecretDefinitions() (reconcile.Result, error) {
	list := &v1.SyncSecretDefinitionList{}
	err := r.ControllerClient.List(context.TODO(), list)

	if err != nil {
		r.Logger.Error(err, "Unexpected error occurred!")
		return reconcile.Result{}, err
	}

	for _, def := range list.Items {
		if def.Spec.Paused {
			continue
		}

		_, err = r.ReconcileSyncSecretDefinition(&def)

		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// HasNamedOwner returns true if the owner array includes a object of the givien kind and name
func HasNamedOwner(refs []metav1.OwnerReference, kind, name string) bool {
	for _, ref := range refs {
		if ref.Controller != nil && *ref.Controller && ref.Kind == kind && (name == "" || name == ref.Name) {
			return true
		}
	}

	return false
}

// GetRelevantNamespaces returns a filtered list of namespaces matching the NamespacedSpec
func (r *Reconciler) GetRelevantNamespaces(selector metav1.LabelSelector, nameSpec v1.NamespaceSpec) []corev1.Namespace {
	if selector.MatchLabels != nil || len(selector.MatchExpressions) > 0 {
		selector, err := metav1.LabelSelectorAsSelector(&selector)
		if err != nil {
			r.Logger.WithValues("Selector", selector).Error(err, "Could not parse LabelSelector or MatchExpression.")
			return nil
		}

		listOptions := metav1.ListOptions{LabelSelector: selector.String()}
		namespaces, err := r.Client.CoreV1().Namespaces().List(context.TODO(), listOptions)
		if err != nil {
			r.Logger.Error(err, "Could not list namespaces.")
			return nil
		}

		return namespaces.Items

	} else if nameSpec.Name != "" {
		namespace, err := r.Client.CoreV1().Namespaces().Get(context.TODO(), nameSpec.Name, metav1.GetOptions{})
		if err != nil {
			r.Logger.WithValues("NsName", nameSpec.Name).Error(err, "Could not find Namespace with name.")
			return nil
		}

		return []corev1.Namespace{*namespace}
	} else {
		r.Logger.Error(nil, "Invalid role binding, namespace or namespaceSelector required")
		return nil
	}
}
