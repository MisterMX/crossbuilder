package build

import (
	"reflect"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/pkg/errors"
)

const (
	errEmptyPath          = "the given path is empty"
	errParseFieldPath     = "cannot parse fieldpath"
	errFmtNotStruct       = "expected struct type, but got %s"
	errFmtNotArrayOrSlice = "expected array or slice type but got %s"
	errFmtFieldNotFound   = "cannot find a field with the JSON key '%s'"
	errGetStructField     = "cannot get struct field"
)

// ValidateFieldPath checks if the JSON path exists for the given object.
func ValidateFieldPath(obj interface{}, path string) error {
	segments, err := fieldpath.Parse(path)
	if err != nil {
		return errors.Wrap(err, errParseFieldPath)
	}
	if len(segments) == 0 {
		return errors.New(errEmptyPath)
	}

	current := reflect.TypeOf(obj)
	for _, segment := range segments {
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}

		switch segment.Type {
		case fieldpath.SegmentField:
			current, err = getObjectField(current, segment.Field)
			if err != nil {
				return errors.Wrap(err, errGetStructField)
			}
		case fieldpath.SegmentIndex:
			if current.Kind() != reflect.Array && current.Kind() != reflect.Slice {
				return errors.Errorf(errFmtNotArrayOrSlice, current.Kind())
			}
			current = current.Elem()
		}
	}
	return nil // Path exists
}

func getObjectField(obj reflect.Type, jsonKey string) (reflect.Type, error) {
	if obj.Kind() != reflect.Struct {
		return nil, errors.Errorf(errFmtNotStruct, obj.Kind())
	}
	for i := 0; i < obj.NumField(); i++ {
		field := obj.Field(i)
		tag := field.Tag.Get("json")
		name, _ := parseTag(tag)

		if name == jsonKey {
			return field.Type, nil
		}
	}
	return nil, errors.Errorf(errFmtFieldNotFound, jsonKey)
}

// parseTag splits the given JSON tag into its name and options
// extracted from
// https://cs.opensource.google/go/go/+/dev.boringcrypto.go1.17:src/encoding/json/tags.go
func parseTag(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}
