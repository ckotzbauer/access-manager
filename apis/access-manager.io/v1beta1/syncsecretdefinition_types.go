package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SyncSecretDefinitionSpec defines the desired state of SyncSecretDefinition
type SyncSecretDefinitionSpec struct {
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	Paused  bool         `json:"paused,omitempty"`
	Source  SourceSpec   `json:"source"`
	Targets []TargetSpec `json:"targets"`
}

type TargetSpec struct {
	Namespace         NamespaceSpec        `json:"namespace,omitempty"`
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector,omitempty"`
}

type SourceSpec struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// SyncSecretDefinitionStatus defines the observed state of SyncSecretDefinition
type SyncSecretDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// SyncSecretDefinition is the Schema for the syncsecretdefinitions API
type SyncSecretDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SyncSecretDefinitionSpec   `json:"spec,omitempty"`
	Status SyncSecretDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SyncSecretDefinitionList contains a list of SyncSecretDefinition
type SyncSecretDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SyncSecretDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SyncSecretDefinition{}, &SyncSecretDefinitionList{})
}
