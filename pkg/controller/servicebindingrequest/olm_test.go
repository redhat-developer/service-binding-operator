package servicebindingrequest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func assertGVKs(t *testing.T, gvks []schema.GroupVersionKind) {
	for _, gvk := range gvks {
		t.Logf("Inspecting GVK: '%s'", gvk)
		require.NotEmpty(t, gvk.Group)
		require.NotEmpty(t, gvk.Version)
		require.NotEmpty(t, gvk.Kind)
	}
}

func TestOLMNew(t *testing.T) {
	ns := "controller"
	csvName := "unit-csv"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV(csvName)
	client := f.FakeDynClient()
	olm := NewOLM(client, ns)

	t.Run("listCSVs", func(t *testing.T) {
		csvs, err := olm.listCSVs()
		require.NoError(t, err)
		require.Len(t, csvs, 1)
	})

	t.Run("ListCSVOwnedCRDs", func(t *testing.T) {
		crds, err := olm.ListCSVOwnedCRDs()
		require.NoError(t, err)
		require.Len(t, crds, 1)
	})

	t.Run("SelectCRDByGVK", func(t *testing.T) {
		// FIXME: include test for populated CRD
		crd, err := olm.SelectCRDByGVK(schema.GroupVersionKind{
			Group:   mocks.CRDName,
			Version: mocks.CRDVersion,
			Kind:    mocks.CRDKind,
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, crd)
		expectedCRDName := strings.ToLower(fmt.Sprintf("%s.%s", mocks.CRDKind, mocks.CRDName))
		require.Equal(t, expectedCRDName, crd.Name)
	})

	t.Run("ListCSVOwnedCRDsAsGVKs", func(t *testing.T) {
		gvks, err := olm.ListCSVOwnedCRDsAsGVKs()
		require.NoError(t, err)
		require.Len(t, gvks, 1)
		assertGVKs(t, gvks)
	})

	t.Run("ListGVKsFromCSVNamespacedName", func(t *testing.T) {
		namespacedName := types.NamespacedName{Namespace: ns, Name: csvName}
		gvks, err := olm.ListGVKsFromCSVNamespacedName(namespacedName)
		require.NoError(t, err)
		require.Len(t, gvks, 1)
		assertGVKs(t, gvks)
	})
}

func TestAnnotationParsing(t *testing.T) {
	annotations := map[string]interface{}{
		"servicebindingoperator.redhat.io/status.dbCredentials-db.password": "binding:env:object:secret",
		"servicebindingoperator.redhat.io/spec.dbName":                      "binding:env:attribute",
		"servicebindingoperator.redhat.io/spec.image":                       "binding:env:attribute",
		"servicebindingoperator.redhat.io/status.dbConfigMap-db.host":       "binding:env:object:configmap",
	}
	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata":   map[string]interface{}{},
			"spec": map[string]interface{}{
				"names": map[string]interface{}{
					"kind": "Carp",
				},
				"group":   "app.dev",
				"version": "v1",
			},
			"status": map[string]interface{}{},
		},
	}
	t.Run("Build CSV from CRD - no annotations", func(t *testing.T) {
		crdDescription, err := buildCRDDescriptionFromCRD(crd)
		require.NoError(t, err)
		require.Len(t, crdDescription.SpecDescriptors, 0)
		require.Len(t, crdDescription.StatusDescriptors, 0)
	})

	t.Run("Build CSV from CRD", func(t *testing.T) {
		crd.Object["metadata"] = map[string]interface{}{
			"annotations": annotations,
		}
		crdDescription, err := buildCRDDescriptionFromCRD(crd)
		require.NoError(t, err)

		require.Len(t, crdDescription.StatusDescriptors, 2)
		require.Len(t, crdDescription.SpecDescriptors, 2)

		expected := map[string]string{
			"dbName":        "binding:env:attribute:spec.dbName",
			"image":         "binding:env:attribute:spec.image",
			"dbCredentials": "binding:env:object:secret:db.password",
			"dbConfigMap":   "binding:env:object:configmap:db.host",
		}

		for _, value := range crdDescription.StatusDescriptors {
			require.Equal(t, expected[value.Path], value.XDescriptors[0])
		}
	})
}
