package reconciler

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	accessmanagerv1beta1 "access-manager/api/v1beta1"
	"access-manager/pkg/util"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	list := &accessmanagerv1beta1.RbacDefinitionList{}
	err := r.ControllerClient.List(context.TODO(), list)

	if err != nil {
		r.Logger.Error(err, "Unexpected error occurred!")
		return reconcile.Result{}, err
	}

	for _, def := range list.Items {
		if def.Spec.Paused {
			continue
		}

		if err = r.DeleteOwnedRoleBindings(instance.Name, def); err != nil {
			return reconcile.Result{}, err
		}

		_, err = r.ReconcileRbacDefinition(&def)

		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// ReconcileServiceAccount applies all desired changes of the ServiceAccount
func (r *Reconciler) ReconcileServiceAccount(instance *corev1.ServiceAccount) (reconcile.Result, error) {
	list := &accessmanagerv1beta1.RbacDefinitionList{}
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

// ReconcileRbacDefinition applies all desired changes of the RbacDefinition
func (r *Reconciler) ReconcileRbacDefinition(instance *accessmanagerv1beta1.RbacDefinition) (reconcile.Result, error) {
	// Define all (Cluster)RoleBindings objects
	roleBindings := r.BuildAllRoleBindings(instance)
	clusterRoleBindings := r.BuildAllClusterRoleBindings(instance)

	if roleBindings == nil {
		return reconcile.Result{}, nil
	}

	for _, rb := range roleBindings {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &rb, r.Scheme); err != nil {
			r.Logger.WithValues("RoleBinding", rb.Namespace+"/"+rb.Name).Error(err, "Failed to set controllerReference.")
			continue
		}

		if _, err := r.CreateOrRecreateRoleBinding(rb); err != nil {
			r.Logger.WithValues("RoleBinding", rb.Namespace+"/"+rb.Name).Error(err, "Failed to reconcile RoleBinding.")
		}
	}

	for _, crb := range clusterRoleBindings {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &crb, r.Scheme); err != nil {
			r.Logger.WithValues("ClusterRoleBinding", crb.Name).Error(err, "Failed to set controllerReference.")
			continue
		}

		if _, err := r.CreateOrRecreateClusterRoleBinding(crb); err != nil {
			r.Logger.WithValues("ClusterRoleBinding", crb.Name).Error(err, "Failed to reconcile ClusterRoleBinding.")
		}
	}

	return reconcile.Result{}, nil
}

