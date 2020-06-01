package rbacdefinition

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	accessmanagerv1beta1 "access-manager/pkg/apis/accessmanager/v1beta1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *ReconcileRbacDefinition) reconcile(instance *accessmanagerv1beta1.RbacDefinition) (reconcile.Result, error) {
	// Define all (Cluster)RoleBindings objects
	roleBindings, err := r.buildAllRoleBindings(instance)
	clusterRoleBindings := buildAllClusterRoleBindings(instance)

	if err != nil {
		return reconcile.Result{}, err
	}

	for _, rb := range roleBindings {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &rb, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		if err = r.createOrUpdateRoleBinding(rb); err != nil {
			return reconcile.Result{}, err
		}
	}

	for _, crb := range clusterRoleBindings {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &crb, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		if err = r.createOrUpdateClusterRoleBinding(crb); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileRbacDefinition) createOrUpdateRoleBinding(rb rbacv1.RoleBinding) error {
	found := &rbacv1.RoleBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: rb.Name, Namespace: rb.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		r.reqLogger.Info("Creating new RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
		err = r.client.Create(context.TODO(), &rb)
		if err != nil {
			return err
		}

		// RoleBinding created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	r.reqLogger.Info("Updating existing RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
	return r.client.Update(context.TODO(), &rb)
}

func (r *ReconcileRbacDefinition) createOrUpdateClusterRoleBinding(crb rbacv1.ClusterRoleBinding) error {
	found := &rbacv1.ClusterRoleBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crb.Name}, found)
	if err != nil && errors.IsNotFound(err) {
		r.reqLogger.Info("Creating new ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
		err = r.client.Create(context.TODO(), &crb)
		if err != nil {
			return err
		}

		// ClusterRoleBinding created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	r.reqLogger.Info("Updating existing ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
	return r.client.Update(context.TODO(), &crb)
}

// buildAllRoleBindings returns a busybox pod with the same name/namespace as the cr
func (r *ReconcileRbacDefinition) buildAllRoleBindings(cr *accessmanagerv1beta1.RbacDefinition) ([]rbacv1.RoleBinding, error) {
	var bindingObjects []rbacv1.RoleBinding = []rbacv1.RoleBinding{}

	for _, nsSpec := range cr.Spec.Namespaced {
		relevantNamespaces, err := r.getRelevantNamespaces(nsSpec)
		if err != nil {
			return nil, err
		}

		for _, ns := range relevantNamespaces {
			for _, bindingSpec := range nsSpec.Bindings {
				roleBinding := rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      bindingSpec.RoleName,
						Namespace: ns.Name,
					},
					RoleRef: rbacv1.RoleRef{
						Name: bindingSpec.RoleName,
						Kind: bindingSpec.Kind,
					},
					Subjects: bindingSpec.Subjects,
				}

				bindingObjects = append(bindingObjects, roleBinding)
			}
		}
	}

	return bindingObjects, nil
}

// buildAllRoleBindings returns a busybox pod with the same name/namespace as the cr
func buildAllClusterRoleBindings(cr *accessmanagerv1beta1.RbacDefinition) []rbacv1.ClusterRoleBinding {
	var bindingObjects []rbacv1.ClusterRoleBinding = []rbacv1.ClusterRoleBinding{}

	for _, bindingSpec := range cr.Spec.Cluster {
		clusterRoleBinding := rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: bindingSpec.ClusterRoleName,
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

func (r *ReconcileRbacDefinition) getRelevantNamespaces(spec accessmanagerv1beta1.NamespacedSpec) ([]corev1.Namespace, error) {
	namespaces := &corev1.NamespaceList{}
	options := &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string(spec.NamespaceSelector.MatchLabels)),
	}

	if err := r.client.List(context.TODO(), namespaces, options); err != nil {
		return nil, err
	}

	return namespaces.Items, nil
}
