package build

import (
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	xapiextv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/mistermx/crossbuilder/pkg/generate/utils"
)

const (
	errEmptyCompositionname                 = "composition name must not be empty"
	errFmtBuildComposedTemplate             = "cannot build composed template at index %d"
	errFmtInvalidPatch                      = "invalid patch at index %d"
	errPatchFromFieldPath                   = "fromFieldPath is invalid"
	errPatchToFieldPath                     = "toFieldPath is invalid"
	errPatchRequireField                    = "missing field %s"
	errPatchCombineEmptyVariables           = "no variables given"
	errFmtPatchCombineVariableFromFieldPath = "fromFieldPath of variable at index %d is invalid"
	errUnknownPatchType                     = "unknown patch type %s"
	errParseRegisteredCompositePaths        = "cannot parse registered composite paths"
	errParseRegisteredComposedPaths         = "cannot parse registered composed paths"

	labelKeyClaimName      = "crossplane.io/claim-name"
	labelKeyClaimNamespace = "crossplane.io/claim-namespace"
)

var (
	// KnownCompositeAnnotations are annotations that will be registered by
	// default
	KnownCompositeAnnotations = []string{}

	// KnownCompositeLabels are labels that will be registered by default.
	KnownCompositeLabels = []string{
		labelKeyClaimName,
		labelKeyClaimNamespace,
	}
	// KnownResourceAnnotations are annotations that will be registered by
	// default
	KnownResourceAnnotations = []string{
		meta.AnnotationKeyExternalName,
		meta.AnnotationKeyExternalCreatePending,
		meta.AnnotationKeyExternalCreateSucceeded,
		meta.AnnotationKeyExternalCreateFailed,
	}
	// KnownResourceLabels are labels that will be registered by default.
	KnownResourceLabels = []string{}
)

// ComposedTemplateSkeleton represents the draft for a compositionSkeleton composeTemplateSkeleton.
type ComposedTemplateSkeleton interface {
	// WithName sets the name of this composeTemplateSkeleton.
	WithName(name string) ComposedTemplateSkeleton

	// WithPatches adds the following patches to this composeTemplateSkeleton.
	WithPatches(patches ...xapiextv1.Patch) ComposedTemplateSkeleton

	// WithUnsafePatches is similar to WithPatches but the field paths of the
	// composeTemplateSkeletons will not be validated.
	WithUnsafePatches(patches ...xapiextv1.Patch) ComposedTemplateSkeleton

	// WithConnectionDetails adds the following connection details to this
	// composeTemplateSkeleton.
	WithConnectionDetails(connectionDetails ...xapiextv1.ConnectionDetail) ComposedTemplateSkeleton

	// WithReadinessChecks adds the following readiness checks to this
	// composeTemplateSkeleton.
	WithReadinessChecks(checks ...xapiextv1.ReadinessCheck) ComposedTemplateSkeleton

	// RegisterAnnotations marks the given resource annotations as safe
	// so they will be treated as a valid field in patch paths.
	RegisterAnnotations(annotationKeys ...string) ComposedTemplateSkeleton

	// RegisterLabels marks the given resource label as safe
	// so they will be treated as a valid field in patch paths.
	RegisterLabels(labelsKeys ...string) ComposedTemplateSkeleton

	// RegisterFieldPaths marks the given resource paths as safe so ti will
	// be treated a valid in patch paths.
	RegisterFieldPaths(paths ...string) ComposedTemplateSkeleton
}

// CompositionSkeleton represents the build time state of a composition.
type CompositionSkeleton interface {
	// WithName sets the metadata.name of the composition to be built.
	WithName(name string) CompositionSkeleton

	// NewResource creates a new ComposedTemplateSkeleton with the given base.
	NewResource(base ObjectKindReference) ComposedTemplateSkeleton

	// WithPublishConnectionDetailsWithStoreConfig sets the
	// PublishConnectionDetailsWithStoreConfig of this CompositionSkeleton.
	WithPublishConnectionDetailsWithStoreConfig(ref *xapiextv1.StoreConfigReference) CompositionSkeleton

	// WithWriteConnectionSecretsToNamespace sets the
	// WriteConnectionSecretsToNamespace of this compositionSkeleton.
	WithWriteConnectionSecretsToNamespace(namespace *string) CompositionSkeleton

	// RegisterCompositeAnnotations marks the given composite annotations as safe
	// so it will be treated as a valid field in patch paths.
	RegisterCompositeAnnotations(annotationKeys ...string) CompositionSkeleton

	// RegisterCompositeLabels marks the given composite labels as safe
	// so it will be treated as a valid field in patch paths.
	RegisterCompositeLabels(labelKeys ...string) CompositionSkeleton

	// RegisterCompositeFieldPaths marks the given composite paths as safe so
	// they will be treated a valid in patch paths.
	RegisterCompositeFieldPaths(paths ...string) CompositionSkeleton
}

