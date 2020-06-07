package rbacdefinition

import corev1 "k8s.io/api/core/v1"

func MapNamespaces(vs []corev1.Namespace, f func(corev1.Namespace) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func MapNamespaceName(ns corev1.Namespace) string {
	return ns.Name
}
