package example

import (
	v1 "github.com/crossplane/crossplane/v2/apis/apiextensions/v1"
	"k8s.io/utils/ptr"

	"github.com/mistermx/crossbuilder/v2/examples/xrd-gen/apis/v1alpha1"
	"github.com/mistermx/crossbuilder/v2/pkg/generate/composition/build"
)

type ExampleBuilder struct{}

func (b *ExampleBuilder) GetCompositeTypeRef() build.ObjectKindReference {
	return build.ObjectKindReference{
		GroupVersionKind: v1alpha1.XExampleGroupVersionKind,
		Object:           &v1alpha1.XExample{},
	}
}

func (b *ExampleBuilder) Build(c build.CompositionSkeleton) {
	c.
		WithName("cluster-role").
		WithWriteConnectionSecretsToNamespace(ptr.To("example-namespace")).
		WithPipelineSteps(
			v1.PipelineStep{
				Step: "example-step",
				FunctionRef: v1.FunctionReference{
					Name: "example-function",
				},
			})
}
