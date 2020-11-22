/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BindingsSpec struct {
	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	Name     string `json:"name"`
	RoleName string `json:"roleName"`
	Kind     string `json:"kind"`
	// +kubebuilder:validation:Optional
	Subjects []rbacv1.Subject `json:"subjects"`
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	AllServiceAccounts bool `json:"allServiceAccounts"`
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
	Name            string           `json:"name"`
	ClusterRoleName string           `json:"clusterRoleName"`
	Subjects        []rbacv1.Subject `json:"subjects"`
}

// RbacDefinitionSpec defines the desired state of RbacDefinition
type RbacDefinitionSpec struct {
	Paused     bool             `json:"paused,omitempty"`
	Namespaced []NamespacedSpec `json:"namespaced,omitempty"`
	Cluster    []ClusterSpec    `json:"cluster,omitempty"`
}

// RbacDefinitionStatus defines the observed state of RbacDefinition
type RbacDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +kubebuilder:object:root=true

// RbacDefinition is the Schema for the rbacdefinitions API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=rbacdefinitions,scope=Cluster
type RbacDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RbacDefinitionSpec   `json:"spec,omitempty"`
	Status RbacDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RbacDefinitionList contains a list of RbacDefinition
type RbacDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RbacDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RbacDefinition{}, &RbacDefinitionList{})
}
