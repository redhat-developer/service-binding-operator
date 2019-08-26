package servicebindingrequest

import (
	"strings"

	"github.com/go-logr/logr"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// OLM represents the actions this operator needs to take upon Operator-Lifecycle-Manager resources,
// Like ClusterServiceVersions (CSV) and CRD-Descriptions.
type OLM struct {
	client dynamic.Interface // kubernetes dynamic client
	ns     string            // namespace
	logger logr.Logger       // logger instance
}

const (
	csvResource = "clusterserviceversions"
)

// listCSVs simple list to all CSV objects in the cluster.
func (o *OLM) listCSVs() ([]unstructured.Unstructured, error) {
	gvr := olmv1alpha1.SchemeGroupVersion.WithResource(csvResource)
	resourceClient := o.client.Resource(gvr).Namespace(o.ns)
	csvs, err := resourceClient.List(metav1.ListOptions{})
	if err != nil {
		o.logger.Error(err, "during listing CSV objects from cluster")
		return nil, err
	}
	return csvs.Items, nil
}

// ListCSVOwnedCRDs return a unstructured list of CRD objects from "owned" section in CSVs.
func (o *OLM) ListCSVOwnedCRDs() ([]*unstructured.Unstructured, error) {
	crds := []*unstructured.Unstructured{}
	csvs, err := o.listCSVs()
	if err != nil {
		return nil, err
	}

	// TODO: add right logger entries like in Planner;
	for _, csv := range csvs {
		ownedPath := []string{"spec", "customresourcedefinitions", "owned"}
		ownedCRDs, exists, err := unstructured.NestedSlice(csv.Object, ownedPath...)
		if err != nil {
			return nil, err
		}
		if !exists {
			continue
		}
		for _, crd := range ownedCRDs {
			data := crd.(map[string]interface{})
			crds = append(crds, &unstructured.Unstructured{Object: data})
		}
	}

	return crds, nil
}

type eachOwnedCRDFn func(crd *olmv1alpha1.CRDDescription)

// loopCSVOwnedCRDs takes a function as parameter and excute this with every CRD object.
func (o *OLM) loopCSVOwnedCRDs(fn eachOwnedCRDFn) error {
	crds, err := o.ListCSVOwnedCRDs()
	if err != nil {
		o.logger.Error(err, "during list CSV owned CRDs")
		return err
	}

	for _, u := range crds {
		crd := &olmv1alpha1.CRDDescription{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, crd)
		if err != nil {
			o.logger.Error(err, "on converting from unstructured to CRD")
			return err
		}
		fn(crd)
	}

	return nil
}

// SelectCRDByGVK return a single CRD based on a given GVK.
func (o *OLM) SelectCRDByGVK(gvk schema.GroupVersionKind) (*olmv1alpha1.CRDDescription, error) {
	logger := o.logger.WithValues("Selector.GVK", gvk)
	crds := []*olmv1alpha1.CRDDescription{}

	err := o.loopCSVOwnedCRDs(func(crd *olmv1alpha1.CRDDescription) {
		logger = logger.WithValues(
			"CRDDescription.Name", crd.Name,
			"CRDDescription.Version", crd.Version,
			"CRDDescription.Kind", crd.Kind,
		)
		logger.Info("Inspecting CRDDescription object...")
		// checking for suffix since is expected to have object type as prefix
		if !strings.EqualFold(strings.ToLower(crd.Kind), strings.ToLower(gvk.Kind)) {
			return
		}
		if crd.Version != "" && strings.ToLower(gvk.Version) != strings.ToLower(crd.Version) {
			return
		}
		logger.Info("CRDDescription object matches selector!")
		crds = append(crds, crd)
	})
	if err != nil {
		return nil, err
	}

	if len(crds) == 0 {
		logger.Info("No CRD could be found for GVK.")
		return nil, nil
	}
	return crds[0], nil
}

// ListCSVOwnedCRDsAsGVKs return the list of owned CRDs from all CSV objects as a list of GVKs.
func (o *OLM) ListCSVOwnedCRDsAsGVKs() ([]schema.GroupVersionKind, error) {
	gvks := []schema.GroupVersionKind{}
	err := o.loopCSVOwnedCRDs(func(crd *olmv1alpha1.CRDDescription) {
		_, gv := schema.ParseResourceArg(crd.Name)
		gvks = append(gvks, schema.GroupVersionKind{
			Group:   gv.Group,
			Version: crd.Version,
			Kind:    crd.Kind,
		})
	})
	if err != nil {
		return []schema.GroupVersionKind{}, err
	}
	return gvks, nil
}

// NewOLM instantiate a new OLM.
func NewOLM(client dynamic.Interface, ns string) *OLM {
	return &OLM{
		client: client,
		ns:     ns,
		logger: logf.Log.WithName("olm"),
	}
}
