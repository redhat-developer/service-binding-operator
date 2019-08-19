package servicebindingrequest

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	Ns             string                      // namespace name
	Name           string                      // plan name, same than ServiceBindingRequest
	CRDDescription *olmv1alpha1.CRDDescription // custom resource definition description
	CR             *ustrv1.Unstructured        // custom resource object
}

// searchCRDDescription based on BackingServiceSelector instance, find a
// CustomResourceDefinitionDescription to return, otherwise creating a not-found error.
func (p *Planner) searchCRDDescription() (*olmv1alpha1.CRDDescription, error) {
	csvResourceName := "clusterserviceversions"
	ns := p.sbr.GetNamespace()
	gvr := schema.GroupVersionResource{
		Group:    olmv1alpha1.GroupName,
		Version:  olmv1alpha1.GroupVersion,
		Resource: csvResourceName,
	}

	logger := p.logger.WithValues(
		"CSV.Group", gvr.Group,
		"CSV.Version", gvr.Version,
		"CSV.Resource", gvr.Resource,
		"CSV.Namespace", ns,
	)
	logger.Info("Searching for ClusterServiceVersion...")

	// FIXME: usually the CSV resources are in the "openshift-operator-lifecycle-manager" namespace;
	uList, err := p.client.Resource(gvr).Namespace(ns).List(metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "during search for CSV")
		return nil, err
	}
	logger.WithValues("CSV.List", len(uList.Items)).Info("CSV resources found...")

	for _, u := range uList.Items {
		logger = logger.WithValues("CSV.Name", u.GetName(), "CSV.APIVersion", u.GetAPIVersion())

		logger.Info("Converting unstructured back to original resource type...")
		csv := olmv1alpha1.ClusterServiceVersion{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &csv)
		if err != nil {
			logger.Error(err, "during unstructured conversion to CSV list object")
			return nil, err
		}

		logger.Info("Searching for CRDDescription matching BackingServiceSelector...")
		crdDescription := p.extractCRDDescription(logger, &csv)
		if crdDescription != nil {
			return crdDescription, nil
		}
	}

	p.logger.Info("Warning: not able to find a CRD description!")
	return nil, errors.NewNotFound(olmv1alpha1.Resource(csvResourceName), "")
}

// extractCRDDescription auxiliary method to identify the CRD-Description object the
// BackingServiceSelector is looking for, otherwise returns nil.
func (p *Planner) extractCRDDescription(
	logger logr.Logger,
	csv *olmv1alpha1.ClusterServiceVersion,
) *olmv1alpha1.CRDDescription {
	bss := p.sbr.Spec.BackingServiceSelector
	logger = p.logger.WithValues(
		"BackingServiceSelector.Group", bss.Group,
		"BackingServiceSelector.Version", bss.Version,
		"BackingServiceSelector.Kind", bss.Kind,
	)

	for _, crd := range csv.Spec.CustomResourceDefinitions.Owned {
		logger = logger.WithValues(
			"CRDDescription.Name", crd.Name,
			"CRDDescription.Version", crd.Version,
			"CRDDescription.Kind", crd.Kind,
		)
		logger.Info("Inspecting CustomResourceDefinitionDescription object...")

		// checking for suffix since is expected to have object type as prefix
		if !strings.EqualFold(strings.ToLower(crd.Kind), strings.ToLower(bss.Kind)) {
			continue
		}
		if crd.Version != "" && strings.ToLower(bss.Version) != strings.ToLower(crd.Version) {
			continue
		}

		logger.Info("CRD matches BackingServiceSelector!")
		return &crd
	}

	return nil
}

// searchCR based on a CustomResourceDefinitionDescription and name, search for the object.
func (p *Planner) searchCR(kind string) (*ustrv1.Unstructured, error) {
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

	return &Plan{
		Ns:             p.sbr.GetNamespace(),
		Name:           p.sbr.GetName(),
		CRDDescription: crdDescription,
		CR:             cr,
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
