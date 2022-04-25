package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type XExampleParameters struct {
	ExampleField string `json:"exampleField"`
}

type XExampleSpec struct {
	Parameters XExampleParameters `json:"parameters"`
}

type XExampleStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +genclient
// +genclient:nonNamespaced

// +kubebuilder:resource:scope=Cluster,categories=crossplane
// +kubebuilder:subresource:status
// +crossbuilder:generate:xrd:claimNames:kind=Example,plural=examples
// +crossbuilder:generate:xrd:defaultCompositionRef:name=example-composition
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
