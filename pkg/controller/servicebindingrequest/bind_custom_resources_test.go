package servicebindingrequest

import (
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
	uc, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&configMap)
	us.Object = uc
	us.SetOwnerReferences([]v1.OwnerReference{reference})
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.AddMockResource(cr)
	f.AddMockResource(us)

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
		},
		f.FakeDynClient(),
	)

	t.Run("Should return configmap as owned resource", func(t *testing.T) {
		resources, err := b.GetOwnedResources()
		assert.NoError(t, err)
		assert.Equal(t, 1,len(resources), "Should return 1 owned resource")
	})

	t.Run("Should return all variables exist in the configmap data section", func(t *testing.T) {
		data, err := b.GetBindableVariables()
		assert.NoError(t, err)
		assert.Equal(t, 2,len(data), "")
	})
}
