package xrd

import (
	"encoding/json"
	"fmt"
	"go/ast"

	xapiext "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

const (
	errGenerateCRDs      = "failed to generate CRDs"
	errConvertCRDtoXRD   = "failed to convert CRD to XRD"
	errConvertJSONSchema = "failed to convert JSON schema"
	errWriteXRD          = "failed to write XRD to YAML"
)

type Generator struct {
	// IgnoreUnexportedFields indicates that we should skip unexported fields.
	//
	// Left unspecified, the default is false.
	IgnoreUnexportedFields *bool `marker:",optional"`

	// AllowDangerousTypes allows types which are usually omitted from CRD generation
	// because they are not recommended.
	//
	// Currently the following additional types are allowed when this is true:
	// float32
	// float64
	//
	// Left unspecified, the default is false
	AllowDangerousTypes *bool `marker:",optional"`

	// MaxDescLen specifies the maximum description length for fields in CRD's OpenAPI schema.
	//
	// 0 indicates drop the description for all fields completely.
	// n indicates limit the description to at most n characters and truncate the description to
	// closest sentence boundary if it exceeds n characters.
	MaxDescLen *int `marker:",optional"`

	// CRDVersions specifies the target API versions of the CRD type itself to
	// generate. Defaults to v1.
	//
	// Currently, the only supported value is v1.
	//
	// The first version listed will be assumed to be the "default" version and
	// will not get a version suffix in the output filename.
	//
	// You'll need to use "v1" to get support for features like defaulting,
	// along with an API server that supports it (Kubernetes 1.16+).
	CRDVersions []string `marker:"crdVersions,optional"`

	// GenerateEmbeddedObjectMeta specifies if any embedded ObjectMeta in the CRD should be generated
	GenerateEmbeddedObjectMeta *bool `marker:",optional"`
}

func (Generator) CheckFilter() loader.NodeFilter {
	return filterTypesForCRDs
}

// filterTypesForCRDs filters out all nodes that aren't used in CRD generation,
// like interfaces and struct fields without JSON tag.
func filterTypesForCRDs(node ast.Node) bool {
	switch node := node.(type) {
	case *ast.InterfaceType:
		// skip interfaces, we never care about references in them
		return false
	case *ast.StructType:
		return true
	case *ast.Field:
		_, hasTag := loader.ParseAstTag(node.Tag).Lookup("json")
		// fields without JSON tags mean we have custom serialization,
		// so only visit fields with tags.
		return hasTag
	default:
		return true
	}
}

func (Generator) RegisterMarkers(into *markers.Registry) error {
	return crdmarkers.Register(into)
}

func (g Generator) Generate(ctx *genall.GenerationContext) error {
	crdStorage := newCRDStorage()
	crdGenerator := crd.Generator{
		AllowDangerousTypes:        g.AllowDangerousTypes,
		MaxDescLen:                 g.MaxDescLen,
		CRDVersions:                g.CRDVersions,
		GenerateEmbeddedObjectMeta: g.GenerateEmbeddedObjectMeta,
		IgnoreUnexportedFields:     g.IgnoreUnexportedFields,
	}
	crdGeneratorCtx := &genall.GenerationContext{
		Collector:  ctx.Collector,
		Roots:      ctx.Roots,
		Checker:    ctx.Checker,
		OutputRule: crdStorage.OutputRule(),
		InputRule:  ctx.InputRule,
	}

	if err := crdGenerator.Generate(crdGeneratorCtx); err != nil {
		return errors.Wrap(err, errGenerateCRDs)
	}

	xrds := []*xapiext.CompositeResourceDefinition{}
	for _, crd := range crdStorage.CRDs {
		xrd, err := convertCRDToXRD(crd)
		if err != nil {
			return errors.Wrap(err, errConvertCRDtoXRD)
		}
		xrds = append(xrds, xrd)
	}

	for _, xrd := range xrds {
		fileName := fmt.Sprintf("%s_%s.yaml", xrd.Spec.Group, xrd.Spec.Names.Plural)
		if err := ctx.WriteYAML(fileName, []interface{}{xrd}, genall.WithTransform(transformRemoveCRDStatus)); err != nil {
			return errors.Wrap(err, errWriteXRD)
		}
	}
	return nil
}

func convertCRDToXRD(crd *apiext.CustomResourceDefinition) (*xapiext.CompositeResourceDefinition, error) {
	xrdVersions, err := buildXRDVersions(crd.Spec.Versions)
	if err != nil {
		return nil, err
	}

	xrd := &xapiext.CompositeResourceDefinition{
		ObjectMeta: crd.ObjectMeta,
		Spec: xapiext.CompositeResourceDefinitionSpec{
			Group: crd.Spec.Group,
			Names: crd.Spec.Names,
			//ClaimNames: ,
			Versions: xrdVersions,
			// DefaultCompositionRef: ,
			// EnforcedCompositionRef: ,
		},
	}
	return xrd, nil
}

func buildXRDVersions(crdVersions []apiext.CustomResourceDefinitionVersion) ([]xapiext.CompositeResourceDefinitionVersion, error) {
	xrdVersions := make([]xapiext.CompositeResourceDefinitionVersion, len(crdVersions))
	for i, cV := range crdVersions {
		schema, err := convertJSONSchemaToRawExtension(cV.Schema.OpenAPIV3Schema)
		if err != nil {
			return nil, errors.Wrap(err, errConvertJSONSchema)
		}

		xrdVersions[i] = xapiext.CompositeResourceDefinitionVersion{
			Name: cV.Name,
			// Referenceable: ,
			Served:             cV.Served,
			Deprecated:         &cV.Deprecated,
			DeprecationWarning: cV.DeprecationWarning,
			Schema: &xapiext.CompositeResourceValidation{
				OpenAPIV3Schema: schema,
			},
			AdditionalPrinterColumns: cV.AdditionalPrinterColumns,
		}
	}
	return xrdVersions, nil
}

func convertJSONSchemaToRawExtension(schema *apiext.JSONSchemaProps) (runtime.RawExtension, error) {
	removeCrossplaneInternalFieldFromSchema(*schema)
	rawExt := runtime.RawExtension{}
	raw, err := json.Marshal(schema)
	rawExt.Raw = raw
	return rawExt, err
}

func removeCrossplaneInternalFieldFromSchema(schema apiext.JSONSchemaProps) apiext.JSONSchemaProps {
	paths := [][]string{
		{"apiVersion"},
		{"kind"},
		{"metadata"},
		{"spec", "claimRef"},
		{"spec", "compositionRef"},
		{"spec", "compositionRevisionRef"},
		{"spec", "compositionSelector"},
		{"spec", "publishConnectionDetailsTo"},
		{"spec", "resourceRefs"},
		{"spec", "writeConnectionSecretToRef"},
		{"status", "conditions"},
		{"status", "connectionDetails"},
	}
	for _, path := range paths {
		schema = removePathFromSchema(schema, path)
	}
	return schema
}

func removePathFromSchema(schema apiext.JSONSchemaProps, pathSegments []string) apiext.JSONSchemaProps {
	propName := pathSegments[0]
	if len(pathSegments) == 1 {
		delete(schema.Properties, propName)
	} else if len(pathSegments) > 1 {
		if prop, ok := schema.Properties[propName]; ok {
			schema.Properties[propName] = removePathFromSchema(prop, pathSegments[1:])
		}
	}
	return schema
}

// transformRemoveCRDStatus ensures we do not write the CRD status field.
func transformRemoveCRDStatus(obj map[string]interface{}) error {
	delete(obj, "status")
	return nil
}