// Object is an extension of the k8s runtime.Object with additional functions
// that are required by Crossbuildec.
type Object interface {
	runtime.Object
	SetGroupVersionKind(gvk schema.GroupVersionKind)
}

// ObjectKindReference contains the group version kind and instance of a
// runtime.Object.
type ObjectKindReference struct {
	// GroupVersionKind is the GroupVersionKind for the composite type.
	GroupVersionKind schema.GroupVersionKind

	// Object is an instance of the composite type.
	Object Object
}

type compositionSkeleton struct {
	composite ObjectKindReference

	registeredPaths                         []string
	name                                    string
	composeTemplateSkeletons                []*composeTemplateSkeleton
	publishConnectionDetailsWithStoreConfig *xapiextv1.StoreConfigReference
	writeConnectionSecretsToNamespace       *string
}

// RegisterCompositeAnnotations marks the given composite annotations as safe so
// it will be treated as a valid field in patch paths.
func (c *compositionSkeleton) RegisterCompositeAnnotations(annotionKeys ...string) CompositionSkeleton {
	paths := make([]string, len(annotionKeys))
	for i, k := range annotionKeys {
		paths[i] = fmt.Sprintf("metadata.annotations[%s]", k)
	}
	return c.RegisterCompositeFieldPaths(paths...)
}

// RegisterCompositeLabels marks the given composite labels as safe so it
// will be treated as a valid field in patch paths.
func (c *compositionSkeleton) RegisterCompositeLabels(labelKeys ...string) CompositionSkeleton {
	paths := make([]string, len(labelKeys))
	for i, k := range labelKeys {
		paths[i] = fmt.Sprintf("metadata.labels[%s]", k)
	}
	return c.RegisterCompositeFieldPaths(paths...)
}

// RegisterCompositeFieldPaths marks the given composite paths as safe so ti will
// be treated a valid in patch paths.
func (c *compositionSkeleton) RegisterCompositeFieldPaths(path ...string) CompositionSkeleton {
	c.registeredPaths = append(c.registeredPaths, path...)
	return c
}

// WithName sets the metadata.name of the composition to be built.
func (c *compositionSkeleton) WithName(name string) CompositionSkeleton {
	c.name = name
	return c
}

// NewResource creates a new composeTemplateSkeleton with the given base.
func (c *compositionSkeleton) NewResource(base ObjectKindReference) ComposedTemplateSkeleton {
	res := &composeTemplateSkeleton{
		base:                base,
		compositionSkeleton: c,
	}
	c.composeTemplateSkeletons = append(c.composeTemplateSkeletons, res)
	return res
}

// WithPublishConnectionDetailsWithStoreConfig sets the
// PublishConnectionDetailsWithStoreConfig of this CompositionSkeleton.
func (c *compositionSkeleton) WithPublishConnectionDetailsWithStoreConfig(ref *xapiextv1.StoreConfigReference) CompositionSkeleton {
	c.publishConnectionDetailsWithStoreConfig = ref
	return c
}

// WithWriteConnectionSecretsToNamespace sets the
// WriteConnectionSecretsToNamespace of this compositionSkeleton.
func (c *compositionSkeleton) WithWriteConnectionSecretsToNamespace(namespace *string) CompositionSkeleton {
	c.writeConnectionSecretsToNamespace = namespace
	return c
}

