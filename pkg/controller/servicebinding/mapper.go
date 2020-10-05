package servicebinding

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

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
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

var serviceBindingRequestGVK = v1alpha1.SchemeGroupVersion.WithKind("ServiceBinding")
var secretGVK = corev1.SchemeGroupVersion.WithKind("Secret")

// isServiceBinding checks whether the given obj is a Service Binding through GVK
// comparison.
func isServiceBinding(obj runtime.Object) bool {
	return obj.GetObjectKind().GroupVersionKind() == serviceBindingRequestGVK
}

// isSecret checks whether the given obj is a Secret through GVK comparison.
func isSecret(obj runtime.Object) bool {
	return obj.GetObjectKind().GroupVersionKind() == secretGVK
}

// isSBRService checks whether the given obj is a service in given sbr.
func isSBRService(sbr *v1alpha1.ServiceBinding, obj runtime.Object) bool {
	services := servicesOrEmptyList(sbr.Spec.Services)
	for _, svc := range services {
		svcGVK := schema.GroupVersionKind{Group: svc.Group, Version: svc.Version, Kind: svc.Kind}
		if obj.GetObjectKind().GroupVersionKind() == svcGVK {
			return true
		}
	}
	return false
}

// isSBRApplication checks whether the given obj is an application in given sbr.
func isSBRApplication(
	restMapper meta.RESTMapper,
	app *v1alpha1.Application,
	gvk schema.GroupVersionKind,
	name string,
) (bool, error) {
	if app == nil {
		return false, nil
	}
	appGVR := schema.GroupVersionResource{
		Group:    app.Group,
		Version:  app.Version,
		Resource: app.Resource,
	}
	appGVK, err := restMapper.KindFor(appGVR)
	if err != nil {
		return false, err
	}

	isEqual := gvk == appGVK

	if len(app.Name) > 0 {
		isEqual = app.Name == name
	}

	return isEqual, nil
}

// isSecretOwnedBySBR checks whether the given obj is a secret owned by the given sbr.
func isSecretOwnedBySBR(obj metav1.Object, sbr *v1alpha1.ServiceBinding) bool {
	return sbr.GetNamespace() == obj.GetNamespace() && sbr.Status.Secret == obj.GetName()
}

// convertToSBR attempts to convert the given obj into a Service Binding.
func convertToSBR(obj map[string]interface{}) (*v1alpha1.ServiceBinding, error) {
	sbr := &v1alpha1.ServiceBinding{}
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
//
// This method is responsible for ingesting arbitrary Kubernetes resources (for example corev1.Secret
// or appsv1.Deployment) and lookup whether they are related to one or more existing Service Binding
// Request resources.
func (m *sbrRequestMapper) Map(obj handler.MapObject) []reconcile.Request {
	log := mapperLog.WithValues(
		"Object.Namespace", obj.Meta.GetNamespace(),
		"Object.Name", obj.Meta.GetName(),
	)

	namespacedNamesToReconcile := make(namespacedNameSet)

	if isServiceBinding(obj.Object) {
		requests := []reconcile.Request{
			{NamespacedName: convertToNamespacedName(obj.Meta)},
		}
		log.Debug("current resource is a SBR", "Requests", requests)
		return requests
	}

	// note(isutton): The client handles retries on the operator behalf, so only unrecoverable errors
	// are left.
	//
	// please see https://github.com/isutton/service-binding-operator/blob/e17445570bd3889bcf7499142350a3b81463c6be/vendor/k8s.io/client-go/rest/request.go#L723-L812
	sbrList, err := m.client.Resource(groupVersion).List(metav1.ListOptions{})
	if err != nil {
		log.Error(err, "listing SBRs")
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
			log.Debug("resource identified as a secret owned by the SBR")
			namespacedNamesToReconcile.add(namespacedName)
		} else {
			log.Trace("resource is not a secret owned by the SBR")
		}

		if isSBRService(sbr, obj.Object) {
			log.Debug("resource identified as service in SBR", "NamespacedName", namespacedName)
			namespacedNamesToReconcile.add(namespacedName)
		} else {
			log.Trace("resource is not a service declared by the SBR")
		}

		if ok, err := isSBRApplication(
			m.restMapper,
			sbr.Spec.Application,
			obj.Object.GetObjectKind().GroupVersionKind(),
			obj.Meta.GetName(),
		); err != nil {
			log.Error(err, "identifying resource as SBR application")
			continue ITEMS
		} else if !ok {
			log.Trace("resource is not an application declared by the SBR")
			continue ITEMS
		} else {
			log.Debug("resource identified as an application in SBR", "NamespacedName", namespacedName)
			namespacedNamesToReconcile.add(namespacedName)
		}
	}

	requests := convertToRequests(namespacedNamesToReconcile)
	if count := len(requests); count > 0 {
		log.Debug("found SBRs for resource", "Count", count, "Requests", requests)
	} else {
		log.Debug("no SBRs found for resource")
	}
	return requests
}
