package servicebindingrequest

import (
	"fmt"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
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
	configMap.SetOwnerReferences([]v1.OwnerReference{ reference })
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.AddMockResource(cr)
	f.AddMockResource(configMap)

	unstructuredCr, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(cr)
	u := &unstructured.Unstructured{Object:unstructuredCr}
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
	resources, err := b.GetOwnedResources()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	fmt.Println(resources)
}

