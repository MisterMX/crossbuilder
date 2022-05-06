package build

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	xapiextv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

// CompositionWriter specifies the interface for a delegate that writes the
// generated composition to the target destination.
type CompositionWriter interface {
	// Write writes the given composition to the destintation output.
	Write(c xapiextv1.Composition) error
}

// NewWriterWriter creates a CompositionWriter that writes to the given
// io.Writer.
func NewWriterWriter(w io.Writer) CompositionWriter {
	return &writerWriter{
		writer: w,
	}
}

type writerWriter struct {
	writer io.Writer
}

func (w *writerWriter) Write(c xapiextv1.Composition) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	_, err = w.writer.Write(b)
	return err
}

// NewDirectoryWriter creates a new CompositionWriter that writes each
// composition to the given directory using the objects name as filename.
func NewDirectoryWriter(dir string) CompositionWriter {
	return &directoryWriter{
		dir: dir,
	}
}

type directoryWriter struct {
	dir string
}

func (w *directoryWriter) Write(c xapiextv1.Composition) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(w.dir, fs.FileMode(0777)); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.yaml", c.GetName())
	return ioutil.WriteFile(filepath.Join(w.dir, filename), b, fs.FileMode(0664))
}
