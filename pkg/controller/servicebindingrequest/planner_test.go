package servicebindingrequest

import (
	"context"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, resourceRef, "", deploymentsGVR, matchLabels)
	sbr.Spec.BackingServiceSelectors = &[]v1alpha1.BackingServiceSelector{
		*sbr.Spec.BackingServiceSelector,
	}
	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()

	planner = NewPlanner(context.TODO(), f.FakeDynClient(), sbr)
	require.NotNil(t, planner)

	// Out of the box, our mocks don't set the namespace
	// ensure SearchCR fails.
	t.Run("search CR with namespace not set", func(t *testing.T) {
		cr, err := planner.searchCR(*sbr.Spec.BackingServiceSelector)
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
		cr, err := planner.searchCR(*sbr.Spec.BackingServiceSelector)
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
		cr, err := planner.searchCR(*sbr.Spec.BackingServiceSelector)
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

func TestPlannerAnnotation(t *testing.T) {
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

func TestPlannerWithCRAnnotations(t *testing.T) {

	ns := "planner"
	name := "service-binding-request"
	f := mocks.NewFake(t, "test")

	// create a Route
	routeCR := mocks.RouteCRMock(ns, "test")
	annotations := map[string]string{
		"servicebindingoperator.redhat.io/spec.host": "binding:env:attribute",
	}
	routeCR.Annotations = annotations
	route, err := runtime.DefaultUnstructuredConverter.ToUnstructured(routeCR)
	require.NoError(t, err)
	f.S.AddKnownTypes(routev1.SchemeGroupVersion, &routev1.Route{})
	f.AddMockResource(&unstructured.Unstructured{Object: route})

	sbr := &v1alpha1.ServiceBindingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			ApplicationSelector: v1alpha1.ApplicationSelector{
				GroupVersionResource: metav1.GroupVersionResource{Group: "g", Version: "v", Resource: "r"},
				ResourceRef:          "app",
			},
			BackingServiceSelectors: &[]v1alpha1.BackingServiceSelector{
				{
					GroupVersionKind: metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Route"},
					ResourceRef:      routeCR.GetName(),
					Namespace:        &ns,
				},
			},
		},
	}
	f.S.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.ServiceBindingRequest{})

	planner := NewPlanner(context.TODO(), f.FakeDynClient(), sbr)

	t.Run("plan with annotated CR", func(t *testing.T) {
		plan, err := planner.Plan()

		require.NoError(t, err)
		require.Len(t, plan.GetRelatedResources().GetCRs(), 1)

		require.Equal(t, "host", plan.RelatedResources[0].CRDDescription.SpecDescriptors[0].Path)
		require.Equal(t, "binding:env:attribute:spec.host", plan.RelatedResources[0].CRDDescription.SpecDescriptors[0].XDescriptors[0])
	})

	sbr.Spec.BackingServiceSelectors = &[]v1alpha1.BackingServiceSelector{
		{
			GroupVersionKind: metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Route"},
			ResourceRef:      "non-existent",
			Namespace:        &ns,
		},
	}
	t.Run("plan with non existent CR", func(t *testing.T) {
		plan, err := planner.Plan()
		require.Error(t, err)
		require.Nil(t, plan)
	})
}
