package util

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MapNamespaces(vs []corev1.Namespace) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = mapNamespaceName(v)
	}
	return vsm
}

func mapNamespaceName(ns corev1.Namespace) string {
	return ns.Name
}

func ContainsSubject(subjects []rbacv1.Subject, subject rbacv1.Subject) bool {
	for _, s := range subjects {
		if reflect.DeepEqual(s, subject) {
			return true
		}
	}

	return false
}

func namespacedName(meta metav1.ObjectMeta) string {
	return meta.Namespace + "/" + meta.Name
}

// IsRoleBindingEqual returns true if both objects are named equal in the same namespace and have the same RoleRef and Subjects
func IsRoleBindingEqual(rb1 rbacv1.RoleBinding, rb2 rbacv1.RoleBinding) bool {
	return namespacedName(rb1.ObjectMeta) == namespacedName(rb2.ObjectMeta) &&
		reflect.DeepEqual(rb1.RoleRef, rb2.RoleRef) &&
		reflect.DeepEqual(rb1.Subjects, rb2.Subjects)
}

// IsClusterRoleBindingEqual returns true if both objects are named equal and have the same RoleRef and Subjects
func IsClusterRoleBindingEqual(crb1 rbacv1.ClusterRoleBinding, crb2 rbacv1.ClusterRoleBinding) bool {
	return crb1.Name == crb2.Name &&
		reflect.DeepEqual(crb1.RoleRef, crb2.RoleRef) &&
		reflect.DeepEqual(crb1.Subjects, crb2.Subjects)
}

// RemoveAPIGroupFromRoleBinding ensures that the apiGroup is not set in RoleRef and Subjects
func RemoveAPIGroupFromRoleBinding(rb *rbacv1.RoleBinding) *rbacv1.RoleBinding {
	rb.RoleRef.APIGroup = ""

	for _, subject := range rb.Subjects {
		subject.APIGroup = ""
	}

	return rb
}

// RemoveAPIGroupFromClusterRoleBinding ensures that the apiGroup is not set in RoleRef and Subjects
func RemoveAPIGroupFromClusterRoleBinding(crb *rbacv1.ClusterRoleBinding) *rbacv1.ClusterRoleBinding {
	crb.RoleRef.APIGroup = ""

	for _, subject := range crb.Subjects {
		subject.APIGroup = ""
	}

	return crb
}
