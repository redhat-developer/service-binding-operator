package annotations

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const configMapValue = "binding:env:object:configmap"

// IsConfigMap returns true if the annotation value should trigger config map handler.
func isConfigMap(s string) bool {
	return configMapValue == s
}

// NewConfigMapHandler constructs an annotation handler that can extract related data from config
// maps.
func newConfigMapHandler(
	client dynamic.Interface,
	bi *bindingInfo,
	resource unstructured.Unstructured,
	restMapper meta.RESTMapper,
) (handler, error) {
	return NewResourceHandler(
		client,
		bi,
		resource,
		schema.GroupVersionResource{
			Version:  "v1",
			Resource: "configmaps",
		},
		&dataPath,
		restMapper,
	)
}
