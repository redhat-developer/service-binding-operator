package servicebindingrequest

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	CRD            *ustrv1.Unstructured        // custom resource definition
}

const (
	connectsToLabel = "connects-to"
)

// extractConnectTo inspect ServiceBindingRequest to extract connects-to value.
func (p *Planner) extractConnectsTo() string {
	value, exists := p.sbr.Spec.ApplicationSelector.MatchLabels[connectsToLabel]
	if !exists {
		return ""
	}
	return value
}

// searchCRDDescription based on BackingServiceSelector instance, find a CustomResourceDefinitionDescription
// to return, otherwise creating a not-found error.
func (p *Planner) searchCRDDescription() (*olmv1alpha1.CRDDescription, error) {
	var backingServiceSelector = p.sbr.Spec.BackingServiceSelector
	var logger = p.logger.WithValues(
		"BackingServiceSelector.ResourceKind", backingServiceSelector.ResourceKind,
		"BackingServiceSelector.ResourceVersion", backingServiceSelector.ResourceVersion,
	)
	var err error

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

			if !strings.HasSuffix(crd.Name, backingServiceSelector.ResourceKind) {
				continue
			}
			if crd.Version != "" && backingServiceSelector.ResourceVersion != crd.Version {
				continue
			}

			logger.Info("CRD matches backing-selector!")
			return &crd, nil
		}
	}

	p.logger.Info("Warning: not able to find a CRD description!")
	return nil, errors.NewNotFound(extv1beta1.Resource("CustomResourceDefinition"), "")
}

// searchCRD based on a CustomResourceDefinitionDescription and name, search for the object.
func (p *Planner) searchCRD(
	kind, name string,
) (*ustrv1.Unstructured, error) {
	var backingServiceSelector = p.sbr.Spec.BackingServiceSelector

	apiVersion := fmt.Sprintf("%s/%s", backingServiceSelector.ResourceKind, backingServiceSelector.ResourceVersion)
	obj := map[string]interface{}{"kind": kind, "apiVersion": apiVersion}
	objList := &ustrv1.UnstructuredList{Object: obj}

	logger := p.logger.WithValues(
		"CustomResourceDefinition.Kind", name,
		"CustomResourceDefinition.Name", name,
		"CRDDescription.APIVersion", apiVersion,
	)
	logger.Info("Searching for CRD instance...")

	err := p.client.List(p.ctx, &client.ListOptions{Namespace: p.ns}, objList)
	if err != nil {
		return nil, err
	}

	// TODO: find a way to load the object directory, without having to loop a list;
	for _, item := range objList.Items {
		if name == item.GetName() {
			logger.Info("CustomResourceDefintion found!")
			return &item, nil
		}
	}

	logger.Info("Warning: not able to find the CustomResourceDefinition!")
	return nil, errors.NewNotFound(apiextv1beta1.Resource("CustomResourceDefinition"), name)
}

// Plan by retrieving the necessary resources related to binding a service backend.
func (p *Planner) Plan() (*Plan, error) {
	// extracting label conects-to, it shows the CRD name we are looking for
	connectsToValue := p.extractConnectsTo()
	if connectsToValue == "" {
		return nil, fmt.Errorf("unable to find label '%s' in service-binding-request '%s'",
			connectsToValue, p.sbr.GetName())
	}

	// find the CRD description object
	crdDescription, err := p.searchCRDDescription()
	if err != nil {
		return nil, err
	}

	// retrieve the CRD based on kind, api-version and name
	crd, err := p.searchCRD(crdDescription.Kind, connectsToValue)
	if err != nil {
		return nil, err
	}

	return &Plan{Ns: p.ns, Name: p.sbr.GetName(), CRDDescription: crdDescription, CRD: crd}, nil
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
