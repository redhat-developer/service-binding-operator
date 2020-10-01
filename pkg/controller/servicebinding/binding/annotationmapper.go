package binding

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
)

type annotationBackedDefinitionBuilder struct {
	kubeClient dynamic.Interface
	name       string
	value      string
}

var _ DefinitionBuilder = (*annotationBackedDefinitionBuilder)(nil)

type modelKey string

const (
	pathModelKey        modelKey = "path"
	objectTypeModelKey  modelKey = "objectType"
	sourceKeyModelKey   modelKey = "sourceKey"
	sourceValueModelKey modelKey = "sourceValue"
	elementTypeModelKey modelKey = "elementType"
	AnnotationPrefix             = "service.binding"
)

func (m *annotationBackedDefinitionBuilder) outputName() (string, error) {
	// bail out in the case the annotation name doesn't start with "service.binding"
	if m.name != AnnotationPrefix && !strings.HasPrefix(m.name, AnnotationPrefix+"/") {
		return "", fmt.Errorf("can't process annotation with name %q", m.name)
	}

	if p := strings.SplitN(m.name, "/", 2); len(p) > 1 && len(p[1]) > 0 {
		return p[1], nil
	}

	return "", nil
}

func (m *annotationBackedDefinitionBuilder) Build() (Definition, error) {

	outputName, err := m.outputName()
	if err != nil {
		return nil, err
	}

	mod, err := newModel(m.value)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create binding model for annotation key %s and value %s", m.name, m.value)
	}

	if len(outputName) == 0 {
		outputName = mod.path[len(mod.path)-1]
	}

	switch {
	case mod.isStringElementType() && mod.isStringObjectType():
		return &stringDefinition{
			outputName: outputName,
			path:       mod.path,
		}, nil

	case mod.isStringElementType() && mod.hasDataField():
		return &stringFromDataFieldDefinition{
			kubeClient: m.kubeClient,
			objectType: mod.objectType,
			outputName: outputName,
			path:       mod.path,
			sourceKey:  mod.sourceKey,
		}, nil

	case mod.isMapElementType() && mod.hasDataField():
		return &mapFromDataFieldDefinition{
			kubeClient:  m.kubeClient,
			objectType:  mod.objectType,
			outputName:  outputName,
			path:        mod.path,
			sourceValue: mod.sourceValue,
		}, nil

	case mod.isMapElementType() && mod.isStringObjectType():
		return &stringOfMapDefinition{
			outputName: outputName,
			path:       mod.path,
		}, nil

	case mod.isSliceOfMapsElementType():
		return &sliceOfMapsFromPathDefinition{
			outputName:  outputName,
			path:        mod.path,
			sourceKey:   mod.sourceKey,
			sourceValue: mod.sourceValue,
		}, nil

	case mod.isSliceOfStringsElementType():
		return &sliceOfStringsFromPathDefinition{
			outputName:  outputName,
			path:        mod.path,
			sourceValue: mod.sourceValue,
		}, nil
	}

	panic(fmt.Sprintf("Annotation %s=%s not implemented!", m.name, m.value))
}
