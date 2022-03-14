package v1

import (
	v1beta1 "github.com/ckotzbauer/access-manager/apis/access-manager.io/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this RbacDefinition to the Hub version (v1beta1).
func (src *RbacDefinition) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.RbacDefinition)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Cluster = mapFromV1ClusterSpecs(src.Spec.Cluster)
	dst.Spec.Namespaced = mapFromV1NamespacedSpecs(src.Spec.Namespaced)
	dst.Spec.Paused = src.Spec.Paused
	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *RbacDefinition) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.RbacDefinition)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Cluster = mapToV1ClusterSpecs(src.Spec.Cluster)
	dst.Spec.Namespaced = mapToV1NamespacedSpecs(src.Spec.Namespaced)
	dst.Spec.Paused = src.Spec.Paused
	return nil
}

func mapToV1ClusterSpecs(specs []v1beta1.ClusterSpec) []ClusterSpec {
	targets := make([]ClusterSpec, 0)

	for _, v := range specs {
		targets = append(targets, ClusterSpec{
			Name:            v.Name,
			ClusterRoleName: v.ClusterRoleName,
			Subjects:        v.Subjects,
		})
	}

	return targets
}

func mapFromV1ClusterSpecs(specs []ClusterSpec) []v1beta1.ClusterSpec {
	targets := make([]v1beta1.ClusterSpec, 0)

	for _, v := range specs {
		targets = append(targets, v1beta1.ClusterSpec{
			Name:            v.Name,
			ClusterRoleName: v.ClusterRoleName,
			Subjects:        v.Subjects,
		})
	}

	return targets
}

func mapToV1NamespacedSpecs(specs []v1beta1.NamespacedSpec) []NamespacedSpec {
	targets := make([]NamespacedSpec, 0)

	for _, v := range specs {
		targets = append(targets, NamespacedSpec{
			Namespace:         NamespaceSpec(v.Namespace),
			NamespaceSelector: v.NamespaceSelector,
			Bindings:          mapToV1BindingsSpecs(v.Bindings),
		})
	}

	return targets
}

func mapFromV1NamespacedSpecs(specs []NamespacedSpec) []v1beta1.NamespacedSpec {
	targets := make([]v1beta1.NamespacedSpec, 0)

	for _, v := range specs {
		targets = append(targets, v1beta1.NamespacedSpec{
			Namespace:         v1beta1.NamespaceSpec(v.Namespace),
			NamespaceSelector: v.NamespaceSelector,
			Bindings:          mapFromV1BindingsSpecs(v.Bindings),
		})
	}

	return targets
}

func mapToV1BindingsSpecs(specs []v1beta1.BindingsSpec) []BindingsSpec {
	targets := make([]BindingsSpec, 0)

	for _, v := range specs {
		targets = append(targets, BindingsSpec{
			Name:               v.Name,
			RoleName:           v.RoleName,
			Kind:               v.Kind,
			Subjects:           v.Subjects,
			AllServiceAccounts: v.AllServiceAccounts,
		})
	}

	return targets
}

func mapFromV1BindingsSpecs(specs []BindingsSpec) []v1beta1.BindingsSpec {
	targets := make([]v1beta1.BindingsSpec, 0)

	for _, v := range specs {
		targets = append(targets, v1beta1.BindingsSpec{
			Name:               v.Name,
			RoleName:           v.RoleName,
			Kind:               v.Kind,
			Subjects:           v.Subjects,
			AllServiceAccounts: v.AllServiceAccounts,
		})
	}

	return targets
}
