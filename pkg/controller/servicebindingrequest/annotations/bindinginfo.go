package annotations

import (
	"errors"
	"strings"
)

const (
	// ServiceBindingOperatorAnnotationPrefix is the prefix of Service Binding Operator related annotations.
	ServiceBindingOperatorAnnotationPrefix = "servicebindingoperator.redhat.io/"
	ServiceBindingSpecAnnotationPrefix     = "servicebinding.dev/"
)

// BindingInfo represents the pieces of a binding as parsed from an annotation.
type BindingInfo struct {
	// ResourceReferencePath is the field in the Backing Service CR referring to a bindable property, either
	// embedded or a reference to a related object..
	ResourceReferencePath string
	// SourcePath is the field that will be collected from the Backing Service CR or a related object.
	SourcePath string
	// Descriptor is the field reference to another manifest.
	Descriptor string
	// Value is the original annotation value.
	Value string
	// Value to be fetched from a secret/ConfigMap
	ObjectType string
	// Extract a specific field from the configmap/Secret from the Kubernetes resource and map it to different name in the binding Secret
	SourceKey string
	// Specifies if the element is to be bound as an environment variable or a volume mount using the keywords envVar and volume
	BindAs bindingType // by default should be an envVar
}

var ErrInvalidAnnotationPrefix = errors.New("invalid annotation prefix")
var ErrInvalidAnnotationName = errors.New("invalid annotation name")
var ErrEmptyAnnotationName = errors.New("empty annotation name")

// NewBindingInfo parses the encoded in the annotation name, returning its parts.
func NewBindingInfo(name string, value string) (*BindingInfo, error) {
	// do not process unknown annotations
	if strings.HasPrefix(name, ServiceBindingOperatorAnnotationPrefix) {

		cleanName := strings.TrimPrefix(name, ServiceBindingOperatorAnnotationPrefix)
		if len(cleanName) == 0 {
			return nil, ErrEmptyAnnotationName
		}

		parts := strings.SplitN(cleanName, "-", 2)

		resourceReferencePath := parts[0]
		sourcePath := parts[0]

		// the annotation is a reference to another manifest
		if len(parts) == 2 {
			sourcePath = parts[1]
		}

		return &BindingInfo{
			ResourceReferencePath: resourceReferencePath,
			SourcePath:            sourcePath,
			Descriptor:            strings.Join([]string{value, sourcePath}, ":"),
			Value:                 value,
		}, nil
	}
	if strings.HasPrefix(name, ServiceBindingSpecAnnotationPrefix) {
		var sourceKey string
		var resourceReferencePath string
		var sourcePath string
		// "servicebinding.dev/dbCredentials": "path={.status.data.dbCredentials},objectType=Secret"

		split := func(r rune) bool {
			return r == '=' || r == ','
		}
		a := strings.FieldsFunc(value, split)
		m := map[string]string{}

		for i := 0; i < len(a); i = i + 2 {

			k := a[i]
			val := a[i+1]
			m[k] = val
		}
		varReference := strings.TrimPrefix(name, ServiceBindingSpecAnnotationPrefix)
		// This will contain dbCredentials(entire secret reference), can be a specific data key from secret/CM
		if len(varReference) == 0 {
			return nil, ErrEmptyAnnotationName
		}
		str := strings.Trim(m["path"], "{")
		path := strings.Trim(str, "}")
		if m["sourceKey"] == "" {
			sourceKey = varReference
		} else {
			sourceKey = m["sourceKey"]
		}
		resourceReferencePath = path

		if m["objectType"] == "Secret" || m["objectType"] == "ConfigMap" {
			sourcePath = varReference
		} else if m["objectType"] == "" { //attribute
			sourcePath = resourceReferencePath
		} else {
			// error
		}

		return &BindingInfo{
			ObjectType:            m["objectType"],
			SourceKey:             sourceKey,
			ResourceReferencePath: resourceReferencePath,
			BindAs:                bindingType(m["bindAs"]),
			SourcePath:            sourcePath,
			Value:                 value,
		}, nil
	}

	return nil, ErrInvalidAnnotationPrefix
}
