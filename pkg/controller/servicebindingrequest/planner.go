package servicebindingrequest

import (
	"context"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
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
	Ns             string                         // namespace name
	Name           string                         // plan name, same than ServiceBindingRequest
	CRDDescription *olmv1alpha1.CRDDescription    // custom resource definition description
	CR             *unstructured.Unstructured     // custom resource object
	SBR            v1alpha1.ServiceBindingRequest // service binding request
}

// searchCR based on a CustomResourceDefinitionDescription and name, search for the object.
func (p *Planner) searchCR(kind string) (*unstructured.Unstructured, error) {
	bss := p.sbr.Spec.BackingServiceSelector
	gvk := schema.GroupVersionKind{Group: bss.Group, Version: bss.Version, Kind: bss.Kind}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	opts := metav1.GetOptions{}

	log := p.logger.WithValues("CR.GVK", gvk.String(), "CR.GVR", gvr.String())
	log.Debug("Searching for CR instance...")

	cr, err := p.client.Resource(gvr).Namespace(p.sbr.GetNamespace()).Get(bss.ResourceRef, opts)
	if err != nil {
		log.Error(err, "during reading CR")
		return nil, err
	}

	log.Debug("Found target CR!", "CR.Name", cr.GetName())
	return cr, nil
}

// Plan by retrieving the necessary resources related to binding a service backend.
func (p *Planner) Plan() (*Plan, error) {
	bss := p.sbr.Spec.BackingServiceSelector
	gvk := schema.GroupVersionKind{Group: bss.Group, Version: bss.Version, Kind: bss.Kind}
	olm := NewOLM(p.client, p.sbr.GetNamespace())
	crdDescription, err := olm.SelectCRDByGVK(gvk)
	if err != nil {
		return nil, err
	}

	// retrieve the CR based on kind, api-version and name
	cr, err := p.searchCR(crdDescription.Kind)
	if err != nil {
		return nil, err
	}

	return &Plan{
		Ns:             p.sbr.GetNamespace(),
		Name:           p.sbr.GetName(),
		CRDDescription: crdDescription,
		CR:             cr,
		SBR:            *p.sbr,
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
