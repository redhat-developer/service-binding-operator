package servicebindingrequest

import (
	"encoding/base64"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	path = map[string][][]string{
		"ConfigMap": {{"data"}},
		"Secret":    {{"data"}},
		"Route":     {{"spec", "host"}},
		"Service":   {{"spec", "clusterIP"}},
	}
)

// DetectBindableResources struct contains information about operator backed CR and
// list of expected GVRs to extract information from.
type DetectBindableResources struct {
	cr               *unstructured.Unstructured
	resourcesToCheck []schema.GroupVersionResource
	client           dynamic.Interface
	data             map[string]interface{}
}

// NewDetectBindableResources returns new instance
func NewDetectBindableResources(
	sbr *v1alpha1.ServiceBindingRequest,
	cr *unstructured.Unstructured,
	resources []schema.GroupVersionResource,
	client dynamic.Interface,
) *DetectBindableResources {
	b := new(DetectBindableResources)
	b.client = client
	b.cr = cr
	b.resourcesToCheck = resources
	b.data = make(map[string]interface{})
	return b
}

// GetOwnedResources returns list of subresources owned by operator backed CR
func (b DetectBindableResources) GetOwnedResources() ([]unstructured.Unstructured, error) {
	var subResources []unstructured.Unstructured
	for _, resource := range b.resourcesToCheck {
		lst, err := b.client.Resource(resource).List(v1.ListOptions{})
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

// GetBindableVariables extracts required key value information from provided GVRs subresources
func (b DetectBindableResources) GetBindableVariables() (map[string]interface{}, error) {
	ownedResources, err := b.GetOwnedResources()
	if err != nil {
		return b.data, err
	}
	for _, resource := range ownedResources {
		switch resource.GetKind() {
		// In case of ConfigMap we would read data field
		case "ConfigMap":
			for _, v := range path[resource.GetKind()] {
				d, exist, err := unstructured.NestedMap(resource.Object, v...)
				if err != nil {
					// skipping on error
					continue
				}
				if exist {
					for k, val := range d {
						b.data[k] = val
					}
				}
			}
		// In case of Secret we would decode and read data field
		case "Secret":
			for _, v := range path[resource.GetKind()] {
				d, exist, err := unstructured.NestedMap(resource.Object, v...)
				if err != nil {
					// skipping on error
					continue
				}
				if exist {
					for k, val := range d {
						value, _ := base64.StdEncoding.DecodeString(val.(string))
						b.data[k] = value
					}
				}
			}

			// In case of Route and Service we would extract information from respective path
		case "Route", "Service":
			for _, v := range path[resource.GetKind()] {
				d, exist, err := unstructured.NestedString(resource.Object, v...)
				if err != nil {
					continue
				}
				if exist {
					val := v[len(v)-1]
					b.data[val] = d
				}
			}
		}
	}
	return b.data, nil
}
