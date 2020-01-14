package servicebindingrequest

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// GetCRs returns a slice of unstructured CRs contained in the internal related resources collection.
func (p *Plan) GetCRs() []*unstructured.Unstructured {
	// return p.RelatedResources.GetCRs()
	panic("implement me")
}

// GetRelatedResources returns the collection of related resources enumerated in the plan.
func (p *Plan) GetRelatedResources() RelatedResources {
	//return p.RelatedResources
	panic("implement me")
}
