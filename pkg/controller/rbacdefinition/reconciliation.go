package rbacdefinition

import (
	"context"
	err "errors"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	accessmanagerv1beta1 "access-manager/pkg/apis/accessmanager/v1beta1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func doReconcilation(instance *accessmanagerv1beta1.RbacDefinition, def *ReconcileRbacDefinition) (reconcile.Result, error) {
	// Define all (Cluster)RoleBindings objects
	roleBindings, err := BuildAllRoleBindings(instance, def)
	clusterRoleBindings := BuildAllClusterRoleBindings(instance)

	if err != nil {
		return reconcile.Result{}, err
	}

	for _, rb := range roleBindings {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &rb, def.Scheme); err != nil {
			def.Logger.WithValues("RoleBinding", rb.Namespace+"/"+rb.Name).Error(err, "Failed to set controllerReference.")
			return reconcile.Result{}, err
		}

		if _, err = CreateOrRecreateRoleBinding(rb, def); err != nil {
			def.Logger.WithValues("RoleBinding", rb.Namespace+"/"+rb.Name).Error(err, "Failed to reconcile RoleBinding.")
			return reconcile.Result{}, err
		}
	}

	for _, crb := range clusterRoleBindings {
		// Set RbacDefinition instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, &crb, def.Scheme); err != nil {
			def.Logger.WithValues("ClusterRoleBinding", crb.Name).Error(err, "Failed to set controllerReference.")
			return reconcile.Result{}, err
		}

		if _, err = CreateOrRecreateClusterRoleBinding(crb, def); err != nil {
			def.Logger.WithValues("ClusterRoleBinding", crb.Name).Error(err, "Failed to reconcile ClusterRoleBinding.")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// CreateOrRecreateRoleBinding creates a new or recreates a existing RoleBinding
func CreateOrRecreateRoleBinding(rb rbacv1.RoleBinding, def *ReconcileRbacDefinition) (*rbacv1.RoleBinding, error) {
	_, err := def.Client.RbacV1().RoleBindings(rb.Namespace).Get(context.TODO(), rb.Name, metav1.GetOptions{})
	if err == nil {
		def.Logger.Info("Deleting RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
		err = def.Client.RbacV1().RoleBindings(rb.Namespace).Delete(context.TODO(), rb.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	def.Logger.Info("Creating new RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
	return def.Client.RbacV1().RoleBindings(rb.Namespace).Create(context.TODO(), &rb, metav1.CreateOptions{})
}

// CreateOrRecreateClusterRoleBinding creates a new or recreates a existing ClusterRoleBinding
func CreateOrRecreateClusterRoleBinding(crb rbacv1.ClusterRoleBinding, def *ReconcileRbacDefinition) (*rbacv1.ClusterRoleBinding, error) {
	_, err := def.Client.RbacV1().ClusterRoleBindings().Get(context.TODO(), crb.Name, metav1.GetOptions{})
	if err == nil {
		def.Logger.Info("Deleting ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
		err = def.Client.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, err
		}
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	def.Logger.Info("Creating new ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
	return def.Client.RbacV1().ClusterRoleBindings().Create(context.TODO(), &crb, metav1.CreateOptions{})
}

// BuildAllRoleBindings returns an array of RoleBindings for the given RbacDefinition
func BuildAllRoleBindings(cr *accessmanagerv1beta1.RbacDefinition, def *ReconcileRbacDefinition) ([]rbacv1.RoleBinding, error) {
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
func GetRelevantNamespaces(spec accessmanagerv1beta1.NamespacedSpec, def *ReconcileRbacDefinition) ([]corev1.Namespace, error) {
	if spec.NamespaceSelector.MatchLabels != nil || len(spec.NamespaceSelector.MatchExpressions) > 0 {
		selector, err := metav1.LabelSelectorAsSelector(&spec.NamespaceSelector)
		if err != nil {
			def.Logger.Error(err, "Could not parse LabelSelector or MatchExpression.")
			return nil, err
		}

		listOptions := metav1.ListOptions{LabelSelector: selector.String()}
		namespaces, err := def.Client.CoreV1().Namespaces().List(context.TODO(), listOptions)
		if err != nil {
			def.Logger.Error(err, "Could not list namespaces.")
			return nil, err
		}

		def.Logger.WithValues("Namespaces", MapNamespaces(namespaces.Items, MapNamespaceName)).Info("Found matching Namespaces.")
		return namespaces.Items, nil

	} else if spec.Namespace.Name != "" {
		namespace, err := def.Client.CoreV1().Namespaces().Get(context.TODO(), spec.Namespace.Name, metav1.GetOptions{})
		if err != nil {
			def.Logger.WithValues("NsName", spec.Namespace.Name).Error(err, "Could not found Namespace with name.")
			return nil, err
		}

		def.Logger.Info("Found namespaces with name name", "name", namespace.Name)
		return []corev1.Namespace{*namespace}, nil
	} else {
		def.Logger.Error(nil, "Invalid role binding, namespace or namespace selector required")
		return nil, err.New("Invalid role binding, namespace or namespace selector required")
	}
}
