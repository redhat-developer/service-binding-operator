package servicebindingrequest

import (
	"fmt"
	"strings"
)

// BindingInfo represents the pieces of a binding as parsed from an annotation.
type BindingInfo struct {
	// FieldPath is the field in the Backing Service CR referring to a bindable property, either
	// embedded or a reference to a related object..
	FieldPath string
	// Path is the field that will be collected from the Backing Service CR or a related object.
	Path string
	// Descriptor is the field reference to another manifest
	Descriptor string
}

// NewBindingInfo parses the encoded in the annotation name, returning its parts.
func NewBindingInfo(name string, value string) (*BindingInfo, error) {
	cleanName := strings.TrimPrefix(name, ServiceBindingOperatorAnnotationPrefix)
	parts := strings.SplitN(cleanName, "-", 2)

	// if there is only one part, it means the value of the referenced field itself will be used
	if len(parts) == 1 {
		return &BindingInfo{
			FieldPath:  parts[0],
			Path:       parts[0],
			Descriptor: strings.Join([]string{value, parts[0]}, ":"),
		}, nil
	}

	// the annotation is a reference to another manifest
	if len(parts) == 2 {
		return &BindingInfo{
			FieldPath:  parts[0],
			Path:       parts[1],
			Descriptor: strings.Join([]string{value, parts[1]}, ":"),
		}, nil
	}

	return nil, fmt.Errorf("should have two parts")
}
