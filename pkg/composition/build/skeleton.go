package build

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	xapiextv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/mistermx/crossbuilder/pkg/utils"
)

const (
	errEmptyCompositionname                 = "composition name must not be empty"
	errFmtBuildComposedTemplate             = "cannot build composed template for composeTemplateSkeleton at index %d"
	errFmtInvalidPatch                      = "invalid patch at index %d"
	errPatchFromFieldPath                   = "fromFieldPath is invalid"
	errPatchToFieldPath                     = "toFieldPath is invalid"
	errPatchRequireField                    = "missing field %s"
	errPatchCombineEmptyVariables           = "no variables given"
	errFmtPatchCombineVariableFromFieldPath = "fromFieldPath of variable at index %d is invalid"
	errUnknownPatchType                     = "unknown patch type %s"
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
}

// CompositionSkeleton represents the build time state of a composition.
type CompositionSkeleton interface {
	// WithName sets the metadata.name of the composition to be built.
	WithName(name string) CompositionSkeleton

	// WithResource creates a new ComposedTemplateSkeleton with the given base.
	WithResource(base ObjectKindReference) ComposedTemplateSkeleton

	// WithPublishConnectionDetailsWithStoreConfig sets the
	// PublishConnectionDetailsWithStoreConfig of this CompositionSkeleton.
	WithPublishConnectionDetailsWithStoreConfig(ref *xpv1.Reference) CompositionSkeleton

	// WithWriteConnectionSecretsToNamespace sets the
	// WriteConnectionSecretsToNamespace of this compositionSkeleton.
	WithWriteConnectionSecretsToNamespace(namespace *string) CompositionSkeleton
}

// Object is an extension of the k8s runtime.Object with additional functions
// that are required by Crossbuilder.
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

	name                                    string
	composeTemplateSkeletons                []*composeTemplateSkeleton
	publishConnectionDetailsWithStoreConfig *xpv1.Reference
	writeConnectionSecretsToNamespace       *string
}

// WithName sets the metadata.name of the composition to be built.
func (c *compositionSkeleton) WithName(name string) CompositionSkeleton {
	c.name = name
	return c
}

// WithResource creates a new composeTemplateSkeleton with the given base.
func (c *compositionSkeleton) WithResource(base ObjectKindReference) ComposedTemplateSkeleton {
	res := &composeTemplateSkeleton{
		base:                base,
		compositionSkeleton: c,
	}
	c.composeTemplateSkeletons = append(c.composeTemplateSkeletons, res)
	return res
}

// WithPublishConnectionDetailsWithStoreConfig sets the
// PublishConnectionDetailsWithStoreConfig of this CompositionSkeleton.
func (c *compositionSkeleton) WithPublishConnectionDetailsWithStoreConfig(ref *xpv1.Reference) CompositionSkeleton {
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

	composedTemplates := make([]xapiextv1.ComposedTemplate, len(c.composeTemplateSkeletons))
	for i, r := range c.composeTemplateSkeletons {
		ct, err := r.ToComposedTemplate()
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

	name              *string
	base              ObjectKindReference
	patches           []patchSkeleton
	connectionDetails []xapiextv1.ConnectionDetail
	readinessChecks   []xapiextv1.ReadinessCheck
}

// WithName sets the name of this composeTemplateSkeleton.
func (r *composeTemplateSkeleton) WithName(name string) ComposedTemplateSkeleton {
	r.name = &name
	return r
}

// WithPatches adds the following patches to this composeTemplateSkeleton.
func (r *composeTemplateSkeleton) WithPatches(patches ...xapiextv1.Patch) ComposedTemplateSkeleton {
	for _, patch := range patches {
		r.patches = append(r.patches, patchSkeleton{
			patch:  patch,
			unsafe: false,
		})
	}
	return r
}

// WithUnsafePatches is similar to WithPatches but the field paths of the
// composeTemplateSkeletons will not be validated.
func (r *composeTemplateSkeleton) WithUnsafePatches(patches ...xapiextv1.Patch) ComposedTemplateSkeleton {
	for _, patch := range patches {
		r.patches = append(r.patches, patchSkeleton{
			patch:  patch,
			unsafe: true,
		})
	}
	return r
}

// WithConnectionDetails adds the following connection details to this
// composeTemplateSkeleton.
func (r *composeTemplateSkeleton) WithConnectionDetails(connectionDetails ...xapiextv1.ConnectionDetail) ComposedTemplateSkeleton {
	r.connectionDetails = append(r.connectionDetails, connectionDetails...)
	return r
}

// WithReadinessChecks adds the following readiness checks to this composeTemplateSkeleton.
func (r *composeTemplateSkeleton) WithReadinessChecks(checks ...xapiextv1.ReadinessCheck) ComposedTemplateSkeleton {
	r.readinessChecks = append(r.readinessChecks, checks...)
	return r
}

// ToComposedTemplate converts this composeTemplateSkeleton into a ComposedTemplate.
func (r *composeTemplateSkeleton) ToComposedTemplate() (xapiextv1.ComposedTemplate, error) {
	patches := make([]xapiextv1.Patch, len(r.patches))
	for i, p := range r.patches {
		if !p.unsafe {
			if err := r.validatePatch(p.patch); err != nil {
				return xapiextv1.ComposedTemplate{}, errors.Wrapf(err, errFmtInvalidPatch, i)
			}
		}
		patches[i] = p.patch
	}

	base := r.base.Object
	base.SetGroupVersionKind(r.base.GroupVersionKind)

	return xapiextv1.ComposedTemplate{
		Name: r.name,
		Base: runtime.RawExtension{
			Object: base,
		},
		Patches:           patches,
		ConnectionDetails: r.connectionDetails,
		ReadinessChecks:   r.readinessChecks,
	}, nil
}

func (r *composeTemplateSkeleton) validatePatch(patch xapiextv1.Patch) error {
	patchType := patch.Type
	if patchType == "" {
		patchType = xapiextv1.PatchTypeFromCompositeFieldPath
	}

	switch patchType {
	case xapiextv1.PatchTypeFromCompositeFieldPath:
		return validatePatch(patch, r.compositionSkeleton.composite.Object, r.base.Object)
	case xapiextv1.PatchTypeToCompositeFieldPath:
		return validatePatch(patch, r.base.Object, r.compositionSkeleton.composite.Object)
	case xapiextv1.PatchTypeCombineFromComposite:
		return validatePatchCombine(patch, r.compositionSkeleton.composite.Object, r.base.Object)
	case xapiextv1.PatchTypeCombineToComposite:
		return validatePatchCombine(patch, r.base.Object, r.compositionSkeleton.composite.Object)
	case xapiextv1.PatchTypePatchSet:
		return errors.New("patch types not supported")
	}
	return errors.Errorf(errUnknownPatchType, patchType)
}

func validatePatch(patch xapiextv1.Patch, from, to runtime.Object) error {
	if err := ValidateFieldPath(from, utils.StringValue(patch.FromFieldPath)); err != nil {
		return errors.Wrap(err, errPatchFromFieldPath)
	}
	if err := ValidateFieldPath(to, utils.StringValue(patch.ToFieldPath)); err != nil {
		return errors.Wrap(err, errPatchToFieldPath)
	}
	return nil
}

func validatePatchCombine(patch xapiextv1.Patch, from, to runtime.Object) error {
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
		if err := ValidateFieldPath(from, v.FromFieldPath); err != nil {
			return errors.Wrapf(err, errFmtPatchCombineVariableFromFieldPath, i)
		}
	}
	return errors.Wrap(ValidateFieldPath(to, utils.StringValue(patch.ToFieldPath)), errPatchToFieldPath)
}
