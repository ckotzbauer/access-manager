package util

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
