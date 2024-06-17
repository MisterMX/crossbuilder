package markers

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	xapiext "github.com/crossplane/crossplane/apis/apiextensions/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// XRDMarkers lists all markers that directly modify the XRD (not validation
// schemas).
var XRDMarkers = []*definitionWithHelp{
	must(markers.MakeDefinition("crossbuilder:generate:xrd:claimNames", markers.DescribesType, ClaimNames{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:defaultCompositionRef", markers.DescribesType, DefaultCompositionRef{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:enforcedCompositionRef", markers.DescribesType, EnforcedCompositionRef{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:defaultCompositeDeletePolicy", markers.DescribesType, DefaultCompositeDeletePolicy{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:connectionSecretKeys", markers.DescribesType, ConnectionSecretKeys(nil))),
}

func init() {
	AllDefinitions = append(AllDefinitions, XRDMarkers...)
}

// +controllertools:marker:generateHelp:category=XRD

// ClaimNames is a marker to specify claim names for generated XRDs.
type ClaimNames struct {
	Kind       string   `marker:"kind"`
	Plural     string   `marker:"plural"`
	Singular   string   `marker:"singular,optional"`
	ShortNames []string `marker:"shortNames,optional"`
	ListKind   string   `marker:"listKind,optional"`
	Categories []string `marker:"categories,optional"`
}

// ApplyToXRD applies the claim names to the XRD.
func (c ClaimNames) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.ClaimNames = &apiext.CustomResourceDefinitionNames{
		Kind:       c.Kind,
		Plural:     c.Plural,
		Singular:   c.Singular,
		ShortNames: c.ShortNames,
		ListKind:   c.ListKind,
		Categories: c.Categories,
	}
	// test(c)
	return nil
}

// +controllertools:marker:generateHelp:category=XRD

// DefaultCompositionRef is a marker to specify the default composition ref of
// an XRD.
type DefaultCompositionRef struct {
	Name string `marker:"name"`
}

// ApplyToXRD applies the default composition ref to the XRD.
func (c DefaultCompositionRef) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.DefaultCompositionRef = &xapiext.CompositionReference{
		Name: c.Name,
	}
	// test(c)
	return nil
}

// +controllertools:marker:generateHelp:category=XRD

// EnforcedCompositionRef is a marker to specify the enforced composition ref of
// an XRD.
type EnforcedCompositionRef struct {
	Name string `marker:"name"`
}

// ApplyToXRD applies the enforced composition ref to the XRD.
func (c EnforcedCompositionRef) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.EnforcedCompositionRef = &xapiext.CompositionReference{
		Name: c.Name,
	}
	// test(c)
	return nil
}

// +controllertools:marker:generateHelp:category=XRD

// DefaultCompositeDeletePolicy is a marker to specify the default composite
// delete policy of an XRD.
type DefaultCompositeDeletePolicy struct {
	Policy xpv1.CompositeDeletePolicy `marker:"policy"`
}

// ApplyToXRD applies the enforced composition ref to the XRD.
func (c DefaultCompositeDeletePolicy) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.DefaultCompositeDeletePolicy = &c.Policy
	// test(c)
	return nil
}

// ConnectionSecretKeys is a marker to specify connection secret keys of an XRD
type ConnectionSecretKeys []string

func (c ConnectionSecretKeys) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.ConnectionSecretKeys = c
	return nil
}
