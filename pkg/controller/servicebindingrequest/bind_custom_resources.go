package servicebindingrequest

import (
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	path = map[string][]string{
		"ConfigMap": {"data"},
		"Secret": {"data"},
		"Route": {"spec", "host"},
	}
)

type BindNonBindableResources struct {
	sbr              *v1alpha1.ServiceBindingRequest
	cr               *unstructured.Unstructured
	resourcesToCheck []schema.GroupVersionResource
	client           dynamic.Interface
	data             map[string]interface{}
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

func (b BindNonBindableResources) GetOwnedResources() ([]unstructured.Unstructured, error) {
	var subResources []unstructured.Unstructured
	for _, resource := range b.resourcesToCheck {
		lst, err := b.client.Resource(resource).Namespace("test").List(v1.ListOptions{
			ResourceVersion: resource.GroupResource().String(),
		})
		if err != nil {
			return subResources, err
		}
		for _, item := range lst.Items {
			owners := item.GetOwnerReferences()
			for _, owner := range owners {
				if owner.UID == b.cr.GetUID() {
					subResources = append(subResources, item)
				}
			}
		}
	}
	return subResources, nil
}

func (b BindNonBindableResources) GetBindableVariables() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	ownedResources, err := b.GetOwnedResources()
	if err != nil {
		return data, err
	}
	for _, resource := range ownedResources {
		switch resource.GetKind() {
		case "ConfigMap", "Secret":
			d, exist, err := unstructured.NestedMap(resource.Object, path["ConfigMap"]...)
			if err != nil {
				continue
			}
			if exist {
				for k, v := range d {
					data[k] = v
				}
			}
			break
		case "Route":
			d, exist, err := unstructured.NestedString(resource.Object, path["Route"]...)
			if err != nil {
				continue
			}
			if exist {
				val := path["Route"][len(path["Route"]) - 1]
				data[val] = d
			}
			break
		}
	}
	return data, nil
}
