package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type XExampleParameters struct {
	Message string `json:"message"`
}

type XExampleSpec struct {
	Parameters XExampleParameters `json:"parameters"`
}

type XExampleStatus struct {
	Ready bool `json:"ready,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Namespaced,categories=crossplane
// +kubebuilder:subresource:status
// +crossbuilder:generate:xrd:defaultCompositionRef:name=example
type XExample struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   XExampleSpec   `json:"spec"`
	Status XExampleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type XExampleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []XExample `json:"items"`
}