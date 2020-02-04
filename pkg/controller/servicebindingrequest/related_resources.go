package servicebindingrequest

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// RelatedResource represents a SBR related resource, composed by its CR and CRDDescription.
type RelatedResource struct {
	CRDDescription *v1alpha1.CRDDescription
	CR             *unstructured.Unstructured
}

// RelatedResources contains a collection of SBR related resources.
type RelatedResources []*RelatedResource

// GetCRs returns a slice of unstructured CRs contained in the collection.
func (rr RelatedResources) GetCRs() []*unstructured.Unstructured {
	var crs []*unstructured.Unstructured
	for _, r := range rr {
		crs = append(crs, r.CR)
	}
	return crs
}
