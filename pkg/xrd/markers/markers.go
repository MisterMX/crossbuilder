package markers

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	xapiext "github.com/crossplane/crossplane/apis/apiextensions/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// CRDMarkers lists all markers that directly modify the CRD (not validation
// schemas).
var XRDMarkers = []*definitionWithHelp{
	must(markers.MakeDefinition("crossbuilder:generate:xrd:claimNames", markers.DescribesType, ClaimNames{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:defaultCompositionRef", markers.DescribesType, DefaultCompositionRef{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:enforcedCompositionRef", markers.DescribesType, EnforcedCompositionRef{})),
}

func init() {
	AllDefinitions = append(AllDefinitions, XRDMarkers...)
}

// +controllertools:marker:generateHelp:category=XRD

type ClaimNames struct {
	Kind       string   `marker:"kind"`
	Plural     string   `marker:"plural"`
	Singular   string   `marker:"singular"`
	ShortNames []string `marker:"shortNames"`
	ListKind   string   `marker:"listKind"`
	Categories []string `marker:"categories"`
}

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

type DefaultCompositionRef struct {
	Name string `marker:"name"`
}

func (c DefaultCompositionRef) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.DefaultCompositionRef = &xpv1.Reference{
		Name: c.Name,
	}
	// test(c)
	return nil
}

// +controllertools:marker:generateHelp:category=XRD

type EnforcedCompositionRef struct {
	Name string `marker:"name"`
}

func (c EnforcedCompositionRef) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.EnforcedCompositionRef = &xpv1.Reference{
		Name: c.Name,
	}
	// test(c)
	return nil
}