// ToComposition generates a Crossplane compositionSkeleton from this compositionSkeleton.
func (c *compositionSkeleton) ToComposition() (xapiextv1.Composition, error) {
	if c.name == "" {
		return xapiextv1.Composition{}, errors.New(errEmptyCompositionname)
	}

	c.RegisterCompositeAnnotations(KnownCompositeAnnotations...)
	c.RegisterCompositeLabels(KnownCompositeLabels...)

	composedTemplates := make([]xapiextv1.ComposedTemplate, len(c.composeTemplateSkeletons))
	for i, c := range c.composeTemplateSkeletons {
		ct, err := c.ToComposedTemplate()
		if err != nil {
			return xapiextv1.Composition{}, errors.Wrapf(err, errFmtBuildComposedTemplate, i)
		}
		composedTemplates[i] = ct
	}

	comp := xapiextv1.Composition{
		Spec: xapiextv1.CompositionSpec{
			CompositeTypeRef:                           xapiextv1.TypeReferenceTo(c.composite.GroupVersionKind),
			Resources:                                  composedTemplates,
			WriteConnectionSecretsToNamespace:          c.writeConnectionSecretsToNamespace,
			PublishConnectionDetailsWithStoreConfigRef: c.publishConnectionDetailsWithStoreConfig,
		},
	}
	comp.SetGroupVersionKind(xapiextv1.CompositionGroupVersionKind)
	comp.SetName(c.name)
	comp.SetCreationTimestamp(v1.Time{})
	return comp, nil
}

type patchSkeleton struct {
	patch  xapiextv1.Patch
	unsafe bool
}

type composeTemplateSkeleton struct {
	compositionSkeleton *compositionSkeleton

	registeredPaths   []string
	name              *string
	base              ObjectKindReference
	patches           []patchSkeleton
	connectionDetails []xapiextv1.ConnectionDetail
	readinessChecks   []xapiextv1.ReadinessCheck
}

// RegisterAnnotations marks the given resource annotations as safe
// so they will be treated as a valid field in patch paths.
func (c *composeTemplateSkeleton) RegisterAnnotations(annotionKeys ...string) ComposedTemplateSkeleton {
	return c.RegisterFieldPaths(makeAnnotationPaths(annotionKeys)...)
}

// RegisterLabels marks the given resource labels as safe
// so they will be treated as a valid field in patch paths.
func (c *composeTemplateSkeleton) RegisterLabels(labelKeys ...string) ComposedTemplateSkeleton {
	return c.RegisterFieldPaths(makeLabelPaths(labelKeys)...)
}

// RegisterFieldPaths marks the given resource paths as safe so they will
// be treated a valid in patch paths.
func (c *composeTemplateSkeleton) RegisterFieldPaths(paths ...string) ComposedTemplateSkeleton {
	c.registeredPaths = append(c.registeredPaths, paths...)
	return c
}

// WithName sets the name of this composeTemplateSkeleton.
func (c *composeTemplateSkeleton) WithName(name string) ComposedTemplateSkeleton {
	c.name = &name
	return c
}

// WithPatches adds the following patches to this composeTemplateSkeleton.
func (c *composeTemplateSkeleton) WithPatches(patches ...xapiextv1.Patch) ComposedTemplateSkeleton {
	for _, patch := range patches {
		c.patches = append(c.patches, patchSkeleton{
			patch:  patch,
			unsafe: false,
		})
	}
	return c
}

// WithUnsafePatches is similar to WithPatches but the field paths of the
// composeTemplateSkeletons will not be validated.
func (c *composeTemplateSkeleton) WithUnsafePatches(patches ...xapiextv1.Patch) ComposedTemplateSkeleton {
	for _, patch := range patches {
		c.patches = append(c.patches, patchSkeleton{
			patch:  patch,
			unsafe: true,
		})
	}
	return c
}

// WithConnectionDetails adds the following connection details to this
// composeTemplateSkeleton.
func (c *composeTemplateSkeleton) WithConnectionDetails(connectionDetails ...xapiextv1.ConnectionDetail) ComposedTemplateSkeleton {
	c.connectionDetails = append(c.connectionDetails, connectionDetails...)
	return c
}

// WithReadinessChecks adds the following readiness checks to this composeTemplateSkeleton.
func (c *composeTemplateSkeleton) WithReadinessChecks(checks ...xapiextv1.ReadinessCheck) ComposedTemplateSkeleton {
	c.readinessChecks = append(c.readinessChecks, checks...)
	return c
}

