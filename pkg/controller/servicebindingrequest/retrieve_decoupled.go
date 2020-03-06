package servicebindingrequest

import (
	"fmt"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Get returns the data read from related resources (see ReadBindableResourcesData and
// ReadCRDDescriptionData).
func (r *Retriever) Get() (map[string][]byte, error) {
	// interpolating custom environment
	envParser := NewCustomEnvParser(r.plan.SBR.Spec.CustomEnvVar, r.cache)
	customVars, err := envParser.Parse()
	if err != nil {
		return nil, err
	}

	// convert values to a map[string][]byte
	result := make(map[string][]byte)
	for k, v := range customVars {
		result[k] = []byte(v.(string))
	}

	// include extracted data from related resources
	for k, v := range r.data {
		result[k] = v
	}
	return result, nil
}

// ReadBindableResourcesData reads all related resources of a given sbr
func (r *Retriever) ReadBindableResourcesData(
	sbr *v1alpha1.ServiceBindingRequest,
	crs []*unstructured.Unstructured,
) error {
	r.logger.Info("Detecting extra resources for binding...")
	for _, cr := range crs {
		b := NewDetectBindableResources(sbr, cr, []schema.GroupVersionResource{
			{Group: "", Version: "v1", Resource: "configmaps"},
			{Group: "", Version: "v1", Resource: "services"},
			{Group: "route.openshift.io", Version: "v1", Resource: "routes"},
		}, r.client)

		vals, err := b.GetBindableVariables()
		if err != nil {
			return err
		}
		for k, v := range vals {
			r.storeInto(cr, k, []byte(fmt.Sprintf("%v", v)))
		}
	}

	return nil
}

func (r *Retriever) storeInto(cr *unstructured.Unstructured, key string, value []byte) {
	r.store(cr, key, value)
}

func (r *Retriever) copyFrom(u *unstructured.Unstructured, path string, fieldPath string, descriptors []string) error {
	if err := r.read(u, path, fieldPath, descriptors); err != nil {
		return err
	}
	return nil
}

// ReadCRDDescriptionData reads data related to given crdDescription
func (r *Retriever) ReadCRDDescriptionData(u *unstructured.Unstructured, crdDescription *olmv1alpha1.CRDDescription) error {
	r.logger.Info("Looking for spec-descriptors in 'spec'...")
	for _, specDescriptor := range crdDescription.SpecDescriptors {
		if err := r.copyFrom(u, "spec", specDescriptor.Path, specDescriptor.XDescriptors); err != nil {
			return err
		}
	}

	r.logger.Info("Looking for status-descriptors in 'status'...")
	for _, statusDescriptor := range crdDescription.StatusDescriptors {
		if err := r.copyFrom(u, "status", statusDescriptor.Path, statusDescriptor.XDescriptors); err != nil {
			return err
		}
	}

	return nil
}
