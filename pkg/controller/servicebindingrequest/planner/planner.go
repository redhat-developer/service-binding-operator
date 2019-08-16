package planner

import (
	"context"
	"fmt"
	"strings"

	"github.com/redhat-developer/service-binding-operator/pkg/resourcepoll"

	"github.com/go-logr/logr"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Planner plans resources needed to bind a given backend service, using OperatorLifecycleManager
// standards and CustomResourceDefinitionDescription data to understand which attributes are needed.
type Planner struct {
	ctx    context.Context                 // request context
	client client.Client                   // Kubernetes API client
	ns     string                          // namespace name
	sbr    *v1alpha1.ServiceBindingRequest // instantiated service binding request
	logger logr.Logger                     // logger instance
}

// Plan outcome, after executing planner.
type Plan struct {
	Ns             string                      // namespace name
	Name           string                      // plan name, same than ServiceBindingRequest
	CRDDescription *olmv1alpha1.CRDDescription // custom resource definition description
	CR             *ustrv1.Unstructured        // custom resource object
}

// searchCRDDescription based on BackingServiceSelector instance, find a CustomResourceDefinitionDescription
// to return, otherwise creating a not-found error.
func (p *Planner) searchCRDDescription() (*olmv1alpha1.CRDDescription, error) {
	var resourceKind = strings.ToLower(
		fmt.Sprintf(".%s", p.sbr.Spec.BackingServiceSelector.ResourceKind))
	var resourceVersion = strings.ToLower(p.sbr.Spec.BackingServiceSelector.ResourceVersion)
	var err error

	logger := p.logger.WithValues(
		"BackingServiceSelector.ResourceKind", resourceKind,
		"BackingServiceSelector.ResourceVersion", resourceVersion)
	logger.Info("Looking for a CSV based on backing-selector")
	csvList := &olmv1alpha1.ClusterServiceVersionList{}

	// list of cluster service version in the namespace matching backing-selector
	if err = p.client.List(p.ctx, &client.ListOptions{Namespace: p.ns}, csvList); err != nil {
		return nil, err
	}

	for _, csv := range csvList.Items {
		logger = logger.WithValues("ClusterServiceVersion.Name", csv.Name)
		logger.Info("Inspecting CSV...")

		for _, crd := range csv.Spec.CustomResourceDefinitions.Owned {
			logger = logger.WithValues(
				"CRDDescription.Name", crd.Name,
				"CRDDescription.Version", crd.Version,
				"CRDDescription.Kind", crd.Kind,
			)
			logger.Info("Inspecting CustomResourceDefinitionDescription object...")

			// checking for suffix since is expected to have object type as prefix
			if !strings.HasSuffix(strings.ToLower(crd.Name), resourceKind) {
				continue
			}
			if crd.Version != "" && resourceVersion != strings.ToLower(crd.Version) {
				continue
			}

			logger.Info("CRD matches backing-selector!")
			return &crd, nil
		}
	}

	p.logger.Info("Warning: not able to find a CRD description!")
	return nil, errors.NewNotFound(appsv1.Resource("CustomResourceDefinition"), "")
}

// searchCR based on a CustomResourceDefinitionDescription and name, search for the object.
func (p *Planner) searchCR(kind string) (*ustrv1.Unstructured, error) {
	var resourceRef = p.sbr.Spec.BackingServiceSelector.ResourceRef
	var apiVersion = fmt.Sprintf("%s/%s",
		p.sbr.Spec.BackingServiceSelector.ResourceKind,
		p.sbr.Spec.BackingServiceSelector.ResourceVersion)
	var err error

	p.logger.WithValues("CR.Name", resourceRef, "CR.Kind", kind, "CR.APIVersion", apiVersion).
		Info("Searching for CR instance...")

	cr := ustrv1.Unstructured{Object: map[string]interface{}{
		"kind":       kind,
		"apiVersion": apiVersion,
	}}
	namespacedName := types.NamespacedName{Namespace: p.ns, Name: resourceRef}

	err = resourcepoll.WaitUntilResourceFound(p.client, namespacedName, &cr)
	if err != nil {
		return nil, err
	}
	if err = p.client.Get(p.ctx, namespacedName, &cr); err != nil {
		return nil, err
	}

	return &cr, nil
}

// Plan by retrieving the necessary resources related to binding a service backend.
func (p *Planner) Plan() (*Plan, error) {
	// find the CRD description object
	crdDescription, err := p.searchCRDDescription()
	if err != nil {
		return nil, err
	}

	// retrieve the CR based on kind, api-version and name
	cr, err := p.searchCR(crdDescription.Kind)
	if err != nil {
		return nil, err
	}

	return &Plan{Ns: p.ns, Name: p.sbr.GetName(), CRDDescription: crdDescription, CR: cr}, nil
}

// NewPlanner instantiate Planner type.
func NewPlanner(
	ctx context.Context, client client.Client, ns string, sbr *v1alpha1.ServiceBindingRequest,
) *Planner {
	return &Planner{
		ctx:    ctx,
		client: client,
		ns:     ns,
		sbr:    sbr,
		logger: logf.Log.WithName("plan"),
	}
}
