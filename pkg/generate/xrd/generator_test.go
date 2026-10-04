package xrd

import (
	"testing"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGeneratorAPIVersionDefaultsToV2(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		want       string
		wantErr    bool
	}{
		{name: "default", want: APIVersionV2},
		{name: "explicit v1", apiVersion: APIVersionV1, want: APIVersionV1},
		{name: "explicit v2", apiVersion: APIVersionV2, want: APIVersionV2},
		{name: "unsupported", apiVersion: "v3", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (Generator{APIVersion: tt.apiVersion}).apiVersion()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected unsupported API version error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve API version: %v", err)
			}
			if got != tt.want {
				t.Errorf("got API version %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertCRDToXRDV2(t *testing.T) {
	crd := &apiext.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "widgets.example.org",
		},
		Spec: apiext.CustomResourceDefinitionSpec{
			Group: "example.org",
			Names: apiext.CustomResourceDefinitionNames{
				Kind:   "Widget",
				Plural: "widgets",
			},
			Scope: apiext.ClusterScoped,
			Versions: []apiext.CustomResourceDefinitionVersion{{
				Name:   "v1alpha1",
				Served: true,
				Schema: &apiext.CustomResourceValidation{OpenAPIV3Schema: &apiext.JSONSchemaProps{Type: "object"}},
			}},
		},
	}

	got, err := convertCRDToXRDV2(crd)
	if err != nil {
		t.Fatalf("convert CRD: %v", err)
	}
	if got.APIVersion != "apiextensions.crossplane.io/v2" {
		t.Errorf("got apiVersion %q, want apiextensions.crossplane.io/v2", got.APIVersion)
	}
	if got.Name != crd.Name {
		t.Errorf("got name %q, want %q", got.Name, crd.Name)
	}
	if got.Spec.Scope != "Cluster" {
		t.Errorf("got scope %q, want Cluster", got.Spec.Scope)
	}
	if len(got.Spec.Versions) != 1 {
		t.Fatalf("got %d versions, want 1", len(got.Spec.Versions))
	}
	version := got.Spec.Versions[0]
	if !version.Served || !version.Referenceable {
		t.Errorf("got served=%t referenceable=%t, want both true", version.Served, version.Referenceable)
	}
	if version.Schema == nil || len(version.Schema.OpenAPIV3Schema.Raw) == 0 {
		t.Error("expected generated OpenAPI schema")
	}
}

func TestConvertCRDToXRDV2DefaultsScope(t *testing.T) {
	crd := &apiext.CustomResourceDefinition{
		Spec: apiext.CustomResourceDefinitionSpec{
			Names: apiext.CustomResourceDefinitionNames{Kind: "Widget", Plural: "widgets"},
			Versions: []apiext.CustomResourceDefinitionVersion{{
				Name:   "v1",
				Served: true,
				Schema: &apiext.CustomResourceValidation{OpenAPIV3Schema: &apiext.JSONSchemaProps{Type: "object"}},
			}},
		},
	}

	got, err := convertCRDToXRDV2(crd)
	if err != nil {
		t.Fatalf("convert CRD: %v", err)
	}
	if got.Spec.Scope != "Namespaced" {
		t.Errorf("got scope %q, want Namespaced", got.Spec.Scope)
	}
}
