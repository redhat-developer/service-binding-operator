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

func TestPlanner(t *testing.T) {
	ns := "planner"
	name := "service-binding-request"
	resourceRef := "db-testing"

	f := mocks.NewFake(t, ns)

	// This test is here to stay, we may rename the function after
	// deprecated fields are removed completely.
	sbr := f.AddMockedServiceBindingRequestV1_1(name, resourceRef, "", deploymentsGVR)

	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCR", func(t *testing.T) {

		for _, service := range *sbr.Spec.Services {
			cr, err := planner.searchCR(service)
			require.NoError(t, err)
			require.NotNil(t, cr)
		}

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

	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequestV1_1(name, resourceRef, "", deploymentsGVR)
	f.AddMockedUnstructuredDatabaseCRD()
	cr := f.AddMockedDatabaseCR("database", ns)

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCRD", func(t *testing.T) {
		crd, err := planner.searchCRD(cr.GetObjectKind().GroupVersionKind())

		require.NoError(t, err)
		require.NotNil(t, crd)
	})
}

func TestPlannerDeprecacted(t *testing.T) {
	ns := "planner"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, resourceRef, "", deploymentsGVR, matchLabels)
	sbr.Spec.BackingServiceSelectors = &[]v1alpha1.BackingServiceSelector{
		sbr.Spec.BackingServiceSelector,
	}
	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	// Out of the box, our mocks don't set the namespace
	// ensure SearchCR fails.
	t.Run("search CR with namespace not set", func(t *testing.T) {
		cr, err := planner.searchCR(sbr.Spec.BackingServiceSelector)
		require.Error(t, err)
		require.Nil(t, cr)
	})

	// Plan should pass because namespaces in the
	// selector are set if missing.
	t.Run("plan", func(t *testing.T) {
		plan, err := planner.Plan()

		require.NoError(t, err)
		require.NotNil(t, plan)
		require.NotEmpty(t, plan.RelatedResources)
		require.Equal(t, ns, plan.Ns)
		require.Equal(t, name, plan.Name)
	})

	// The searchCR contract only cares about the backingServiceNamespace
	sbr.Spec.BackingServiceSelector.Namespace = &ns
	t.Run("searchCR", func(t *testing.T) {
		cr, err := planner.searchCR(sbr.Spec.BackingServiceSelector)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestPlannerWithExplicitBackingServiceNamespace(t *testing.T) {

	ns := "planner"
	backingServiceNamespace := "backing-service-namespace"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, &backingServiceNamespace, resourceRef, "", deploymentsGVR, matchLabels)
	require.NotNil(t, sbr.Spec.BackingServiceSelector.Namespace)

	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, backingServiceNamespace)
	f.AddMockedUnstructuredDatabaseCRD()

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCR", func(t *testing.T) {
		cr, err := planner.searchCR(sbr.Spec.BackingServiceSelector)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})

	t.Run("plan : backing service in different namespace", func(t *testing.T) {
		plan, err := planner.Plan()

		require.NoError(t, err)
		require.NotNil(t, plan)
		require.NotEmpty(t, plan.RelatedResources)
		require.NotEmpty(t, plan.RelatedResources.GetCRs())
		require.Equal(t, backingServiceNamespace, plan.RelatedResources.GetCRs()[0].GetNamespace())
		require.Equal(t, ns, plan.Ns)
		require.Equal(t, name, plan.Name)

	})
}

func TestPlannerAnnotationDeprecated(t *testing.T) {
	ns := "planner"
	name := "service-binding-request"
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, resourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredDatabaseCRD()
	cr := f.AddMockedDatabaseCR("database", ns)

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	t.Run("searchCRD", func(t *testing.T) {
		crd, err := planner.searchCRD(cr.GetObjectKind().GroupVersionKind())

		require.NoError(t, err)
		require.NotNil(t, crd)
	})
}
