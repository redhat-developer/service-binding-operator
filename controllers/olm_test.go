package controllers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func init() {
	log.SetLogger(zap.New(zap.UseDevMode((true))))
}

func TestOLMWithoutCSVCRD(t *testing.T) {
	ns := "controller"
	f := mocks.NewFake(t, ns)
	client := f.FakeDynClient()
	gvr := olmv1alpha1.SchemeGroupVersion.WithResource(csvResource)

	// the original FakeDynClient would not return error for unknown resource
	// prepend our reactor to mock a not found error like a real API server
	client.PrependReactor("*", "*", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if gvr.String() == action.GetResource().String() {
			return true, nil, errors.NewNotFound(gvr.GroupResource(), "the server could not find the requested resource")
		}
		return false, nil, nil
	})
	olm := newOLM(client, ns)

	t.Run("listCSVs without CSV CRD installed", func(t *testing.T) {
		resourceClient := client.Resource(gvr).Namespace(ns)
		objs, err := resourceClient.List(context.TODO(), metav1.ListOptions{})
		require.Error(t, err)
		require.True(t, errors.IsNotFound(err))
		require.Nil(t, objs)

		csvs, err := olm.listCSVs()
		require.NoError(t, err)
		require.Len(t, csvs, 0)
	})
}

func TestOLMNew(t *testing.T) {
	ns := "controller"
	csvName := "unit-csv"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV(csvName)
	client := f.FakeDynClient()
	olm := newOLM(client, ns)

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
		crd, err := olm.selectCRDByGVK(schema.GroupVersionKind{
			Group:   mocks.CRDName,
			Version: mocks.CRDVersion,
			Kind:    mocks.CRDKind,
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, crd)
		expectedCRDName := strings.ToLower(fmt.Sprintf("%s.%s", mocks.CRDKind, mocks.CRDName))
		require.Equal(t, expectedCRDName, crd.Name)
	})
}
