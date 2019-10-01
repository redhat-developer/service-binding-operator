package servicebindingrequest

import (
	v12 "github.com/openshift/api/route/v1"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestBindNonBindableResources_GetOwnedResources(t *testing.T) {
	f := mocks.NewFake(t, "test")
	cr := mocks.DatabaseCRMock("test", "test")
	trueBool := true
	reference := v1.OwnerReference{
		APIVersion:         cr.APIVersion,
		Kind:               cr.Kind,
		Name:               cr.Name,
		UID:                cr.UID,
		Controller:         &trueBool,
		BlockOwnerDeletion: &trueBool,
	}
	configMap := mocks.ConfigMapMock("test", "test_database")
	us := &unstructured.Unstructured{}
	uc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&configMap)
	assert.NoError(t, err)
	us.Object = uc
	us.SetOwnerReferences([]v1.OwnerReference{reference})
	route, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mocks.RouteCRMock("test", "test"))
	assert.NoError(t, err)
	usRoute := &unstructured.Unstructured{Object: route}
	usRoute.SetOwnerReferences([]v1.OwnerReference{reference})
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.S.AddKnownTypes(v12.SchemeGroupVersion, &v12.Route{})
	f.AddMockResource(cr)
	f.AddMockResource(us)
	f.AddMockResource(&unstructured.Unstructured{Object: route})

	unstructuredCr, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	u := &unstructured.Unstructured{Object: unstructuredCr}
	b := NewBindNonBindable(
		nil,
		u,
		[]schema.GroupVersionResource{
			{
				Group:    "",
				Version:  "v1",
				Resource: "configmaps",
			},
			{
				Group:    "",
				Version:  "v1",
				Resource: "routes",
			},
		},
		f.FakeDynClient(),
	)

	t.Run("Should return configmap as owned resource", func(t *testing.T) {
		resources, err := b.GetOwnedResources()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(resources), "Should return 1 owned resource")
	})

	t.Run("Should return all variables exist in the configmap data section", func(t *testing.T) {
		data, err := b.GetBindableVariables()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(data), "")
	})
}
