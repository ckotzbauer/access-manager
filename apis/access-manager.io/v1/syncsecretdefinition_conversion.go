package v1

import (
	v1beta1 "github.com/ckotzbauer/access-manager/apis/access-manager.io/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this SyncSecretDefinition to the Hub version (v1beta1).
func (src *SyncSecretDefinition) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.SyncSecretDefinition)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Source = mapFromV1SourceSpec(src.Spec.Source)
	dst.Spec.Targets = mapFromV1TargetSpecs(src.Spec.Targets)
	dst.Spec.Paused = src.Spec.Paused
	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *SyncSecretDefinition) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.SyncSecretDefinition)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Source = mapToV1SourceSpec(src.Spec.Source)
	dst.Spec.Targets = mapToV1TargetSpecs(src.Spec.Targets)
	dst.Spec.Paused = src.Spec.Paused
	return nil
}

func mapToV1SourceSpec(spec v1beta1.SourceSpec) SourceSpec {
	return SourceSpec{
		Namespace: spec.Namespace,
		Name:      spec.Name,
	}
}

func mapFromV1SourceSpec(spec SourceSpec) v1beta1.SourceSpec {
	return v1beta1.SourceSpec{
		Namespace: spec.Namespace,
		Name:      spec.Name,
	}
}

func mapToV1TargetSpecs(specs []v1beta1.TargetSpec) []TargetSpec {
	targets := make([]TargetSpec, 0)

	for _, v := range specs {
		targets = append(targets, TargetSpec{
			Namespace:         NamespaceSpec(v.Namespace),
			NamespaceSelector: v.NamespaceSelector,
		})
	}

	return targets
}

func mapFromV1TargetSpecs(specs []TargetSpec) []v1beta1.TargetSpec {
	targets := make([]v1beta1.TargetSpec, 0)

	for _, v := range specs {
		targets = append(targets, v1beta1.TargetSpec{
			Namespace:         v1beta1.NamespaceSpec(v.Namespace),
			NamespaceSelector: v.NamespaceSelector,
		})
	}

	return targets
}
