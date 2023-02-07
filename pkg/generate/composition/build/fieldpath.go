package build

import (
	"reflect"
	"strings"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/pkg/errors"
)

const (
	errEmptyPath           = "the given path is empty"
	errParseFieldPath      = "cannot parse fieldpath"
	errFmtNotStruct        = "expected struct type, but got %s"
	errFmtNotArrayOrSlice  = "expected array or slice type but got %s"
	errFmtFieldNotFound    = "no field with JSON key '%s'"
	errGetStructField      = "cannot get field"
	errMapTypeNotSupported = "static path validation is not supported for maps"
	errFmtInvalidFieldPath = "invalid field path '%s'"
)

// ValidateFieldPath checks if the JSON path exists for the given object.
func ValidateFieldPath(obj interface{}, path string, knownPaths []fieldpath.Segments) error {
	segments, err := fieldpath.Parse(path)
	if err != nil {
		return errors.Wrap(err, errParseFieldPath)
	}
	if len(segments) == 0 {
		return errors.New(errEmptyPath)
	}
	if isKnownPath(segments, knownPaths) {
		return nil // path is a registered path
	}

	return errors.Wrap(validatePath(obj, segments), path)
}

func validatePath(obj interface{}, segments fieldpath.Segments) error {
	current := reflect.TypeOf(obj)
	for _, segment := range segments {
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}

		switch segment.Type {
		case fieldpath.SegmentField:
			var err error
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

func isKnownPath(path fieldpath.Segments, knownPaths []fieldpath.Segments) bool {
	for _, known := range knownPaths {
		if len(path) != len(known) {
			continue
		}
		for i, knownSeg := range known {
			if path[i] != knownSeg {
				continue
			}
		}
		return true
	}
	return false
}

func parseFieldPaths(paths []string) ([]fieldpath.Segments, error) {
	parsed := make([]fieldpath.Segments, len(paths))
	for i, p := range paths {
		seg, err := fieldpath.Parse(p)
		if err != nil {
			return nil, errors.Wrapf(err, errFmtInvalidFieldPath, p)
		}
		parsed[i] = seg
	}
	return parsed, nil
}

func getObjectField(obj reflect.Type, jsonKey string) (reflect.Type, error) {
	if obj.Kind() == reflect.Map {
		return nil, errors.New(errMapTypeNotSupported)
	}
	if obj.Kind() != reflect.Struct {
		return nil, errors.Errorf(errFmtNotStruct, obj.Kind())
	}
	for i := 0; i < obj.NumField(); i++ {
		field := obj.Field(i)
		tag := field.Tag.Get("json")
		name, options := parseTag(tag)

		if options.Contains("inline") {
			res, err := getObjectField(field.Type, jsonKey)
			if err == nil && res != nil {
				return res, nil
			}
		} else if name == jsonKey {
			return field.Type, nil
		}
	}
	return nil, errors.Errorf(errFmtFieldNotFound, jsonKey)
}

// The following code is extracted from
// https://cs.opensource.google/go/go/+/release-branch.go1.17:src/encoding/json/tags.go

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}
