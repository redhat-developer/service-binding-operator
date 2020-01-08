package servicebindingrequest

import (
	"testing"

	v12 "github.com/openshift/api/route/v1"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var trueBool = true

func TestBindNonBindableResources_ConfigMap_GetOwnedResources(t *testing.T) {
	f := mocks.NewFake(t, "test")
	cr := mocks.DatabaseCRMock("test", "test")
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
	require.NoError(t, err)
	us.Object = uc
	us.SetOwnerReferences([]v1.OwnerReference{reference})
	route, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mocks.RouteCRMock("test", "test"))
	require.NoError(t, err)
	usRoute := &unstructured.Unstructured{Object: route}
	usRoute.SetOwnerReferences([]v1.OwnerReference{reference})
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.S.AddKnownTypes(v12.SchemeGroupVersion, &v12.Route{})
	f.AddMockResource(cr)
	f.AddMockResource(us)
	f.AddMockResource(&unstructured.Unstructured{Object: route})

	unstructuredCr, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	u := &unstructured.Unstructured{Object: unstructuredCr}
	b := NewDetectBindableResources(
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
		require.NoError(t, err)
		require.Equal(t, 2, len(resources), "Should return 2 owned resource")
	})

	t.Run("Should return all variables exist in the configmap data section and route", func(t *testing.T) {
		data, err := b.GetBindableVariables()
		str1 := b.data["user"]
		require.NoError(t, err)
		require.Equal(t, 3, len(data), "")
		require.Equal(t, str1, "user", "The intermediate data values are equal ")
	})
}

func TestBindNonBindableResources_Secret_GetOwnedResources(t *testing.T) {
	f := mocks.NewFake(t, "test")
	cr := mocks.DatabaseCRMock("test", "test")
	reference := v1.OwnerReference{
		APIVersion:         cr.APIVersion,
		Kind:               cr.Kind,
		Name:               cr.Name,
		UID:                cr.UID,
		Controller:         &trueBool,
		BlockOwnerDeletion: &trueBool,
	}
	secret := mocks.SecretMock("test", "test_database")
	us := &unstructured.Unstructured{}
	uc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&secret)
	require.NoError(t, err)
	us.Object = uc
	us.SetOwnerReferences([]v1.OwnerReference{reference})
	route, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mocks.RouteCRMock("test", "test"))
	require.NoError(t, err)
	usRoute := &unstructured.Unstructured{Object: route}
	usRoute.SetOwnerReferences([]v1.OwnerReference{reference})
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.S.AddKnownTypes(v12.SchemeGroupVersion, &v12.Route{})
	f.AddMockResource(cr)
	f.AddMockResource(us)
	f.AddMockResource(&unstructured.Unstructured{Object: route})

	unstructuredCr, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	u := &unstructured.Unstructured{Object: unstructuredCr}
	b := NewDetectBindableResources(
		nil,
		u,
		[]schema.GroupVersionResource{
			{
				Group:    "",
				Version:  "v1",
				Resource: "secrets",
			},
			{
				Group:    "",
				Version:  "v1",
				Resource: "routes",
			},
		},
		f.FakeDynClient(),
	)

	t.Run("Should return secret as owned resource", func(t *testing.T) {
		resources, err := b.GetOwnedResources()
		require.NoError(t, err)
		require.Equal(t, 2, len(resources), "Should return 2 owned resource")
	})

	t.Run("Should return all variables exist in the secret data section and route", func(t *testing.T) {
		data, err := b.GetBindableVariables()
		str2 := (b.data["user"]).([]byte)
		require.NoError(t, err)
		require.Equal(t, 3, len(data), "")
		require.Equal(t, str2, []byte("user"), "The intermediate data values are equal ")
	})
}
