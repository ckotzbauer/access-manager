package util

import corev1 "k8s.io/api/core/v1"

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
