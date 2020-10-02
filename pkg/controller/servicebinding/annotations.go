package servicebinding

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	sbrNamespaceAnnotation = "service-binding-operator.operators.coreos.com/binding-namespace"
	sbrNameAnnotation      = "service-binding-operator.operators.coreos.com/binding-name"
)

var (
	annotationsLog = log.NewLog("annotations")
)

// updateUnstructuredObj generic call to update the unstructured resource informed. It can return
// error when API update call does.
func updateUnstructuredObj(client dynamic.Interface, obj *unstructured.Unstructured) error {
	gvk := obj.GroupVersionKind()
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	opts := metav1.UpdateOptions{}

	log := annotationsLog.WithValues(
		"SBR.Namespace", obj.GetNamespace(),
		"SBR.Name", obj.GetName(),
		"Resource.GVK", gvk,
		"Resource.Namespace", obj.GetNamespace(),
		"Resource.Name", obj.GetName(),
	)
	log.Debug("Updating resource annotations...")

	_, err := client.Resource(gvr).Namespace(obj.GetNamespace()).Update(obj, opts)
	if err != nil {
		log.Error(err, "unable to set/update annotations in object")
	}
	return err
}

// setSBRAnnotations set annotations to include SBR information and return a new object.
func setSBRAnnotations(namespacedName types.NamespacedName,
	obj *unstructured.Unstructured) *unstructured.Unstructured {
	newObj := obj.DeepCopy()
	annotations := newObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[sbrNamespaceAnnotation] = namespacedName.Namespace
	annotations[sbrNameAnnotation] = namespacedName.Name
	newObj.SetAnnotations(annotations)
	return newObj
}

// removeSBRAnnotations removes SBR related annotations and return a new object.
func removeSBRAnnotations(obj *unstructured.Unstructured) *unstructured.Unstructured {
	newObj := obj.DeepCopy()
	annotations := newObj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	delete(annotations, sbrNameAnnotation)
	delete(annotations, sbrNamespaceAnnotation)
	newObj.SetAnnotations(annotations)
	return newObj
}

// removeAndUpdateSBRAnnotations removes SBR related annotations from all the objects and updates them using
// the given client.
func removeAndUpdateSBRAnnotations(client dynamic.Interface, objs []*unstructured.Unstructured) error {
	for _, obj := range objs {
		newObj := removeSBRAnnotations(obj)
		equal, err := nestedUnstructuredComparison(obj, newObj, []string{"metadata", "annotations"}...)
		if err != nil {
			return err
		}
		if !equal.Success {
			if err := updateUnstructuredObj(client, newObj); err != nil {
				return err
			}
		}
	}
	return nil
}
