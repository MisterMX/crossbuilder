package build

import (
	xapiextv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
)

const (
	errWriteComposition    = "failed to write composition"
	errFmtBuildComposition = "failed to generate composition from skeleton at index %d"
)

// CompositionBuilder specifies the interface for user defined type that is
// able to build a composition.
type CompositionBuilder interface {
	// GetCompositeTypeRef returns the CompositeTypeReference for the
	// composition to be build.
	GetCompositeTypeRef() ObjectKindReference

	// Build builds a composition using the type.
	Build(composition CompositionSkeleton)
}

// RunnerConfig specifies a new composition runner config.
type RunnerConfig struct {
	Builder []CompositionBuilder
	Writer  CompositionWriter
}

// CompositionBuildRunner specifies the interface for a composition builder.
type CompositionBuildRunner interface {
	// Build generates all compositions from the builder delegates and sends
	// them to the output writer.
	Build() error
}

// NewRunner creates a new CompositionBuildRunner instance.
func NewRunner(config RunnerConfig) CompositionBuildRunner {
	return &compositionBuildRunner{
		config: config,
	}
}

type compositionBuildRunner struct {
	config RunnerConfig
}

// Build generates all compositions from the builders and sends them to the
// output writer.
func (b *compositionBuildRunner) Build() error {
	compositions := make([]xapiextv1.Composition, len(b.config.Builder))
	for i, builder := range b.config.Builder {
		compSkeleton := &compositionSkeleton{
			composite: builder.GetCompositeTypeRef(),
		}
		builder.Build(compSkeleton)

		comp, err := compSkeleton.ToComposition()
		if err != nil {
			return errors.Wrapf(err, errFmtBuildComposition, i)
		}
		compositions[i] = comp
	}

	for _, comp := range compositions {
		if err := b.config.Writer.Write(comp); err != nil {
			return errors.Wrap(err, errWriteComposition)
		}
	}
	return nil
}
