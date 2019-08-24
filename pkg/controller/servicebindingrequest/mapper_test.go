package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

func TestSBRRequestMapperMap(t *testing.T) {
	mapper := &SBRRequestMapper{}

	// not containing annotations, should return empty
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"})
	mapObj := handler.MapObject{Object: u.DeepCopyObject()}
	mappedRequests := mapper.Map(mapObj)
	require.Equal(t, 0, len(mappedRequests))

	request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "ns", Name: "name"}}

	// with annotations in place it should return the actual values
	u.SetAnnotations(map[string]string{sbrNamespaceAnnotation: "ns", sbrNameAnnotation: "name"})
	mapObj = handler.MapObject{Object: u.DeepCopyObject()}
	mappedRequests = mapper.Map(mapObj)
	require.Equal(t, 1, len(mappedRequests))
	assert.Equal(t, request, mappedRequests[0])

	// it should also understand a actual SBR as well, so return not empty
	sbr := &unstructured.Unstructured{}
	sbr.SetGroupVersionKind(v1alpha1.SchemeGroupVersion.WithKind("ServiceBindingRequest"))
	sbr.SetNamespace("ns")
	sbr.SetName("name")
	mapObj = handler.MapObject{Object: sbr.DeepCopyObject()}
	mappedRequests = mapper.Map(mapObj)
	require.Equal(t, 1, len(mappedRequests))
	assert.Equal(t, request, mappedRequests[0])
}
