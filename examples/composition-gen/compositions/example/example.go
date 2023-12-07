package example

import (
	"reflect"

	xv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/mistermx/crossbuilder/examples/xrd-gen/apis/v1alpha1"
	"github.com/mistermx/crossbuilder/pkg/generate/composition/build"
)

type ExampleBuilder struct{}

func (b *ExampleBuilder) GetCompositeTypeRef() build.ObjectKindReference {
	return build.ObjectKindReference{
		GroupVersionKind: v1alpha1.XExampleGroupVersionKind,
		Object:           &v1alpha1.XExample{},
	}
}

func (b *ExampleBuilder) Build(c build.CompositionSkeleton) {
	c.WithName("example")

	c.
		NewResource(build.ObjectKindReference{
			GroupVersionKind: rbacv1.SchemeGroupVersion.WithKind(reflect.TypeOf(rbacv1.ClusterRole{}).Name()),
			Object: &rbacv1.ClusterRole{
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"GET"},
						APIGroups: []string{"v1"},
						Resources: []string{""}, // patched
					},
					{
						Verbs:         []string{"GET"},
						APIGroups:     []string{"v1"},
						ResourceNames: []string{""}, // patched
					},
				},
			},
		}).
		RegisterLabels(
			"crossplane.io/claim-namespace",
			"crossplane.io/claim-name",
		).
		WithName("cluster-role").
		WithPatches(
			simplePatch(
				"spec.parameters.exampleField",
				"rules[0].resources[0]",
			),
			simplePatch(
				"spec.providerConfigRef.name",
				"rules[1].resourceNames[0]",
			),
			simplePatch(
				"metadata.labels[crossplane.io/claim-namespace]",
				"metadata.labels[crossplane.io/claim-namespace]",
			),
			simplePatch(
				"metadata.labels[crossplane.io/claim-name]",
				"metadata.labels[crossplane.io/claim-name]",
			),
		)
}

func simplePatch(from, to string) xv1.Patch {
	return xv1.Patch{
		FromFieldPath: &from,
		ToFieldPath:   &to,
	}
}
