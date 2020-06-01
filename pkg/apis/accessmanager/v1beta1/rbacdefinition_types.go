package v1beta1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type BindingsSpec struct {
	RoleName string           `json:"roleName"`
	Kind     string           `json:"kind"`
	Subjects []rbacv1.Subject `json:"subjects"`
}

type NamespaceSpec struct {
	Name string `json:"name"`
}

type NamespacedSpec struct {
	Namespace         NamespaceSpec        `json:"namespace,omitempty"`
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector,omitempty"`
	Bindings          []BindingsSpec       `json:"bindings"`
}

type ClusterSpec struct {
	ClusterRoleName string           `json:"clusterRoleName"`
	Subjects        []rbacv1.Subject `json:"subjects"`
}

// RbacDefinitionSpec defines the desired state of RbacDefinition
type RbacDefinitionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Namespaced []NamespacedSpec `json:"namespaced,omitempty"`
	Cluster    []ClusterSpec    `json:"cluster,omitempty"`
}

// RbacDefinitionStatus defines the observed state of RbacDefinition
type RbacDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RbacDefinition is the Schema for the rbacdefinitions API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=rbacdefinitions,scope=Namespaced
type RbacDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RbacDefinitionSpec   `json:"spec,omitempty"`
	Status RbacDefinitionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RbacDefinitionList contains a list of RbacDefinition
type RbacDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RbacDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RbacDefinition{}, &RbacDefinitionList{})
}
