package servicebindingrequest

import (
	"fmt"
	"reflect"
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

// Test_splitBindingInfo exercises annotation binding information parsing.
func TestNewBindingInfo(t *testing.T) {
	type args struct {
		s string
		d string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *BindingInfo
	}{
		{
			args: args{s: "status.configMapRef-password", d: "binding"},
			want: &BindingInfo{
				FieldPath:  "status.configMapRef",
				Descriptor: "binding:password",
				Path:       "password",
			},
			name:    "{fieldPath}-{path} annotation",
			wantErr: false,
		},
		{
			args: args{s: "status.connectionString", d: "binding"},
			want: &BindingInfo{
				Descriptor: "binding:status.connectionString",
				FieldPath:  "status.connectionString",
				Path:       "status.connectionString",
			},
			name:    "{path} annotation",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBindingInfo(tt.args.s, tt.args.d)
			if err != nil && !tt.wantErr {
				t.Errorf("NewBindingInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil {
				text, err := yamlDiff(tt.want, b)
				require.NoError(t, err)
				if len(text) > 0 {
					t.Errorf("expected is different from actual:\n%s", text)
				}
			}
		})
	}
}

func Test_buildDescriptorsFromAnnotations(t *testing.T) {
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
					"servicebindingoperator.redhat.io/status.dbCredentials-data.user":     "binding:env:object:secret",
					"servicebindingoperator.redhat.io/status.dbCredentials-data.password": "binding:env:object:secret",
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
			specDescriptor, statusDescriptor, err := buildDescriptorsFromAnnotations(tt.args.annotations)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildDescriptorsFromAnnotations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			text, err := yamlDiff(specDescriptor, tt.specDescriptor)
			if err != nil {
				t.Errorf("buildDescriptorsFromAnnotations() specDescriptor = %v, want %v", specDescriptor, tt.specDescriptor)
			}
			if len(text) > 0 {
				t.Errorf("expected is different from actual:\n%s", text)
			}

			text, err = yamlDiff(statusDescriptor, tt.statusDescriptor)
			if err != nil {
				t.Errorf("buildDescriptorsFromAnnotations() statusDescriptor = %v, want %v", statusDescriptor, tt.statusDescriptor)
			}
			if len(text) > 0 {
				t.Errorf("expected is different from actual:\n%s", text)
			}
		})
	}
}

func Test_buildCRDDescriptionFromCRDAnnotations(t *testing.T) {
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
			name: "",
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
								"servicebindingoperator.redhat.io/status.dbCredentials-data.user":     "binding:env:object:secret",
								"servicebindingoperator.redhat.io/status.dbCredentials-data.password": "binding:env:object:secret",
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
			text, err := yamlDiff(crdDescription, tt.crdDescription)
			if !reflect.DeepEqual(crdDescription, tt.crdDescription) {
				t.Errorf("buildCRDDescriptionFromCRD() got = %v, want %v", crdDescription, tt.crdDescription)
			}
			if len(text) > 0 {
				t.Errorf("expected is different from actual:\n%s", text)
			}
		})
	}
}

func yamlDiff(a interface{}, b interface{}) (string, error) {
	yamlActual, _ := yaml.Marshal(a)
	yamlExpected, _ := yaml.Marshal(b)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(yamlExpected)),
		B:        difflib.SplitLines(string(yamlActual)),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  4,
	}

	return difflib.GetUnifiedDiffString(diff)
}
