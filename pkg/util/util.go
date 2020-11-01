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

// ContainsSubject returns true, if the list of Subjects contains the given subject
func ContainsSubject(subjects []rbacv1.Subject, subject rbacv1.Subject) bool {
	for _, s := range subjects {
		if reflect.DeepEqual(s, subject) {
			return true
		}
	}

	return false
}

// ContainsRoleBinding returns true, if the list of RBs contains a Binding with the specified name and namespace
func ContainsRoleBinding(rbs []rbacv1.RoleBinding, name, ns string) bool {
	for _, rb := range rbs {
		if rb.Name == name && rb.Namespace == ns {
			return true
		}
	}

	return false
}

// ContainsClusterRoleBinding returns true, if the list of CRBs contains a Binding with the specified name
func ContainsClusterRoleBinding(crbs []rbacv1.ClusterRoleBinding, name string) bool {
	for _, crb := range crbs {
		if crb.Name == name {
			return true
		}
	}

	return false
}

// ContainsNamespace returns true, if the list of namespaces contains a object with the specified name
func ContainsNamespace(namespaces []corev1.Namespace, name string) bool {
	if namespaces == nil {
		return false
	}

	for _, ns := range namespaces {
		if ns.Name == name {
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
	rb1 = removeAPIGroupFromRoleBinding(rb1)
	name := namespacedName(rb1.ObjectMeta) == namespacedName(rb2.ObjectMeta)
	roleRef := reflect.DeepEqual(rb1.RoleRef, rb2.RoleRef)
	subjects := reflect.DeepEqual(rb1.Subjects, rb2.Subjects)
	return name && roleRef && subjects
}

// IsClusterRoleBindingEqual returns true if both objects are named equal and have the same RoleRef and Subjects
func IsClusterRoleBindingEqual(crb1 rbacv1.ClusterRoleBinding, crb2 rbacv1.ClusterRoleBinding) bool {
	crb1 = removeAPIGroupFromClusterRoleBinding(crb1)
	name := crb1.Name == crb2.Name
	roleRef := reflect.DeepEqual(crb1.RoleRef, crb2.RoleRef)
	subjects := reflect.DeepEqual(crb1.Subjects, crb2.Subjects)
	return name && roleRef && subjects
}

func removeAPIGroupFromRoleBinding(rb rbacv1.RoleBinding) rbacv1.RoleBinding {
	rb.RoleRef.APIGroup = ""
	var subjects []rbacv1.Subject = []rbacv1.Subject{}

	for _, subject := range rb.Subjects {
		subject.APIGroup = ""
		subjects = append(subjects, subject)
	}

	rb.Subjects = subjects
	return rb
}

func removeAPIGroupFromClusterRoleBinding(crb rbacv1.ClusterRoleBinding) rbacv1.ClusterRoleBinding {
	crb.RoleRef.APIGroup = ""
	var subjects []rbacv1.Subject = []rbacv1.Subject{}

	for _, subject := range crb.Subjects {
		subject.APIGroup = ""
		subjects = append(subjects, subject)
	}

	crb.Subjects = subjects
	return crb
}
