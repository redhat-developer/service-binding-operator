package servicebindingrequest

import (
	"context"
	"testing"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var planner *Planner

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestPlannerNew(t *testing.T) {
	ns := "planner"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, resourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef)
	f.AddMockedUnstructuredDatabaseCRD()

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCR", func(t *testing.T) {
		cr, err := planner.searchCR()

		require.NoError(t, err)
		require.NotNil(t, cr)
	})

	t.Run("plan", func(t *testing.T) {
		plan, err := planner.Plan()

		require.NoError(t, err)
		require.NotNil(t, plan)
		require.NotNil(t, plan.CRDDescription)
		require.NotNil(t, plan.CR)
		require.Equal(t, ns, plan.Ns)
		require.Equal(t, name, plan.Name)
	})
}

func TestPlannerAnnotation(t *testing.T) {
	ns := "planner"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, resourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredDatabaseCRD()

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCRD", func(t *testing.T) {
		crd, err := planner.searchCRD()

		require.NoError(t, err)
		require.NotNil(t, crd)
	})
}
