package xrd

import (
	"io"
	"io/ioutil"

	"github.com/pkg/errors"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/yaml"

	xbuilderio "github.com/mistermx/crossbuilder/pkg/utils/io"
)

const (
	errParseCRD   = "failed to parse generated CRD"
	errReadResult = "failed to read result"
)

func newCRDStorage() crdStorage {
	return crdStorage{
		CRDs: map[string]*apiext.CustomResourceDefinition{},
	}
}

type crdStorage struct {
	CRDs map[string]*apiext.CustomResourceDefinition
}

func (s *crdStorage) OutputRule() *crdStorageOutputRule {
	return &crdStorageOutputRule{
		storage: s,
	}
}

type crdStorageOutputRule struct {
	storage *crdStorage
}

func (o *crdStorageOutputRule) Open(pkg *loader.Package, itemPath string) (io.WriteCloser, error) {
	writer := xbuilderio.NewOnCloseWriter(nil, func(r io.Reader, len int64) (err error) {
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrap(err, errReadResult)
		}
		crd := &apiext.CustomResourceDefinition{}
		if err := yaml.Unmarshal(data, crd); err != nil {
			return errors.Wrap(err, errParseCRD)
		}
		o.storage.CRDs[crd.GetName()] = crd
		return nil
	})
	return writer, nil
}
