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

func doReconcilation(instance *accessmanagerv1beta1.RbacDefinition, def ReconcileDefinition) (reconcile.Result, error) {
	// Define all (Cluster)RoleBindings objects
	roleBindings, err := BuildAllRoleBindings(instance, def)
	clusterRoleBindings := BuildAllClusterRoleBindings(instance)

	if err != nil {
		return reconcile.Result{}, err
	}

	for _, rb := range roleBindings {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &rb, def.Scheme()); err != nil {
			return reconcile.Result{}, err
		}

		if err = CreateOrRecreateRoleBinding(rb, def); err != nil {
			return reconcile.Result{}, err
		}
	}

	for _, crb := range clusterRoleBindings {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &crb, def.Scheme()); err != nil {
			return reconcile.Result{}, err
		}

		if err = CreateOrRecreateClusterRoleBinding(crb, def); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// CreateOrRecreateRoleBinding creates a new or recreates a existing RoleBinding
func CreateOrRecreateRoleBinding(rb rbacv1.RoleBinding, def ReconcileDefinition) error {
	found := &rbacv1.RoleBinding{}
	err := def.Client().Get(context.TODO(), types.NamespacedName{Name: rb.Name, Namespace: rb.Namespace}, found)
	if err == nil {
		def.Logger().Info("Deleting RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
		err = def.Client().Delete(context.TODO(), &rb)
		if err != nil {
			return err
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return err
	}

	def.Logger().Info("Creating new RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
	return def.Client().Create(context.TODO(), &rb)
}

// CreateOrRecreateClusterRoleBinding creates a new or recreates a existing ClusterRoleBinding
func CreateOrRecreateClusterRoleBinding(crb rbacv1.ClusterRoleBinding, def ReconcileDefinition) error {
	found := &rbacv1.ClusterRoleBinding{}
	err := def.Client().Get(context.TODO(), types.NamespacedName{Name: crb.Name}, found)
	if err == nil {
		def.Logger().Info("Deleting ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
		err = def.Client().Delete(context.TODO(), found)
		if err != nil {
			return err
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return err
	}

	def.Logger().Info("Creating new ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
	return def.Client().Create(context.TODO(), &crb)
}

// BuildAllRoleBindings returns an array of RoleBindings for the given RbacDefinition
func BuildAllRoleBindings(cr *accessmanagerv1beta1.RbacDefinition, def ReconcileDefinition) ([]rbacv1.RoleBinding, error) {
	var bindingObjects []rbacv1.RoleBinding = []rbacv1.RoleBinding{}

	for _, nsSpec := range cr.Spec.Namespaced {
		relevantNamespaces, err := GetRelevantNamespaces(nsSpec, def)
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

// BuildAllClusterRoleBindings returns an array of ClusterRoleBindings for the given RbacDefinition
func BuildAllClusterRoleBindings(cr *accessmanagerv1beta1.RbacDefinition) []rbacv1.ClusterRoleBinding {
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

// GetRelevantNamespaces returns a filtered list of namespaces matching the NamespacedSpec
func GetRelevantNamespaces(spec accessmanagerv1beta1.NamespacedSpec, def ReconcileDefinition) ([]corev1.Namespace, error) {
	namespaces := &corev1.NamespaceList{}
	options := &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string(spec.NamespaceSelector.MatchLabels)),
	}

	if err := def.Client().List(context.TODO(), namespaces, options); err != nil {
		return nil, err
	}

	return namespaces.Items, nil
}
