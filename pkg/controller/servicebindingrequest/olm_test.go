package servicebindingrequest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func Test_splitBindingInfo(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "ref",
			args:    struct{ s string }{s: "status.configMapRef-password"},
			want:    "status.configMapRef",
			want1:   "password",
			wantErr: false,
		},
		{
			name:    "embedded",
			args:    struct{ s string }{s: "status.connectionString"},
			want:    "status.connectionString",
			want1:   "status.connectionString",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBindingInfo(tt.args.s)
			if err != nil && !tt.wantErr {
				t.Errorf("NewBindingInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil {
				if b.FieldPath != tt.want {
					t.Errorf("NewBindingInfo() got = %v, want %v", b.FieldPath, tt.want)
				}
				if b.Path != tt.want1 {
					t.Errorf("NewBindingInfo() got1 = %v, want %v", b.Path, tt.want1)
				}

			}

		})
	}
}
