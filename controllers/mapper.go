package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

var (
	mapperLog = log.NewLog("mapper")
)

// sbrRequestMapper is the handler.Mapper interface implementation. It should influence the
// enqueue process considering the resources informed.
type sbrRequestMapper struct {
	client     dynamic.Interface
	typeLookup K8STypeLookup
}

// isServiceBinding checks whether the given obj is a Service Binding through GVK
// comparison.
func isServiceBinding(obj runtime.Object) bool {
	return obj.GetObjectKind().GroupVersionKind() == v1alpha1.GroupVersionKind
}

// convertToNamespacedName returns a NamespacedName with information extracted from given obj.
func convertToNamespacedName(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
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

	if isServiceBinding(obj.Object) {
		requests := []reconcile.Request{
			{NamespacedName: convertToNamespacedName(obj.Meta)},
		}
		log.Debug("current resource is a SBR", "Requests", requests)
		return requests
	}
	return []reconcile.Request{}

}

type Referable interface {
	GroupVersionResource() (*schema.GroupVersionResource, error)
	GroupVersionKind() (*schema.GroupVersionKind, error)
}

type K8STypeLookup interface {
	ResourceForReferable(obj Referable) (*schema.GroupVersionResource, error)
	ResourceForKind(gvk schema.GroupVersionKind) (*schema.GroupVersionResource, error)
	KindForResource(gvr schema.GroupVersionResource) (*schema.GroupVersionKind, error)
}

func (r *ServiceBindingReconciler) ResourceForReferable(obj Referable) (*schema.GroupVersionResource, error) {
	gvr, err := obj.GroupVersionResource()
	if err == nil {
		return gvr, nil
	}
	gvk, err := obj.GroupVersionKind()
	if err != nil {
		return nil, err
	}
	return r.ResourceForKind(*gvk)
}

func (r *ServiceBindingReconciler) ResourceForKind(gvk schema.GroupVersionKind) (*schema.GroupVersionResource, error) {
	mapping, err := r.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return &mapping.Resource, nil
}

func (r *ServiceBindingReconciler) KindForResource(gvr schema.GroupVersionResource) (*schema.GroupVersionKind, error) {
	gvk, err := r.restMapper.KindFor(gvr)
	return &gvk, err
}
