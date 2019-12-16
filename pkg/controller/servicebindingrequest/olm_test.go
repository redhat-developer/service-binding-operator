package servicebindingrequest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

const (
	dataUserAnnotation     = "servicebindingoperator.redhat.io/status.dbCredentials-data.user"
	dataPasswordAnnotation = "servicebindingoperator.redhat.io/status.dbCredentials-data.password"
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

	t.Run("SelectCRDDescriptionByGVK", func(t *testing.T) {
		// FIXME: include test for populated CRD
		crd, err := olm.SelectCRDDescriptionByGVK(schema.GroupVersionKind{
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

func TestOLMBuildDescriptorsFromAnnotations(t *testing.T) {
	type args struct {
		annotations map[string]string
	}
	tests := []struct {
		name             string
		args             args
		specDescriptor   *olmv1alpha1.SpecDescriptor
		statusDescriptor *olmv1alpha1.StatusDescriptor
		wantErr          bool
	}{
		{
			name: "secret",
			args: args{
				annotations: map[string]string{
					dataUserAnnotation:     "binding:env:object:secret",
					dataPasswordAnnotation: "binding:env:object:secret",
				},
			},
			specDescriptor: nil,
			statusDescriptor: &olmv1alpha1.StatusDescriptor{
				Path:        "dbCredentials",
				DisplayName: "",
				Description: "",
				XDescriptors: []string{
					"binding:env:object:secret:data.password",
					"binding:env:object:secret:data.user",
				},
				Value: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specDescriptor, statusDescriptor, err :=
				buildDescriptorsFromAnnotations(tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"buildDescriptorsFromAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			requireYamlEqual(t, specDescriptor, tt.specDescriptor)
			requireYamlEqual(t, statusDescriptor, tt.statusDescriptor)
		})
	}
}

func TestOLMBuildCRDDescriptionFromCRDAnnotations(t *testing.T) {
	type args struct {
		crd *unstructured.Unstructured
	}

	tests := []struct {
		name           string
		args           args
		crdDescription *olmv1alpha1.CRDDescription
		wantErr        bool
	}{
		{
			name: "happy path",
			args: args{
				crd: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"spec": map[string]interface{}{
							"names": map[string]interface{}{
								"kind": "CustomResourceDefinitionKind",
							},
							"version": "v1alpha1",
						},
						"metadata": map[string]interface{}{
							"annotations": map[string]interface{}{
								dataUserAnnotation:     "binding:env:object:secret",
								dataPasswordAnnotation: "binding:env:object:secret",
							},
						},
					},
				},
			},
			crdDescription: &olmv1alpha1.CRDDescription{
				Kind:    "CustomResourceDefinitionKind",
				Version: "v1alpha1",
				StatusDescriptors: []olmv1alpha1.StatusDescriptor{
					{
						Path: "dbCredentials",
						XDescriptors: []string{
							"binding:env:object:secret:data.password",
							"binding:env:object:secret:data.user",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crdDescription, err := buildCRDDescriptionFromCRD(tt.args.crd)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildCRDDescriptionFromCRD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			requireYamlEqual(t, crdDescription, tt.crdDescription)
		})
	}
}

// requireYamlEqual compares two values and return a message with the different context
func requireYamlEqual(t *testing.T, a interface{}, b interface{}) {
	yamlActual, _ := yaml.Marshal(a)
	yamlExpected, _ := yaml.Marshal(b)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(yamlExpected)),
		B:        difflib.SplitLines(string(yamlActual)),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  4,
	}

	message, err := difflib.GetUnifiedDiffString(diff)
	require.NoError(t, err)
	if len(message) > 0 {
		t.Errorf("expected is different from actual:\n%s", message)
	}
}
