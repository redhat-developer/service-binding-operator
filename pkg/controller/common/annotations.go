package common

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const (
	sbrNamespaceAnnotation = "service-binding-operator.apps.openshift.io/binding-namespace"
	sbrNameAnnotation      = "service-binding-operator.apps.openshift.io/binding-name"
)

func extractNamespacedName(data map[string]string) types.NamespacedName {
	ns, exists := data[sbrNamespaceAnnotation]
	if !exists {
		return types.NamespacedName{}
	}
	name, exists := data[sbrNameAnnotation]
	if !exists {
		return types.NamespacedName{}
	}
	return types.NamespacedName{Namespace: ns, Name: name}
}

func GetSBRSelectorFromObject(obj runtime.Object) (types.NamespacedName, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return types.NamespacedName{}, err
	}

	u := &unstructured.Unstructured{Object: data}
	return extractNamespacedName(u.GetAnnotations()), nil
}

func IsSBRSelectorEmpty(namespacedName types.NamespacedName) bool {
	return namespacedName.Namespace == "" || namespacedName.Name == ""
}

func SetSBRSelectorInObject() error {
	return nil
}
