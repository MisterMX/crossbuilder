package markers

import (
	"fmt"
	"slices"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xapiext "github.com/crossplane/crossplane/v2/apis/apiextensions/v2"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// XRDMarkers lists all markers that directly modify the XRD (not validation schemas).
var XRDMarkers = []*definitionWithHelp{
	must(markers.MakeDefinition("crossbuilder:generate:xrd:defaultCompositionRef", markers.DescribesType, DefaultCompositionRef{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:enforcedCompositionRef", markers.DescribesType, EnforcedCompositionRef{})),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:defaultCompositionUpdatePolicy", markers.DescribesType, DefaultCompositionUpdatePolicy(""))),
	must(markers.MakeDefinition("crossbuilder:generate:xrd:metadata", markers.DescribesType, Metadata{})),
}

func init() {
	AllDefinitions = append(AllDefinitions, XRDMarkers...)
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

// DefaultCompositionUpdatePolicy is the policy used when updating composites
// after a new Composition Revision has been created if no policy has been
// specified on the composite.
type DefaultCompositionUpdatePolicy string

func (p DefaultCompositionUpdatePolicy) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	policy := xpv1.UpdatePolicy(p)

	// Validate that the policy is a valid UpdatePolicy enum value
	policies := []xpv1.UpdatePolicy{
		xpv1.UpdateAutomatic,
		xpv1.UpdateManual,
	}

	if !slices.Contains(policies, policy) {
		return fmt.Errorf("invalid DefaultCompositionUpdatePolicy: %q", p)
	}

	xrd.Spec.DefaultCompositionUpdatePolicy = &policy
	return nil
}

// +controllertools:marker:generateHelp:category=XRD

type Metadata struct {
	Labels      *map[string]string `marker:"labels"`
	Annotations *map[string]string `marker:"annotations"`
}

func (m Metadata) ApplyToXRD(xrd *xapiext.CompositeResourceDefinition, version string) error {
	xrd.Spec.Metadata = &xapiext.CompositeResourceDefinitionSpecMetadata{
		Labels:      ptr.Deref(m.Labels, map[string]string{}),
		Annotations: ptr.Deref(m.Annotations, map[string]string{}),
	}
	return nil
}
