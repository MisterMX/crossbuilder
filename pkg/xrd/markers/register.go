package markers

import (
	"reflect"

	"sigs.k8s.io/controller-tools/pkg/markers"
)

type definitionWithHelp struct {
	*markers.Definition
	Help *markers.DefinitionHelp
}

func (d *definitionWithHelp) WithHelp(help *markers.DefinitionHelp) *definitionWithHelp {
	d.Help = help
	return d
}

func (d *definitionWithHelp) Register(reg *markers.Registry) error {
	if err := reg.Register(d.Definition); err != nil {
		return err
	}
	if d.Help != nil {
		reg.AddHelp(d.Definition, d.Help)
	}
	return nil
}

func must(def *markers.Definition, err error) *definitionWithHelp {
	return &definitionWithHelp{
		Definition: markers.Must(def, err),
	}
}

// AllDefinitions contains all marker definitions for this package.
var AllDefinitions []*definitionWithHelp

type hasHelp interface {
	Help() *markers.DefinitionHelp
}

// mustMakeAllWithPrefix converts each object into a marker definition using
// the object's type's with the prefix to form the marker name.
func mustMakeAllWithPrefix(prefix string, target markers.TargetType, objs ...interface{}) []*definitionWithHelp {
	defs := make([]*definitionWithHelp, len(objs))
	for i, obj := range objs {
		name := prefix + ":" + reflect.TypeOf(obj).Name()
		def, err := markers.MakeDefinition(name, target, obj)
		if err != nil {
			panic(err)
		}
		defs[i] = &definitionWithHelp{Definition: def, Help: obj.(hasHelp).Help()}
	}

	return defs
}

// Register registers all definitions for CRD generation to the given registry.
func Register(reg *markers.Registry) error {
	for _, def := range AllDefinitions {
		if err := def.Register(reg); err != nil {
			return err
		}
	}

	return nil
}