// CreateOrRecreateRoleBinding creates a new or recreates a existing RoleBinding
func (r *Reconciler) CreateOrRecreateRoleBinding(rb rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	existing, err := r.Client.RbacV1().RoleBindings(rb.Namespace).Get(context.TODO(), rb.Name, metav1.GetOptions{})
	if err == nil {
		if !r.hasRbacDefinitionOwner(existing.OwnerReferences) {
			r.Logger.Info("Existing RoleBinding is not owned by a RbacDefinition. Ignoring", "RoleBinding.Name", existing.Name)
			return existing, nil
		}

		existing = util.RemoveAPIGroupFromRoleBinding(existing)
		if util.IsRoleBindingEqual(*existing, rb) {
			return existing, nil
		}

		r.Logger.Info("Deleting RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
		err = r.Client.RbacV1().RoleBindings(rb.Namespace).Delete(context.TODO(), rb.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	r.Logger.Info("Creating new RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
	return r.Client.RbacV1().RoleBindings(rb.Namespace).Create(context.TODO(), &rb, metav1.CreateOptions{})
}

// CreateOrRecreateClusterRoleBinding creates a new or recreates a existing ClusterRoleBinding
func (r *Reconciler) CreateOrRecreateClusterRoleBinding(crb rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	existing, err := r.Client.RbacV1().ClusterRoleBindings().Get(context.TODO(), crb.Name, metav1.GetOptions{})
	if err == nil {
		if !r.hasRbacDefinitionOwner(existing.OwnerReferences) {
			r.Logger.Info("Existing ClusterRoleBinding is not owned by a RbacDefinition. Ignoring", "ClusterRoleBinding.Name", existing.Name)
			return existing, nil
		}

		existing = util.RemoveAPIGroupFromClusterRoleBinding(existing)
		if util.IsClusterRoleBindingEqual(*existing, crb) {
			return existing, nil
		}

		r.Logger.Info("Deleting ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
		err = r.Client.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	r.Logger.Info("Creating new ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
	return r.Client.RbacV1().ClusterRoleBindings().Create(context.TODO(), &crb, metav1.CreateOptions{})
}

// BuildAllRoleBindings returns an array of RoleBindings for the given RbacDefinition
func (r *Reconciler) BuildAllRoleBindings(cr *accessmanagerv1beta1.RbacDefinition) []rbacv1.RoleBinding {
	var bindingObjects []rbacv1.RoleBinding = []rbacv1.RoleBinding{}

	for _, nsSpec := range cr.Spec.Namespaced {
		relevantNamespaces := r.GetRelevantNamespaces(nsSpec)
		if relevantNamespaces == nil {
			return nil
		}

		for _, ns := range relevantNamespaces {
			for _, bindingSpec := range nsSpec.Bindings {
				name := bindingSpec.Name

				if name == "" {
					name = bindingSpec.RoleName
				}

				subjects := bindingSpec.Subjects

				if bindingSpec.AllServiceAccounts {
					subjects = r.appendServiceAccountSubjects(r.getServiceAccounts(ns.Name), subjects)
				}

				if len(subjects) == 0 {
					continue
				}

				roleBinding := rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: ns.Name,
					},
					RoleRef: rbacv1.RoleRef{
						Name: bindingSpec.RoleName,
						Kind: bindingSpec.Kind,
					},
					Subjects: subjects,
				}

				bindingObjects = append(bindingObjects, roleBinding)
			}
		}
	}

	return bindingObjects
}

// BuildAllClusterRoleBindings returns an array of ClusterRoleBindings for the given RbacDefinition
func (r *Reconciler) BuildAllClusterRoleBindings(cr *accessmanagerv1beta1.RbacDefinition) []rbacv1.ClusterRoleBinding {
	var bindingObjects []rbacv1.ClusterRoleBinding = []rbacv1.ClusterRoleBinding{}

	for _, bindingSpec := range cr.Spec.Cluster {
		name := bindingSpec.Name

		if name == "" {
			name = bindingSpec.ClusterRoleName
		}

		if len(bindingSpec.Subjects) == 0 {
			continue
		}

		clusterRoleBinding := rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			RoleRef: rbacv1.RoleRef{
				Name: bindingSpec.ClusterRoleName,
				Kind: "ClusterRole",
			},
			Subjects: bindingSpec.Subjects,
		}

		bindingObjects = append(bindingObjects, clusterRoleBinding)
	}

	return bindingObjects
}

// GetRelevantNamespaces returns a filtered list of namespaces matching the NamespacedSpec
func (r *Reconciler) GetRelevantNamespaces(spec accessmanagerv1beta1.NamespacedSpec) []corev1.Namespace {
	if spec.NamespaceSelector.MatchLabels != nil || len(spec.NamespaceSelector.MatchExpressions) > 0 {
		selector, err := metav1.LabelSelectorAsSelector(&spec.NamespaceSelector)
		if err != nil {
			r.Logger.WithValues("Selector", spec.NamespaceSelector).Error(err, "Could not parse LabelSelector or MatchExpression.")
			return nil
		}

		listOptions := metav1.ListOptions{LabelSelector: selector.String()}
		namespaces, err := r.Client.CoreV1().Namespaces().List(context.TODO(), listOptions)
		if err != nil {
			r.Logger.Error(err, "Could not list namespaces.")
			return nil
		}

		r.Logger.WithValues("Namespaces", util.MapNamespaces(namespaces.Items)).Info("Found matching Namespaces.")
		return namespaces.Items

	} else if spec.Namespace.Name != "" {
		namespace, err := r.Client.CoreV1().Namespaces().Get(context.TODO(), spec.Namespace.Name, metav1.GetOptions{})
		if err != nil {
			r.Logger.WithValues("NsName", spec.Namespace.Name).Error(err, "Could not find Namespace with name.")
			return nil
		}

		r.Logger.WithValues("NsName", namespace.Name).Info("Found namespaces with name.")
		return []corev1.Namespace{*namespace}
	} else {
		r.Logger.Error(nil, "Invalid role binding, namespace or namespaceSelector required")
		return nil
	}
}

// DeleteOwnedRoleBindings deletes all RoleBindings in namespace owned by the RbacDefinition
func (r *Reconciler) DeleteOwnedRoleBindings(namespace string, def accessmanagerv1beta1.RbacDefinition) error {
	list, err := r.Client.RbacV1().RoleBindings(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return err
	}

	for _, rb := range list.Items {
		for _, ref := range rb.OwnerReferences {
			if *ref.Controller && ref.Kind == "RbacDefinition" && ref.Name == def.Name {
				r.Logger.Info("Deleting owned RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
				err = r.Client.RbacV1().RoleBindings(namespace).Delete(context.TODO(), rb.Name, metav1.DeleteOptions{})

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *Reconciler) hasRbacDefinitionOwner(refs []metav1.OwnerReference) bool {
	for _, ref := range refs {
		if ref.Kind == "RbacDefinition" {
			return true
		}
	}

	return false
}

func (r *Reconciler) getServiceAccounts(ns string) []corev1.ServiceAccount {
	accountList, err := r.Client.CoreV1().ServiceAccounts(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		r.Logger.WithValues("NsName", ns).Error(err, "Could not list ServiceAccounts in namespace.")
		return nil
	}

	return accountList.Items
}

func (r *Reconciler) appendServiceAccountSubjects(accounts []corev1.ServiceAccount, subjects []rbacv1.Subject) []rbacv1.Subject {
	for _, account := range accounts {
		subject := rbacv1.Subject{
			Kind: "ServiceAccount",
			Name: account.Name,
		}

		if !util.ContainsSubject(subjects, subject) {
			subjects = append(subjects, subject)
		}
	}

	return subjects
}
