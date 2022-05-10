# Crossbuilder

Crossbuilder is a tool that allows the generation of Crossplane XRDs and 
composition from go code.

## XRD Generation

Crossbuilder's `xrd-gen` wraps around [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
and [controller-gen](https://github.com/kubernetes-sigs/controller-tools) but
instead of generating CRDs (Custom Resource Definitions) it generates XRDs
(Composite Resource Definitions) that are part of the Crossplane ecosystem.

Crossbuilder has every feature that Kubebuilder's CRD generator provides plus
the ability to define XRD specific fields.

Take a look at the [xrd-gen examples](./examples/xrd-gen/apis/generate.go) for
more details.

## Composition Generation

Crossbuilder provides a toolkit that allows building compositions from Go and
write them out as YAML.

Since go is a statically typed language, Crossbuilder is able to perform
additional validation checks, such as patch path validation, that is a common
cause of errors when writing Crossplane compositions.

See the [composition-gen examples](./examples/composition-gen/cmd/generate/generate.go)
to learn how to use it.
