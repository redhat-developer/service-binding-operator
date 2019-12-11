package servicebindingrequest

import (
	"fmt"
	"strings"

	omlv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

// BindingInfo represents the pieces of a binding as parsed from an annotation.
type BindingInfo struct {
	// FieldPath is the field in the Backing Service CR referring to a bindable property, either
	// embedded or a reference to a related object..
	FieldPath string
	// Path is the field that will be collected from the Backing Service CR or a related object.
	Path string
}

// TrimmedPath returns a copy of Path without the given prefix.
func (b *BindingInfo) TrimmedPath(prefix string) string {
	prefix = fmt.Sprintf("%s.", prefix)
	return strings.TrimPrefix(b.Path, prefix)
}

// StatusDescriptor creates a StatusDescriptor.
func (b *BindingInfo) StatusDescriptor(value string) omlv1alpha1.StatusDescriptor {
	xDescriptor := strings.Join([]string{value, b.FieldPath}, ":")

	statusDescriptor := omlv1alpha1.StatusDescriptor{
		Path:         b.TrimmedPath("status"),
		XDescriptors: []string{xDescriptor},
	}

	return statusDescriptor
}

// SpecDescriptor creates a SpecDescriptor.
func (b *BindingInfo) SpecDescriptor(value string) omlv1alpha1.SpecDescriptor {
	xDescriptor := strings.Join([]string{value, b.FieldPath}, ":")

	statusDescriptor := omlv1alpha1.SpecDescriptor{
		Path:         b.TrimmedPath("spec"),
		XDescriptors: []string{xDescriptor},
	}

	return statusDescriptor
}

// NewBindingInfo parses the encoded in the annotation name, returning its parts.
func NewBindingInfo(annotationName string) (*BindingInfo, error) {
	cleanAnnotationName := strings.TrimPrefix(annotationName, ServiceBindingOperatorAnnotationPrefix)
	parts := strings.SplitN(cleanAnnotationName, "-", 2)

	if len(parts) == 1 {
		return &BindingInfo{FieldPath: parts[0], Path: parts[0]}, nil
	}

	if len(parts) == 2 {
		return &BindingInfo{FieldPath: parts[0], Path: parts[1]}, nil
	}

	return nil, fmt.Errorf("should have two parts")
}
