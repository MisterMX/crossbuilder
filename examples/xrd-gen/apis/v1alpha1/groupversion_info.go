package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	XRDGroup   = "test.example.com"
	XRDVersion = "v1alpha1"
)

var (
	// GroupVersion is the API Group Version used to register the objects
	GroupVersion = schema.GroupVersion{Group: XRDGroup, Version: XRDVersion}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
