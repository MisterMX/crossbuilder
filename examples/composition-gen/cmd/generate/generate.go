package main

import (
	"log"

	"github.com/mistermx/crossbuilder/v2/examples/composition-gen/compositions/example"
	"github.com/mistermx/crossbuilder/v2/pkg/generate/composition/build"
)

func main() {
	runner := build.NewRunner(build.RunnerConfig{
		Writer: build.NewDirectoryWriter("../../package/compositions"),
		Builder: []build.CompositionBuilder{
			&example.ExampleBuilder{},
		},
	})

	if err := runner.Build(); err != nil {
		log.Fatal(err)
	}
}
