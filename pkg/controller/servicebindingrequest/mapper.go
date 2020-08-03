package servicebindingrequest

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

var (
	mapperLog = log.NewLog("mapper")
)

// sbrRequestMapper is the handler.Mapper interface implementation. It should influence the
// enqueue process considering the resources informed.
type sbrRequestMapper struct {
	client     dynamic.Interface
	restMapper meta.RESTMapper
}

var serviceBindingRequestGVK = v1alpha1.SchemeGroupVersion.WithKind("ServiceBindingRequest")
var secretGVK = corev1.SchemeGroupVersion.WithKind("Secret")

// isServiceBindingRequest checks whether the given obj is a Service Binding Request through GVK
// comparison.
func isServiceBindingRequest(obj runtime.Object) bool {
	return obj.GetObjectKind().GroupVersionKind() == serviceBindingRequestGVK
}

// isSecret checks whether the given obj is a Secret through GVK comparison.
func isSecret(obj runtime.Object) bool {
	return obj.GetObjectKind().GroupVersionKind() == secretGVK
}

// isSBRService checks whether the given obj is a service in given sbr.
func isSBRService(sbr *v1alpha1.ServiceBindingRequest, obj runtime.Object) bool {
	services := extractServiceSelectors(sbr)
	for _, svc := range services {
		svcGVK := schema.GroupVersionKind{Group: svc.Group, Version: svc.Version, Kind: svc.Kind}
		if obj.GetObjectKind().GroupVersionKind() == svcGVK {
			return true
		}
	}
	return false
}

// isApplication checks whether the given obj is an application in given sbr.
func isApplication(
	restMapper meta.RESTMapper,
	sbr *v1alpha1.ServiceBindingRequest,
	obj runtime.Object,
) (bool, error) {
	appGVR := schema.GroupVersionResource{
		Group:    sbr.Spec.ApplicationSelector.Group,
		Version:  sbr.Spec.ApplicationSelector.Version,
		Resource: sbr.Spec.ApplicationSelector.Resource,
	}
	appGVK, err := restMapper.KindFor(appGVR)
	if err != nil {
		return false, err
	}

	isEqual := obj.GetObjectKind().GroupVersionKind() == appGVK

	return isEqual, nil
}

// isSecretOwnedBySBR checks whether the given obj is a secret owned by the given sbr.
func isSecretOwnedBySBR(obj metav1.Object, sbr *v1alpha1.ServiceBindingRequest) bool {
	return sbr.GetNamespace() == obj.GetNamespace() && sbr.Status.Secret == obj.GetName()
}

// convertToSBR attempts to convert the given obj into a Service Binding Request.
func convertToSBR(obj map[string]interface{}) (*v1alpha1.ServiceBindingRequest, error) {
	sbr := &v1alpha1.ServiceBindingRequest{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, sbr)
	return sbr, err
}

// convertToNamespacedName returns a NamespacedName with information extracted from given obj.
func convertToNamespacedName(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

// namespacedNameSet is a set of NamespacedNames.
type namespacedNameSet map[types.NamespacedName]bool

// add adds the given namespaced name n into the set.
func (t namespacedNameSet) add(n types.NamespacedName) {
	t[n] = true
}

func convertToRequests(t namespacedNameSet) []reconcile.Request {
	toReconcile := make([]reconcile.Request, 0)
	for n := range t {
		toReconcile = append(
			toReconcile,
			reconcile.Request{NamespacedName: n},
		)
	}
	return toReconcile

}

// Map execute the mapping of a resource with the requests it would produce. Here we inspect the
// given object trying to identify if this object is part of a SBR, or a actual SBR resource.
func (m *sbrRequestMapper) Map(obj handler.MapObject) []reconcile.Request {
	log := mapperLog.WithValues(
		"Object.Namespace", obj.Meta.GetNamespace(),
		"Object.Name", obj.Meta.GetName(),
	)

	namespacedNamesToReconcile := make(namespacedNameSet)

	if isServiceBindingRequest(obj.Object) {
		requests := []reconcile.Request{
			{NamespacedName: convertToNamespacedName(obj.Meta)},
		}
		log.Debug("current resource is a SBR", "Requests", requests)
		return requests
	}

	sbrList, err := m.client.Resource(groupVersion).List(metav1.ListOptions{})
	if err != nil {
		log.Error(err, "listing SBRs")
		// NOTE: nothing to do anymore if obj isn't a SBR and there's an error listing SBRs; I'm not
		// sure whether retrying in the same reconcile loop is the proper approach so we'll rely on
		// the resync period to re-process all watched objects.
		return []reconcile.Request{}
	}

ITEMS:
	for _, item := range sbrList.Items {
		namespacedName := convertToNamespacedName(&item)

		sbr, err := convertToSBR(item.Object)
		if err != nil {
			log.Error(err, "converting unstructured to SBR")
			continue ITEMS
		}

		if isSecret(obj.Object) && isSecretOwnedBySBR(obj.Meta, sbr) {
			log.Debug("current resource is a secret declared in SBR")
			namespacedNamesToReconcile.add(namespacedName)
		}

		if isSBRService(sbr, obj.Object) {
			log.Debug("resource is declared as service in SBR")
			namespacedNamesToReconcile.add(namespacedName)
		} else {
			log.Debug("resource does not match any declared service")
		}

		if ok, err := isApplication(m.restMapper, sbr, obj.Object); err != nil {
			log.Error(err, "verifying if resource is an application in SBR")
			continue ITEMS
		} else if !ok {
			log.Debug("not an application")
			continue ITEMS
		} else {
			namespacedNamesToReconcile.add(namespacedName)
		}
	}

	requests := convertToRequests(namespacedNamesToReconcile)
	log.Debug("did finish request mapper", "Requests", requests)
	return requests
}
