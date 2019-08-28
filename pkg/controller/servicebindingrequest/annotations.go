package servicebindingrequest

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

const (
	sbrNamespaceAnnotation = "service-binding-operator.apps.openshift.io/binding-namespace"
	sbrNameAnnotation      = "service-binding-operator.apps.openshift.io/binding-name"
)

// extractNamespacedName returns a types.NamespacedName if the required service binding request keys
// are present in the given data
func extractNamespacedName(data map[string]string) types.NamespacedName {
	namespacedName := types.NamespacedName{}
	ns, exists := data[sbrNamespaceAnnotation]
	if !exists {
		return namespacedName
	}
	name, exists := data[sbrNameAnnotation]
	if !exists {
		return namespacedName
	}
	namespacedName.Namespace = ns
	namespacedName.Name = name
	return namespacedName
}

// GetSBRNamespacedNameFromObject returns a types.NamespacedName if the required service binding
// request annotations are present in the given runtime.Object, empty otherwise. When annotations are
// not present, it checks if the object is an actual SBR, returning the details when positive. An
// error can be returned in the case the object can't be decoded.
func GetSBRNamespacedNameFromObject(obj runtime.Object) (types.NamespacedName, error) {
	namespacedName := types.NamespacedName{}
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return namespacedName, err
	}

	u := &unstructured.Unstructured{Object: data}

	namespacedName = extractNamespacedName(u.GetAnnotations())
	logger := log.WithValues(
		"Resource.GVK", u.GroupVersionKind(),
		"Resource.Namespace", u.GetNamespace(),
		"Resource.Name", u.GetName(),
		"Extracted.NamespacedName", namespacedName.String(),
	)
	if !IsSBRNamespacedNameEmpty(namespacedName) {
		logger.Info("Not able to define SBR namespaced-name based on annotations!")
		return namespacedName, nil
	}

	if u.GroupVersionKind() != v1alpha1.SchemeGroupVersion.WithKind(ServiceBindingRequestKind) {
		logger.Info("Object is also not a SBR resource type.")
		return namespacedName, nil
	}

	logger.Info("Creating namespaced-name for a actual SBR object.")
	namespacedName.Namespace = u.GetNamespace()
	namespacedName.Name = u.GetName()
	return namespacedName, nil
}

// IsSBRNamespacedNameEmpty returns true if any of the fields from the given namespacedName is empty.
func IsSBRNamespacedNameEmpty(namespacedName types.NamespacedName) bool {
	return namespacedName.Namespace == "" || namespacedName.Name == ""
}

// SetSBRAnnotations update existing annotations to include operator's. The annotations added are
// referring to a existing SBR namespaced name.
func SetSBRAnnotations(
	ctx context.Context,
	client dynamic.Interface,
	namespacedName types.NamespacedName,
	objs []*unstructured.Unstructured,
) error {
	for _, obj := range objs {
		annotations := obj.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}

		annotations[sbrNamespaceAnnotation] = namespacedName.Namespace
		annotations[sbrNameAnnotation] = namespacedName.Name
		obj.SetAnnotations(annotations)

		gvk := obj.GroupVersionKind()
		gvr, _ := meta.UnsafeGuessKindToResource(gvk)
		opts := metav1.UpdateOptions{}

		logger := log.WithValues(
			"SBR.Namespace", namespacedName.Namespace,
			"SBR.Name", namespacedName.Name,
			"Resource.GVK", gvk,
			"Resource.Namespace", obj.GetNamespace(),
			"Resource.Name", obj.GetName(),
		)
		logger.Info("Updating resource annotations...")
		_, err := client.Resource(gvr).Namespace(obj.GetNamespace()).Update(obj, opts)
		if err != nil {
			logger.Error(err, "unable to set/update annotations in object")
			return err
		}
	}
	return nil
}
