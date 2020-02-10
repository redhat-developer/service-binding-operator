package servicebindingrequest

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

var (
	plannerLog = log.NewLog("planner")
)

// Planner plans resources needed to bind a given backend service, using OperatorLifecycleManager
// standards and CustomResourceDefinitionDescription data to understand which attributes are needed.
type Planner struct {
	ctx    context.Context                 // request context
	client dynamic.Interface               // kubernetes dynamic api client
	sbr    *v1alpha1.ServiceBindingRequest // instantiated service binding request
	logger *log.Log                        // logger instance
}

// Plan outcome, after executing planner.
type Plan struct {
	Ns               string                         // namespace name
	Name             string                         // plan name, same than ServiceBindingRequest
	SBR              v1alpha1.ServiceBindingRequest // service binding request
	RelatedResources RelatedResources               // CR and CRDDescription pairs SBR related
}

// searchCR based on a CustomResourceDefinitionDescription and name, search for the object.
func (p *Planner) searchCR(namespace string, selector v1alpha1.BackingServiceSelector) (*unstructured.Unstructured, error) {
	// gvr is the plural guessed resource for the given selector
	gvk := schema.GroupVersionKind{
		Group:   selector.Group,
		Version: selector.Version,
		Kind:    selector.Kind,
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	// delegate the search selector's namespaced resource client
	return p.client.Resource(gvr).Namespace(namespace).Get(selector.ResourceRef, metav1.GetOptions{})
}

// CRDGVR is the plural GVR for Kubernetes CRDs.
var CRDGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1beta1",
	Resource: "customresourcedefinitions",
}

// searchCRD returns the CRD related to the gvk.
func (p *Planner) searchCRD(gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	// gvr is the plural guessed resource for the given GVK
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	// crdName is the string'fied GroupResource, e.g. "deployments.apps"
	crdName := gvr.GroupResource().String()
	// delegate the search to the CustomResourceDefinition resource client
	return p.client.Resource(CRDGVR).Get(crdName, metav1.GetOptions{})
}

// Plan by retrieving the necessary resources related to binding a service backend.
func (p *Planner) Plan() (*Plan, error) {
	ns := p.sbr.GetNamespace()
	selectors := append([]v1alpha1.BackingServiceSelector{}, p.sbr.Spec.BackingServiceSelectors...)
	selector := p.sbr.Spec.BackingServiceSelector
	if len(selector.ResourceRef) > 0 {
		selectors = append(selectors, selector)
	}

	relatedResources := make([]*RelatedResource, 0)
	for _, s := range selectors {
		bssGVK := schema.GroupVersionKind{Kind: s.Kind, Version: s.Version, Group: s.Group}

		// resolve the CRD using the service's GVK
		crd, err := p.searchCRD(bssGVK)
		if err != nil {
			return nil, err
		}
		p.logger.Debug("Resolved CRD", "CRD", crd)

		// resolve the CRDDescription based on the service's GVK and the resolved CRD
		olm := NewOLM(p.client, ns)
		crdDescription, err := olm.SelectCRDByGVK(bssGVK, crd)
		if err != nil {
			return nil, err
		}
		p.logger.Debug("Resolved CRDDescription", "CRDDescription", crdDescription)

		cr, err := p.searchCR(ns, s)
		if err != nil {
			return nil, err
		}

		r := &RelatedResource{
			CRDDescription: crdDescription,
			CR:             cr,
		}
		relatedResources = append(relatedResources, r)
		p.logger.Debug("Resolved related resource", "RelatedResource", r)
	}

	return &Plan{
		Name:             p.sbr.GetName(),
		Ns:               ns,
		RelatedResources: relatedResources,
		SBR:              *p.sbr,
	}, nil
}

// NewPlanner instantiate Planner type.
func NewPlanner(
	ctx context.Context,
	client dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
) *Planner {
	return &Planner{
		ctx:    ctx,
		client: client,
		sbr:    sbr,
		logger: plannerLog,
	}
}
