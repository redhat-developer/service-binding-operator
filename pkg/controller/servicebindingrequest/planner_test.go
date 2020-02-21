package servicebindingrequest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
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
	sbr.Spec.BackingServiceSelectors = []v1alpha1.BackingServiceSelector{
		sbr.Spec.BackingServiceSelector,
	}
	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef)
	f.AddMockedUnstructuredDatabaseCRD()

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCR", func(t *testing.T) {
		cr, err := planner.searchCR(ns, sbr.Spec.BackingServiceSelector)

		require.NoError(t, err)
		require.NotNil(t, cr)
	})

	t.Run("plan", func(t *testing.T) {
		plan, err := planner.Plan()

		require.NoError(t, err)
		require.NotNil(t, plan)
		require.NotEmpty(t, plan.RelatedResources)
		require.Equal(t, ns, plan.Ns)
		require.Equal(t, name, plan.Name)
	})
}

func TestPlannerNewWithMultipleNamespaces(t *testing.T) {
	ns := "planner"
	backingServiceNs := "planner-backing-service-ns"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequestWithMultiNamespaces(name, resourceRef, backingServiceNs, "", deploymentsGVR, matchLabels)
	sbr.Spec.BackingServiceSelectors = []v1alpha1.BackingServiceSelector{
		sbr.Spec.BackingServiceSelector,
	}
	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddCrossNamespaceMockedDatabaseCR(resourceRef, backingServiceNs)
	f.AddMockedUnstructuredDatabaseCRD()

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCR", func(t *testing.T) {
		cr, err := planner.searchCR(ns, sbr.Spec.BackingServiceSelector)

		require.NoError(t, err)
		require.NotNil(t, cr)
	})

	t.Run("plan", func(t *testing.T) {
		plan, err := planner.Plan()

		require.NoError(t, err)
		require.NotNil(t, plan)
		require.NotEmpty(t, plan.RelatedResources)
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
	cr := f.AddMockedDatabaseCR("database")

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCRD", func(t *testing.T) {
		crd, err := planner.searchCRD(cr.GetObjectKind().GroupVersionKind())

		require.NoError(t, err)
		require.NotNil(t, crd)
	})
}
