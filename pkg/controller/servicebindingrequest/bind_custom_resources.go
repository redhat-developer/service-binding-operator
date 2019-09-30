package servicebindingrequest

import (
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type BindNonBindableResources struct {
	sbr *v1alpha1.ServiceBindingRequest
	cr *unstructured.Unstructured
	resourcesToCheck []schema.GroupVersionResource
	client dynamic.Interface
}

func NewBindNonBindable(
	sbr *v1alpha1.ServiceBindingRequest,
	cr *unstructured.Unstructured,
	resources []schema.GroupVersionResource,
	client dynamic.Interface,
	) *BindNonBindableResources {
	b := new(BindNonBindableResources)
	b.client = client
	b.cr = cr
	b.resourcesToCheck = resources
	b.sbr = sbr
	return b
}

func (b BindNonBindableResources) GetOwnedResources() ([]unstructured.Unstructured ,error) {
	var subResources []unstructured.Unstructured
	for _, resource := range b.resourcesToCheck {
		lst, err := b.client.Resource(resource).Namespace("test").List(v1.ListOptions{})
		if err != nil {
			return subResources, err
		}
		for _, item := range lst.Items {
			owners := item.GetOwnerReferences()
			for _, owner := range owners {
				if owner.UID == b.cr.GetUID() {
					uItem, err := runtime.DefaultUnstructuredConverter.ToUnstructured(item)
					if err != nil {
						return subResources, nil
					}
					subResources = append(subResources, unstructured.Unstructured{Object:uItem})
				}
			}
		}
	}
	return subResources, nil
}
