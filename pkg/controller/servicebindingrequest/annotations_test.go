package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

func TestAnnotationsExtractNamespacedName(t *testing.T) {
	assert.Equal(t, types.NamespacedName{}, extractNamespacedName(map[string]string{}))

	data := map[string]string{sbrNamespaceAnnotation: "ns", sbrNameAnnotation: "name"}
	assert.Equal(t, types.NamespacedName{Namespace: "ns", Name: "name"}, extractNamespacedName(data))
}

func TestAnnotationsGetSBRNamespacedNameFromObject(t *testing.T) {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"})

	// not containing annotations, should return empty
	t.Run("returns-empty", func(t *testing.T) {
		namespacedName, err := GetSBRNamespacedNameFromObject(u.DeepCopyObject())
		assert.Nil(t, err)
		assert.Equal(t, types.NamespacedName{}, namespacedName)
	})

	// with annotations in place it should return the actual values
	t.Run("from-annotations", func(t *testing.T) {
		u.SetAnnotations(map[string]string{sbrNamespaceAnnotation: "ns", sbrNameAnnotation: "name"})
		namespacedName, err := GetSBRNamespacedNameFromObject(u.DeepCopyObject())
		assert.Nil(t, err)
		assert.Equal(t, types.NamespacedName{Namespace: "ns", Name: "name"}, namespacedName)
	})

	// it should also understand a actual SBR as well, so return not empty
	t.Run("actual-sbr-object", func(t *testing.T) {
		sbr := &unstructured.Unstructured{}
		sbr.SetGroupVersionKind(v1alpha1.SchemeGroupVersion.WithKind(ServiceBindingRequestKind))
		sbr.SetNamespace("ns")
		sbr.SetName("name")
		namespacedName, err := GetSBRNamespacedNameFromObject(sbr.DeepCopyObject())
		assert.Nil(t, err)
		assert.Equal(t, types.NamespacedName{Namespace: "ns", Name: "name"}, namespacedName)
	})
}

func TestAnnotationsIsSBRNamespacedNameEmpty(t *testing.T) {
	assert.True(t, IsSBRNamespacedNameEmpty(types.NamespacedName{}))
	assert.True(t, IsSBRNamespacedNameEmpty(types.NamespacedName{Namespace: "ns"}))
	assert.True(t, IsSBRNamespacedNameEmpty(types.NamespacedName{Name: "name"}))
	assert.False(t, IsSBRNamespacedNameEmpty(types.NamespacedName{Namespace: "ns", Name: "name"}))
}
