package servicebindingrequest

import (
	"context"

	"github.com/go-logr/logr"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Planner plans resources needed to bind a given backend service, using OperatorLifecycleManager
// standards and CustomResourceDefinitionDescription data to understand which attributes are needed.
type Planner struct {
	ctx    context.Context                 // request context
	client dynamic.Interface               // kubernetes dynamic api client
	sbr    *v1alpha1.ServiceBindingRequest // instantiated service binding request
	logger logr.Logger                     // logger instance
}

// Plan outcome, after executing planner.
type Plan struct {
	Ns             string                         // namespace name
	Name           string                         // plan name, same than ServiceBindingRequest
	CRDDescription *olmv1alpha1.CRDDescription    // custom resource definition description
	CR             *unstructured.Unstructured     // custom resource object
	SBR            v1alpha1.ServiceBindingRequest // service binding request
	Annotations    map[string]string              // annotations in the backing service CRD
}

// searchCR based on a CustomResourceDefinitionDescription and name, search for the object.
func (p *Planner) searchCR(kind string) (*unstructured.Unstructured, error) {
	bss := p.sbr.Spec.BackingServiceSelector
	gvk := schema.GroupVersionKind{Group: bss.Group, Version: bss.Version, Kind: bss.Kind}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	opts := metav1.GetOptions{}

	logger := p.logger.WithValues("CR.GVK", gvk.String(), "CR.GVR", gvr.String())
	logger.Info("Searching for CR instance...")

	cr, err := p.client.Resource(gvr).Namespace(p.sbr.GetNamespace()).Get(bss.ResourceRef, opts)

	if err != nil {
		logger.Error(err, "during reading CR")
		return nil, err
	}

	logger.WithValues("CR.Name", cr.GetName()).Info("Found target CR!")
	return cr, nil
}

// searchCRD based on a CustomResourceDefinitionDescription and name, search for the object.
func (p *Planner) searchCRD() (*unstructured.Unstructured, error) {
	bss := p.sbr.Spec.BackingServiceSelector
	gvk := schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition"}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	opts := metav1.GetOptions{}

	logger := p.logger.WithValues("CR.GVK", gvk.String(), "CR.GVR", gvr.String())
	logger.Info("Searching for CR instance...")

	crd, err := p.client.Resource(gvr).Namespace(p.sbr.GetNamespace()).Get(bss.Kind, opts)

	if err != nil {
		logger.Error(err, "during reading CR")
		return nil, err
	}

	logger.WithValues("CR.Name", crd.GetName()).Info("Found target CR!")
	return crd, nil
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

	// retrieve the CR based on kind, api-version and name
	crd, err := p.searchCRD()
	if err != nil {
		return nil, err
	}

	return &Plan{
		Ns:             p.sbr.GetNamespace(),
		Name:           p.sbr.GetName(),
		CRDDescription: crdDescription,
		CR:             cr,
		SBR:            *p.sbr,
		Annotations:    crd.GetAnnotations(),
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
		logger: logf.Log.WithName("plan"),
	}
}
