package servicebindingrequest

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	v1 "github.com/openshift/custom-resource-status/conditions/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/conditions"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

var (
	plannerLog                 = log.NewLog("planner")
	errBackingServiceNamespace = errors.New("backing Service Namespace is unspecified")
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
func (p *Planner) searchCR(selector v1alpha1.BackingServiceSelector) (*unstructured.Unstructured, error) {
	// gvr is the plural guessed resource for the given selector
	gvk := schema.GroupVersionKind{
		Group:   selector.Group,
		Version: selector.Version,
		Kind:    selector.Kind,
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)

	if selector.Namespace == nil {
		return nil, errBackingServiceNamespace
	}

	// delegate the search selector's namespaced resource client
	return p.client.Resource(gvr).Namespace(*selector.Namespace).Get(selector.ResourceRef, metav1.GetOptions{})
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

var EmptyBackingServiceSelectorsErr = errors.New("backing service selectors are empty")

// Plan by retrieving the necessary resources related to binding a service backend.
func (p *Planner) Plan() (*Plan, error) {
	ns := p.sbr.GetNamespace()

	var emptyApplication v1alpha1.ApplicationSelector
	if p.sbr.Spec.ApplicationSelector == emptyApplication {
		v1.SetStatusCondition(&p.sbr.Status.Conditions, v1.Condition{
			Type:    conditions.BindingReady,
			Status:  corev1.ConditionFalse,
			Reason:  "EmptyApplicationSelector",
			Message: EmptyApplicationSelectorErr.Error(),
		})

		p.sbr.Status.Applications = nil
	}

	var selectors []v1alpha1.BackingServiceSelector
	if p.sbr.Spec.BackingServiceSelector != nil {
		selectors = append(selectors, *p.sbr.Spec.BackingServiceSelector)
	}
	if p.sbr.Spec.BackingServiceSelectors != nil {
		selectors = append(selectors, *p.sbr.Spec.BackingServiceSelectors...)
	}

	if len(selectors) == 0 {
		return nil, EmptyBackingServiceSelectorsErr
	}

	relatedResources := make([]*RelatedResource, 0)
	for _, s := range selectors {

		var crdDescription *olmv1alpha1.CRDDescription

		if s.Namespace == nil {
			s.Namespace = &ns
		}

		bssGVK := schema.GroupVersionKind{Kind: s.Kind, Version: s.Version, Group: s.Group}

		// Start with looking up if the resource exists
		// If yes, errors during lookups of the CRD and
		// the CRD could be ignored.
		cr, err := p.searchCR(s)
		if err != nil {
			return nil, err
		}

		// resolve the CRD using the service's GVK
		crd, err := p.searchCRD(bssGVK)
		if err != nil {
			// expected this to work, but didn't
			// if k8sError.IsNotFound(err) {...}
			p.logger.Error(err, "Probably not a CRD")

		} else {

			p.logger.Debug("Resolved CRD", "CRD", crd)

			olm := NewOLM(p.client, ns)

			// Parse annotations from the OLM descriptors or the CRD
			crdDescription, err = olm.SelectCRDByGVK(bssGVK, crd)
			if err != nil {
				p.logger.Error(err, "Probably not an OLM operator")
			}
			p.logger.Debug("Tentatively resolved CRDDescription", "CRDDescription", crdDescription)
		}

		// Parse ( and override ) annotations from the CR or kubernetes object
		if crdDescription == nil {
			crdDescription = &olmv1alpha1.CRDDescription{}
		}
		err = buildCRDDescriptionFromCR(cr, crdDescription)
		if err != nil {
			return nil, err
		}

		p.logger.Debug("Computed CRDDescription", "CRDDescription", crdDescription)
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
