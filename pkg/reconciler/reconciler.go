package reconciler

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	accessmanagerv1beta1 "access-manager/apis/access-manager.io/v1beta1"
	"access-manager/pkg/util"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var rbacName = "RbacDefinition"

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

		if !r.IsServiceAccountRelevant(def, instance.Namespace) {
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
	roleBindingsToCreate := r.BuildAllRoleBindings(instance)
	clusterRoleBindingsToCreate := r.BuildAllClusterRoleBindings(instance)

	r.RemoveAllDeletableRoleBindings(instance.Name, roleBindingsToCreate)
	r.RemoveAllDeletableClusterRoleBindings(instance.Name, clusterRoleBindingsToCreate)

	for _, rb := range roleBindingsToCreate {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &rb, r.Scheme); err != nil {
			r.Logger.WithValues("RoleBinding", rb.Namespace+"/"+rb.Name).Error(err, "Failed to set controllerReference.")
			continue
		}

		if _, err := r.CreateOrRecreateRoleBinding(rb); err != nil {
			r.Logger.WithValues("RoleBinding", rb.Namespace+"/"+rb.Name).Error(err, "Failed to reconcile RoleBinding.")
		}
	}

	for _, crb := range clusterRoleBindingsToCreate {
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

// RemoveAllDeletableRoleBindings deletes all RoleBindings which wouldn't be created again.
func (r *Reconciler) RemoveAllDeletableRoleBindings(defName string, roleBindingsToCreate []rbacv1.RoleBinding) {
	roleBindingsToDelete, err := r.getRoleBindingsToDelete(defName, roleBindingsToCreate)

	if err != nil {
		r.Logger.Error(err, "Failed to fetch all deletable RoleBindings.")
	}

	for _, rb := range roleBindingsToDelete {
		r.Logger.Info("Deleting RoleBinding", "Name", fmt.Sprintf("%s/%s", rb.Namespace, rb.Name))
		err = r.Client.RbacV1().RoleBindings(rb.Namespace).Delete(context.TODO(), rb.Name, metav1.DeleteOptions{})

		if err != nil {
			r.Logger.WithValues("Name", fmt.Sprintf("%s/%s", rb.Namespace, rb.Name)).Error(err, "Failed to delete RoleBinding.")
		}
	}
}

// RemoveAllDeletableClusterRoleBindings deletes all ClusterRoleBindings which wouldn't be created again.
func (r *Reconciler) RemoveAllDeletableClusterRoleBindings(defName string, clusterRoleBindingsToCreate []rbacv1.ClusterRoleBinding) {
	clusterRoleBindingsToDelete, err := r.getClusterRoleBindingsToDelete(defName, clusterRoleBindingsToCreate)

	if err != nil {
		r.Logger.Error(err, "Failed to fetch all deletable ClusterRoleBindings.")
	}

	for _, crb := range clusterRoleBindingsToDelete {
		r.Logger.Info("Deleting ClusterRoleBinding", "Name", crb.Name)
		err = r.Client.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.Name, metav1.DeleteOptions{})

		if err != nil {
			r.Logger.WithValues("ClusterRoleBinding", crb.Name).Error(err, "Failed to delete ClusterRoleBinding.")
		}
	}
}

