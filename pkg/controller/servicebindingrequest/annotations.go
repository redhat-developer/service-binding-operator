package servicebindingrequest

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/logging"
)

const (
	sbrNamespaceAnnotation = "service-binding-operator.apps.openshift.io/binding-namespace"
	sbrNameAnnotation      = "service-binding-operator.apps.openshift.io/binding-name"
)

var (
	annotationsLogger = logging.Logger("annotations")
)

// extractSBRNamespacedName returns a types.NamespacedName if the required service binding request keys
// are present in the given data
func extractSBRNamespacedName(data map[string]string) types.NamespacedName {
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
	sbrNamespacedName := types.NamespacedName{}
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return sbrNamespacedName, err
	}

	u := &unstructured.Unstructured{Object: data}

	sbrNamespacedName = extractSBRNamespacedName(u.GetAnnotations())
	log := annotationsLogger.WithValues(
		"Resource.GVK", u.GroupVersionKind(),
		"Resource.Namespace", u.GetNamespace(),
		"Resource.Name", u.GetName(),
		"SBR.NamespacedName", sbrNamespacedName.String(),
	)

	if IsNamespacedNameEmpty(sbrNamespacedName) {
		logging.Debug(&log, "SBR information not present in annotations, continue inspecting object")
	} else {
		// FIXME: Increase V level for tracing info to avoid flooding logs with this information.
		logging.Trace(&log, "SBR information found in annotations, returning it")
		return sbrNamespacedName, nil
	}

	if u.GroupVersionKind() == v1alpha1.SchemeGroupVersion.WithKind(ServiceBindingRequestKind) {
		logging.Debug(&log, "Object is a SBR, returning its namespaced name")
		sbrNamespacedName.Namespace = u.GetNamespace()
		sbrNamespacedName.Name = u.GetName()
		return sbrNamespacedName, nil
	}

	// FIXME: Increase V level for tracing info to avoid flooding logs with this information.
	logging.Trace(&log, "Object is not a SBR, returning an empty namespaced name")
	return sbrNamespacedName, nil
}

// SetSBRAnnotations update existing annotations to include operator's. The annotations added are
// referring to a existing SBR namespaced name.
func SetSBRAnnotations(
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

		log := annotationsLogger.WithValues(
			"SBR.Namespace", namespacedName.Namespace,
			"SBR.Name", namespacedName.Name,
			"Resource.GVK", gvk,
			"Resource.Namespace", obj.GetNamespace(),
			"Resource.Name", obj.GetName(),
		)
		logging.Debug(&log, "Updating resource annotations...")
		_, err := client.Resource(gvr).Namespace(obj.GetNamespace()).Update(obj, opts)
		if err != nil {
			logging.Error(err, &log, "unable to set/update annotations in object")
			return err
		}
	}
	return nil
}