// ToComposedTemplate converts this composeTemplateSkeleton into a ComposedTemplate.
func (c *composeTemplateSkeleton) ToComposedTemplate() (xapiextv1.ComposedTemplate, error) {
	registeredCompositePaths, err := parseFieldPaths(c.compositionSkeleton.registeredPaths)
	if err != nil {
		return xapiextv1.ComposedTemplate{}, errors.Wrap(err, errParseRegisteredCompositePaths)
	}
	registeredPaths, err := parseFieldPaths(c.registeredPaths)
	if err != nil {
		return xapiextv1.ComposedTemplate{}, errors.Wrap(err, errParseRegisteredComposedPaths)
	}

	c.RegisterAnnotations(KnownResourceAnnotations...)
	c.RegisterLabels(KnownResourceLabels...)

	patches := make([]xapiextv1.Patch, len(c.patches))
	for i, p := range c.patches {
		if !p.unsafe {
			if err := c.validatePatch(p.patch, registeredCompositePaths, registeredPaths); err != nil {
				return xapiextv1.ComposedTemplate{}, errors.Wrapf(err, errFmtInvalidPatch, i)
			}
		}
		patches[i] = p.patch
	}

	base := c.base.Object
	base.SetGroupVersionKind(c.base.GroupVersionKind)

	return xapiextv1.ComposedTemplate{
		Name: c.name,
		Base: runtime.RawExtension{
			Object: base,
		},
		Patches:           patches,
		ConnectionDetails: c.connectionDetails,
		ReadinessChecks:   c.readinessChecks,
	}, nil
}

func (c *composeTemplateSkeleton) validatePatch(patch xapiextv1.Patch, registeredCompositePaths, registeredPaths []fieldpath.Segments) error {
	patchType := patch.Type
	if patchType == "" {
		patchType = xapiextv1.PatchTypeFromCompositeFieldPath
	}

	switch patchType {
	case xapiextv1.PatchTypeFromCompositeFieldPath:
		return validatePatch(patch, c.compositionSkeleton.composite.Object, c.base.Object, registeredCompositePaths, registeredPaths)
	case xapiextv1.PatchTypeToCompositeFieldPath:
		return validatePatch(patch, c.base.Object, c.compositionSkeleton.composite.Object, registeredPaths, registeredCompositePaths)
	case xapiextv1.PatchTypeCombineFromComposite:
		return validatePatchCombine(patch, c.compositionSkeleton.composite.Object, c.base.Object, registeredCompositePaths, registeredPaths)
	case xapiextv1.PatchTypeCombineToComposite:
		return validatePatchCombine(patch, c.base.Object, c.compositionSkeleton.composite.Object, registeredPaths, registeredCompositePaths)
	case xapiextv1.PatchTypePatchSet:
		return errors.New("patch types not supported")
	}
	return errors.Errorf(errUnknownPatchType, patchType)
}

func validatePatch(patch xapiextv1.Patch, from, to runtime.Object, fromKnownPaths, toKnownPaths []fieldpath.Segments) error {
	if err := ValidateFieldPath(from, utils.StringValue(patch.FromFieldPath), fromKnownPaths); err != nil {
		return errors.Wrap(err, errPatchFromFieldPath)
	}
	if err := ValidateFieldPath(to, utils.StringValue(patch.ToFieldPath), toKnownPaths); err != nil {
		return errors.Wrap(err, errPatchToFieldPath)
	}
	return nil
}

func validatePatchCombine(patch xapiextv1.Patch, from, to runtime.Object, fromKnownPaths, toKnownPaths []fieldpath.Segments) error {
	if patch.Combine == nil {
		return errors.Errorf(errPatchRequireField, "combine")
	}
	if patch.Combine.Variables == nil {
		return errors.Errorf(errPatchRequireField, "combine.variables")
	}
	if len(patch.Combine.Variables) == 0 {
		return errors.New(errPatchCombineEmptyVariables)
	}

	for i, v := range patch.Combine.Variables {
		if err := ValidateFieldPath(from, v.FromFieldPath, fromKnownPaths); err != nil {
			return errors.Wrapf(err, errFmtPatchCombineVariableFromFieldPath, i)
		}
	}
	return errors.Wrap(ValidateFieldPath(to, utils.StringValue(patch.ToFieldPath), toKnownPaths), errPatchToFieldPath)
}

func makeLabelPaths(keys []string) []string {
	paths := make([]string, len(keys))
	for i, k := range keys {
		paths[i] = fmt.Sprintf("metadata.labels[%s]", k)
	}
	return paths
}

func makeAnnotationPaths(keys []string) []string {
	paths := make([]string, len(keys))
	for i, k := range keys {
		paths[i] = fmt.Sprintf("metadata.annotations[%s]", k)
	}
	return paths
}
