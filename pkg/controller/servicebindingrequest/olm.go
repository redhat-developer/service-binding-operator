package servicebindingrequest

import (
	"fmt"
	"strings"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// OLM represents the actions this operator needs to take upon Operator-Lifecycle-Manager resources,
// like ClusterServiceVersions (CSV) and CRDDescriptions.
type OLM struct {
	client dynamic.Interface // kubernetes dynamic client
	ns     string            // namespace
	logger *log.Log          // logger instance
}

const (
	csvResource = "clusterserviceversions"
)

var (
	olmLog = log.NewLog("olm")
)

// listCSVs simple list to all CSV objects in the cluster.
func (o *OLM) listCSVs() ([]unstructured.Unstructured, error) {
	log := o.logger
	gvr := olmv1alpha1.SchemeGroupVersion.WithResource(csvResource)
	resourceClient := o.client.Resource(gvr).Namespace(o.ns)
	csvs, err := resourceClient.List(metav1.ListOptions{})
	if err != nil {
		log.Error(err, "during listing CSV objects from cluster")
		return nil, err
	}
	return csvs.Items, nil
}

// extractOwnedCRDs from a list of CSV objects.
func (o *OLM) extractOwnedCRDs(
	csvs []unstructured.Unstructured,
) ([]unstructured.Unstructured, error) {
	crds := []unstructured.Unstructured{}
	for _, csv := range csvs {
		ownedPath := []string{"spec", "customresourcedefinitions", "owned"}
		log := o.logger.WithValues("OwnedPath", ownedPath, "CSV.Name", csv.GetName())

		ownedCRDs, exists, err := unstructured.NestedSlice(csv.Object, ownedPath...)
		if err != nil {
			log.Error(err, "on extracting nested slice")
			return nil, err
		}
		if !exists {
			continue
		}

		for _, crd := range ownedCRDs {
			data := crd.(map[string]interface{})
			crds = append(crds, unstructured.Unstructured{Object: data})
		}
	}

	return crds, nil
}

// ListCSVOwnedCRDs return a unstructured list of CRD objects from "owned" section in CSVs.
func (o *OLM) ListCSVOwnedCRDs() ([]unstructured.Unstructured, error) {
	log := o.logger
	csvs, err := o.listCSVs()
	if err != nil {
		log.Error(err, "on listting CSVs")
		return nil, err
	}
	return o.extractOwnedCRDs(csvs)
}

// eachCRDDescriptionFn function to be called against each CRDDescription in a slice.
type eachCRDDescriptionFn func(crdDescription *olmv1alpha1.CRDDescription)

// loopCRDDescriptions takes a list of CRDDescriptions (extracted from "owned") and converts to a
// actual type instance, before calling out for informed function. This method can return error in
// case of issues to convert unstructured into CRDDescription.
func (o *OLM) loopCRDDescriptions(
	crdDescriptions []unstructured.Unstructured,
	fn eachCRDDescriptionFn,
) error {
	for _, u := range crdDescriptions {
		log := o.logger
		crdDescription := &olmv1alpha1.CRDDescription{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, crdDescription)
		if err != nil {
			log.Error(err, "on converting from unstructured to CRD")
			return err
		}
		log.Debug("Inspecting CRDDescription...", "CRDDescription", crdDescription)
		if crdDescription.Name == "" {
			log.Debug("Skipping empty CRDDescription!")
			continue
		}
		fn(crdDescription)
	}
	return nil
}

// SelectCRDByGVK return a single CRD based on a given GVK.
func (o *OLM) SelectCRDByGVK(gvk schema.GroupVersionKind) (*olmv1alpha1.CRDDescription, error) {
	log := o.logger.WithValues("Selector.GVK", gvk)
	ownedCRDs, err := o.ListCSVOwnedCRDs()
	if err != nil {
		log.Error(err, "on listing owned CRDs")
		return nil, err
	}

	crdDescriptions := []*olmv1alpha1.CRDDescription{}
	err = o.loopCRDDescriptions(ownedCRDs, func(crdDescription *olmv1alpha1.CRDDescription) {
		log = o.logger.WithValues(
			"CRDDescription.Name", crdDescription.Name,
			"CRDDescription.Version", crdDescription.Version,
			"CRDDescription.Kind", crdDescription.Kind,
		)
		log.Debug("Inspecting CRDDescription object...")
		// checking for suffix since is expected to have object type as prefix
		if !strings.EqualFold(strings.ToLower(crdDescription.Kind), strings.ToLower(gvk.Kind)) {
			return
		}
		// matching resource version, unless when not informed
		if crdDescription.Version != "" &&
			!strings.EqualFold(strings.ToLower(gvk.Version), strings.ToLower(crdDescription.Version)) {
			return
		}
		log.Debug("CRDDescription object matches selector!")
		crdDescriptions = append(crdDescriptions, crdDescription)
	})
	if err != nil {
		return nil, err
	}

	if len(crdDescriptions) == 0 {
		log.Debug("No CRD could be found for GVK.")
		return nil, fmt.Errorf("no crd could be found for gvk")
	}
	return crdDescriptions[0], nil
}

// extractGVKs loop owned objects and extract the GVK information from them.
func (o *OLM) extractGVKs(
	crdDescriptions []unstructured.Unstructured,
) ([]schema.GroupVersionKind, error) {
	log := o.logger
	gvks := []schema.GroupVersionKind{}
	err := o.loopCRDDescriptions(crdDescriptions, func(crdDescription *olmv1alpha1.CRDDescription) {
		log.Debug("Extracting GVK from CRDDescription", "CRDDescription.Name", crdDescription.Name)
		_, gv := schema.ParseResourceArg(crdDescription.Name)
		gvks = append(gvks, schema.GroupVersionKind{
			Group:   gv.Group,
			Version: crdDescription.Version,
			Kind:    crdDescription.Kind,
		})
	})
	if err != nil {
		return []schema.GroupVersionKind{}, err
	}
	return gvks, nil
}

// ListCSVOwnedCRDsAsGVKs return the list of owned CRDs from all CSV objects as a list of GVKs.
func (o *OLM) ListCSVOwnedCRDsAsGVKs() ([]schema.GroupVersionKind, error) {
	log := o.logger
	ownedCRDs, err := o.ListCSVOwnedCRDs()
	if err != nil {
		log.Error(err, "on listting CSVs")
		return nil, err
	}
	return o.extractGVKs(ownedCRDs)
}

// ListGVKsFromCSVNamespacedName return the list of owned GVKs for a given CSV namespaced named.
func (o *OLM) ListGVKsFromCSVNamespacedName(
	namespacedName types.NamespacedName,
) ([]schema.GroupVersionKind, error) {
	log := o.logger.WithValues("CSV.NamespacedName", namespacedName)
	log.Debug("Reading CSV to extract GVKs...")
	gvr := olmv1alpha1.SchemeGroupVersion.WithResource(csvResource)
	resourceClient := o.client.Resource(gvr).Namespace(namespacedName.Namespace)
	u, err := resourceClient.Get(namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		// the CSV might have disappeared between discovery and check, so not found is not an error
		if errors.IsNotFound(err) {
			return []schema.GroupVersionKind{}, nil
		}
		log.Error(err, "on reading CSV object")
		return []schema.GroupVersionKind{}, err
	}

	unstructuredCSV := *u
	csvs := []unstructured.Unstructured{unstructuredCSV}

	ownedCRDs, err := o.extractOwnedCRDs(csvs)
	if err != nil {
		log.Error(err, "on extracting owned CRDs")
		return []schema.GroupVersionKind{}, err
	}

	return o.extractGVKs(ownedCRDs)
}

// NewOLM instantiate a new OLM.
func NewOLM(client dynamic.Interface, ns string) *OLM {
	return &OLM{
		client: client,
		ns:     ns,
		logger: olmLog,
	}
}