// CreateOrRecreateRoleBinding creates a new or recreates a existing RoleBinding
func (r *Reconciler) CreateOrRecreateRoleBinding(rb rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	existing, err := r.Client.RbacV1().RoleBindings(rb.Namespace).Get(context.TODO(), rb.Name, metav1.GetOptions{})
	if err == nil {
		if !HasNamedOwner(existing.OwnerReferences, rbacName, "") {
			r.Logger.Info("Existing RoleBinding is not owned by a RbacDefinition. Ignoring", "Name", fmt.Sprintf("%s/%s", existing.Namespace, existing.Name))
			return existing, nil
		}

		if util.IsRoleBindingEqual(*existing, rb) {
			return existing, nil
		}

		r.Logger.Info("Deleting RoleBinding", "Name", fmt.Sprintf("%s/%s", rb.Namespace, rb.Name))
		err = r.Client.RbacV1().RoleBindings(rb.Namespace).Delete(context.TODO(), rb.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	r.Logger.Info("Creating new RoleBinding", "Name", fmt.Sprintf("%s/%s", rb.Namespace, rb.Name))
	return r.Client.RbacV1().RoleBindings(rb.Namespace).Create(context.TODO(), &rb, metav1.CreateOptions{})
}

// CreateOrRecreateClusterRoleBinding creates a new or recreates a existing ClusterRoleBinding
func (r *Reconciler) CreateOrRecreateClusterRoleBinding(crb rbacv1.ClusterRoleBinding) (*rbacv1.ClusterRoleBinding, error) {
	existing, err := r.Client.RbacV1().ClusterRoleBindings().Get(context.TODO(), crb.Name, metav1.GetOptions{})
	if err == nil {
		if !HasNamedOwner(existing.OwnerReferences, rbacName, "") {
			r.Logger.Info("Existing ClusterRoleBinding is not owned by a RbacDefinition. Ignoring", "Name", existing.Name)
			return existing, nil
		}

		if util.IsClusterRoleBindingEqual(*existing, crb) {
			return existing, nil
		}

		r.Logger.Info("Deleting ClusterRoleBinding", "Name", crb.Name)
		err = r.Client.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	r.Logger.Info("Creating new ClusterRoleBinding", "Name", crb.Name)
	return r.Client.RbacV1().ClusterRoleBindings().Create(context.TODO(), &crb, metav1.CreateOptions{})
}

// BuildAllRoleBindings returns an array of RoleBindings for the given RbacDefinition
func (r *Reconciler) BuildAllRoleBindings(cr *accessmanagerv1beta1.RbacDefinition) []rbacv1.RoleBinding {
	var bindingObjects []rbacv1.RoleBinding = []rbacv1.RoleBinding{}

	for _, nsSpec := range cr.Spec.Namespaced {
		relevantNamespaces := r.GetRelevantNamespaces(nsSpec.NamespaceSelector, nsSpec.Namespace)
		if relevantNamespaces == nil {
			return nil
		}

		r.Logger.WithValues("Namespaces", util.MapNamespaces(relevantNamespaces)).Info("Found matching Namespaces.")

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

// DeleteOwnedRoleBindings deletes all RoleBindings in namespace owned by the RbacDefinition
func (r *Reconciler) DeleteOwnedRoleBindings(namespace string, def accessmanagerv1beta1.RbacDefinition) error {
	list, err := r.Client.RbacV1().RoleBindings(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return err
	}

	for _, rb := range list.Items {
		for _, ref := range rb.OwnerReferences {
			if HasNamedOwner([]metav1.OwnerReference{ref}, rbacName, def.Name) {
				r.Logger.Info("Deleting owned RoleBinding", "Name", fmt.Sprintf("%s/%s", rb.Namespace, rb.Name))
				err = r.Client.RbacV1().RoleBindings(namespace).Delete(context.TODO(), rb.Name, metav1.DeleteOptions{})

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
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

func (r *Reconciler) getRoleBindingsToDelete(defName string, creating []rbacv1.RoleBinding) ([]rbacv1.RoleBinding, error) {
	list, err := r.Client.RbacV1().RoleBindings("").List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	var bindings []rbacv1.RoleBinding = []rbacv1.RoleBinding{}

	for _, rb := range list.Items {
		if HasNamedOwner(rb.OwnerReferences, rbacName, defName) && !util.ContainsRoleBinding(creating, rb.Name, rb.Namespace) {
			bindings = append(bindings, rb)
		}
	}

	return bindings, nil
}

func (r *Reconciler) getClusterRoleBindingsToDelete(defName string, creating []rbacv1.ClusterRoleBinding) ([]rbacv1.ClusterRoleBinding, error) {
	list, err := r.Client.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	var bindings []rbacv1.ClusterRoleBinding = []rbacv1.ClusterRoleBinding{}

	for _, crb := range list.Items {
		if HasNamedOwner(crb.OwnerReferences, rbacName, defName) && !util.ContainsClusterRoleBinding(creating, crb.Name) {
			bindings = append(bindings, crb)
		}
	}

	return bindings, nil
}

// IsServiceAccountRelevant checks if the given definition includes all serviceaccounts
func (r *Reconciler) IsServiceAccountRelevant(spec accessmanagerv1beta1.RbacDefinition, ns string) bool {
	for _, nsSpec := range spec.Spec.Namespaced {
		namespaces := r.GetRelevantNamespaces(nsSpec.NamespaceSelector, nsSpec.Namespace)

		if util.ContainsNamespace(namespaces, ns) {
			for _, bindingSpec := range nsSpec.Bindings {
				if bindingSpec.AllServiceAccounts {
					return true
				}
			}
		}
	}

	return false
}
