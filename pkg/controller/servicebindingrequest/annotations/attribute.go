package annotations

import (
	"strings"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/nested"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const attributeValue = "binding:env:attribute"

// attributeHandler handles "binding:env:attribute" annotations.
type attributeHandler struct {
	// inputPath is the path that should be extracted from the resource. Required.
	inputPath string
	// outputPath is the path the extracted data should be placed under in the
	// resulting unstructured object in Handler. Required.
	outputPath string
	// resource is the unstructured object to extract data using inputPath. Required.
	resource unstructured.Unstructured
}

// Handle returns a unstructured object according to the "binding:env:attribute"
// annotation strategy.
func (h *attributeHandler) Handle() (result, error) {
	val, _, err := nested.GetValue(h.resource.Object, h.inputPath, h.outputPath)
	if err != nil {
		return result{}, err
	}
	return result{
		Data: val,
	}, nil
}

// IsAttribute returns true if the annotation value should trigger the attribute
// handler.
func isAttribute(s string) bool {
	return attributeValue == s
}

// NewAttributeHandler constructs an AttributeHandler.
func newAttributeHandler(
	bi *bindingInfo,
	resource unstructured.Unstructured,
) *attributeHandler {
	outputPath := bi.SourcePath
	if len(bi.ResourceReferencePath) > 0 {
		outputPath = bi.ResourceReferencePath
	}
	inputPath := bi.SourcePath
	if inputPath == "" {
		inputPath = bi.ResourceReferencePath
	}

	// the current implementation removes "status." and "spec." from fields exported through
	// annotations.
	for _, prefix := range []string{"status.", "spec."} {
		if strings.HasPrefix(outputPath, prefix) {
			outputPath = strings.Replace(outputPath, prefix, "", 1)
		}
	}

	return &attributeHandler{
		inputPath:  inputPath,
		outputPath: outputPath,
		resource:   resource,
	}
}
