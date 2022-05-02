# Crossbuilder

Crossbuilder is a wrapper around [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) and [controller-gen](https://github.com/kubernetes-sigs/controller-tools) that allows generating Crossplane XRDs from Go.

## Features

Everything that kubebuilder's CRD generator can do plus defining XRD specific fields.

## How to use

Take a look at the [examples](./examples/xrd-gen/apis/generate.go) to see how to use crossbuilder.
