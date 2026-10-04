package markers

import (
	"testing"

	xapiextv2 "github.com/crossplane/crossplane/v2/apis/apiextensions/v2"
)

func TestV2MarkersRejectUnsupportedFeatures(t *testing.T) {
	tests := map[string]func() error{
		"composite delete policy": func() error {
			return (DefaultCompositeDeletePolicy{}).ApplyToXRDV2(&xapiextv2.CompositeResourceDefinition{}, "v1")
		},
		"connection secret keys": func() error {
			return ConnectionSecretKeys(nil).ApplyToXRDV2(&xapiextv2.CompositeResourceDefinition{}, "v1")
		},
	}

	for name, apply := range tests {
		t.Run(name, func(t *testing.T) {
			if err := apply(); err == nil {
				t.Fatal("expected unsupported marker error")
			}
		})
	}
}

func TestV2ClaimNamesUseLegacyClusterScope(t *testing.T) {
	xrd := &xapiextv2.CompositeResourceDefinition{}
	claimNames := ClaimNames{
		Kind:       "Widget",
		Plural:     "widgets",
		Singular:   "widget",
		ShortNames: []string{"wdg"},
		ListKind:   "WidgetList",
		Categories: []string{"example"},
	}

	if err := claimNames.ApplyToXRDV2(xrd, "v1"); err != nil {
		t.Fatalf("apply claim names: %v", err)
	}
	if xrd.Spec.Scope != xapiextv2.CompositeResourceScope("LegacyCluster") {
		t.Errorf("got scope %q, want LegacyCluster", xrd.Spec.Scope)
	}
	if got := xrd.Spec.ClaimNames; got == nil ||
		got.Kind != claimNames.Kind || got.Plural != claimNames.Plural ||
		got.Singular != claimNames.Singular || got.ListKind != claimNames.ListKind ||
		len(got.ShortNames) != 1 || got.ShortNames[0] != claimNames.ShortNames[0] ||
		len(got.Categories) != 1 || got.Categories[0] != claimNames.Categories[0] {
		t.Errorf("got claim names %#v, want %#v", got, claimNames)
	}
}

func TestV2CompositionReferenceMarkers(t *testing.T) {
	xrd := &xapiextv2.CompositeResourceDefinition{}
	if err := (DefaultCompositionRef{Name: "default"}).ApplyToXRDV2(xrd, "v1"); err != nil {
		t.Fatalf("apply default composition ref: %v", err)
	}
	if xrd.Spec.DefaultCompositionRef == nil || xrd.Spec.DefaultCompositionRef.Name != "default" {
		t.Errorf("got default composition ref %#v, want name default", xrd.Spec.DefaultCompositionRef)
	}

	if err := (EnforcedCompositionRef{Name: "enforced"}).ApplyToXRDV2(xrd, "v1"); err != nil {
		t.Fatalf("apply enforced composition ref: %v", err)
	}
	if xrd.Spec.EnforcedCompositionRef == nil || xrd.Spec.EnforcedCompositionRef.Name != "enforced" {
		t.Errorf("got enforced composition ref %#v, want name enforced", xrd.Spec.EnforcedCompositionRef)
	}
}
