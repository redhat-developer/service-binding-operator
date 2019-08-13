package planner

import (
	"context"
	"testing"

	pgapis "github.com/baijum/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/baijum/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

var planner *Planner

func TestPlannerNew(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	ns := "testing"
	resourceRef := "db-testing"
	s := scheme.Scheme

	sbr := mocks.ServiceBindingRequestMock(ns, "binding-request", resourceRef, map[string]string{
		"connects-to": "database",
		"environment": "planner",
	})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &sbr)

	require.Nil(t, olmv1alpha1.AddToScheme(s))
	csvList := mocks.ClusterServiceVersionListMock(ns, "cluster-service-version-list")
	s.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &csvList)

	require.Nil(t, pgapis.AddToScheme(s))
	crList := mocks.DatabaseCRListMock(ns, resourceRef)
	s.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &crList)

	objs := []runtime.Object{&sbr, &csvList, &crList}
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	planner = NewPlanner(context.TODO(), fakeClient, ns, &sbr)
	require.NotNil(t, planner)
}

func TestPlannerSearchCRDDescription(t *testing.T) {
	TestPlannerNew(t)
	crdDescription, err := planner.searchCRDDescription()

	assert.Nil(t, err)
	assert.NotNil(t, crdDescription)
}

func TestPlannerSearchCR(t *testing.T) {
	cr, err := planner.searchCR("Database", mocks.WaitCallbackMock{})

	assert.Nil(t, err)
	assert.NotNil(t, cr)
}

func TestPlannerPlan(t *testing.T) {
	plan, err := planner.Plan(mocks.WaitCallbackMock{})

	assert.Nil(t, err)
	assert.NotNil(t, plan)
	assert.NotNil(t, plan.CRDDescription)
	assert.NotNil(t, plan.CR)
	assert.Equal(t, "testing", plan.Ns)
	assert.Equal(t, "binding-request", plan.Name)
}
