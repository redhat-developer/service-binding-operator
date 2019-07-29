package servicebindingrequest

import (
	"context"
	"testing"

	pgapis "github.com/baijum/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/baijum/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

const (
	reconcilerNs   = "testing"
	reconcilerName = "binding-request"
)

var reconciler *Reconciler
var reconcilerFakeClient client.Client

// TestReconcilerNew this method acts as a "new" call would, but in this scenario bootstraping the
// types and requirements to test Reconcile.
func TestReconcilerNew(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	s := scheme.Scheme
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}

	sbr := mocks.ServiceBindingRequestMock(reconcilerNs, reconcilerName, resourceRef, matchLabels)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &sbr)

	require.Nil(t, olmv1alpha1.AddToScheme(s))
	csvList := mocks.ClusterServiceVersionListMock(reconcilerNs, "cluster-service-version-list")
	s.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &csvList)

	require.Nil(t, pgapis.AddToScheme(s))
	crList := mocks.DatabaseCRListMock(reconcilerNs, resourceRef)
	s.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &crList)

	dbSecret := mocks.SecretMock(reconcilerNs, "db-credentials")

	require.Nil(t, appsv1.AddToScheme(s))
	d := mocks.DeploymentMock(reconcilerNs, reconcilerName, matchLabels)
	s.AddKnownTypes(appsv1.SchemeGroupVersion, &d)

	objs := []runtime.Object{&sbr, &csvList, &crList, &dbSecret, &d}
	reconcilerFakeClient = fake.NewFakeClientWithScheme(s, objs...)
	reconciler = &Reconciler{client: reconcilerFakeClient, scheme: s}
}

func TestReconcilerReconcile(t *testing.T) {
	TestReconcilerNew(t)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: reconcilerNs,
			Name:      reconcilerName,
		},
	}

	res, err := reconciler.Reconcile(req)
	t.Logf("Reconcile error: '%#v'", err)
	require.Nil(t, err)
	require.False(t, res.Requeue)

	namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
	d := appsv1.Deployment{}
	require.Nil(t, reconcilerFakeClient.Get(context.TODO(), namespacedName, &d))

	containers := d.Spec.Template.Spec.Containers
	require.Equal(t, 1, len(containers))
	require.Equal(t, 1, len(containers[0].EnvFrom))
	assert.NotNil(t, containers[0].EnvFrom[0].SecretRef)
	assert.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)
}
