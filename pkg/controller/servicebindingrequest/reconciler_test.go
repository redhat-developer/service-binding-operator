package servicebindingrequest

import (
	"context"
	"testing"

	pgapis "github.com/baijum/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/baijum/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
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

// TestReconcilerNew this method acts as a "new" call would, but in this scenario bootstraping the
// types and requirements to test Reconcile.
func TestReconcilerNew(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var reconciler *Reconciler
	var reconcilerFakeClient client.Client

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

	require.Nil(t, extv1beta1.AddToScheme(s))
	d := mocks.DeploymentMock(reconcilerNs, reconcilerName, matchLabels)
	s.AddKnownTypes(extv1beta1.SchemeGroupVersion, &d)

	objs := []runtime.Object{&sbr, &csvList, &crList, &dbSecret, &d}
	reconcilerFakeClient = fake.NewFakeClientWithScheme(s, objs...)
	reconciler = &Reconciler{client: reconcilerFakeClient, scheme: s}

	t.Run("reconcile", func(t *testing.T) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: reconcilerNs,
				Name:      reconcilerName,
			},
		}

		res, err := reconciler.Reconcile(req)
		assert.Nil(t, err)
		assert.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		d := extv1beta1.Deployment{}
		require.Nil(t, reconcilerFakeClient.Get(context.TODO(), namespacedName, &d))

		containers := d.Spec.Template.Spec.Containers
		assert.Equal(t, 1, len(containers))
		assert.Equal(t, 1, len(containers[0].EnvFrom))
		assert.NotNil(t, containers[0].EnvFrom[0].SecretRef)
		assert.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)

		sbrOutput := v1alpha1.ServiceBindingRequest{}
		require.Nil(t, reconcilerFakeClient.Get(context.TODO(), namespacedName, &sbrOutput))
		require.Equal(t, v1alpha1.BindingSuccess, sbrOutput.Status.BindingStatus)
		require.Equal(t, reconcilerName, sbrOutput.Status.LabelObjects[0])
	})

}

// TestReconcilerVolumeMount method acts as a "new" call would, but in this scenario bootstraping the
// types and requirements to test Reconcile.
func TestReconcilerVolumeMount(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var reconciler *Reconciler
	var reconcilerFakeClient client.Client

	s := scheme.Scheme
	resourceRef := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}

	sbr := mocks.ServiceBindingRequestMock(reconcilerNs, reconcilerName, resourceRef, matchLabels)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &sbr)

	require.Nil(t, olmv1alpha1.AddToScheme(s))
	csvList := mocks.ClusterServiceVersionListVolumeMountMock(reconcilerNs, "cluster-service-version-list")
	s.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &csvList)

	require.Nil(t, pgapis.AddToScheme(s))
	crList := mocks.DatabaseCRListMock(reconcilerNs, resourceRef)
	s.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &crList)

	dbSecret := mocks.SecretMock(reconcilerNs, "db-credentials")

	require.Nil(t, extv1beta1.AddToScheme(s))
	d := mocks.DeploymentMock(reconcilerNs, reconcilerName, matchLabels)
	s.AddKnownTypes(extv1beta1.SchemeGroupVersion, &d)

	objs := []runtime.Object{&sbr, &csvList, &crList, &dbSecret, &d}
	reconcilerFakeClient = fake.NewFakeClientWithScheme(s, objs...)
	reconciler = &Reconciler{client: reconcilerFakeClient, scheme: s}

	t.Run("reconcile", func(t *testing.T) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: reconcilerNs,
				Name:      reconcilerName,
			},
		}

		res, err := reconciler.Reconcile(req)
		assert.Nil(t, err)
		assert.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		d := extv1beta1.Deployment{}
		require.Nil(t, reconcilerFakeClient.Get(context.TODO(), namespacedName, &d))

		containers := d.Spec.Template.Spec.Containers
		assert.Equal(t, 1, len(containers))
		assert.Equal(t, 1, len(containers[0].EnvFrom))
		assert.NotNil(t, containers[0].EnvFrom[0].SecretRef)
		assert.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)

		assert.Equal(t, 1, len(containers[0].VolumeMounts))
		assert.Equal(t, "/var/redhat", containers[0].VolumeMounts[0].MountPath)
		assert.Equal(t, reconcilerName, containers[0].VolumeMounts[0].Name)

		volumes := d.Spec.Template.Spec.Volumes
		assert.Equal(t, 1, len(volumes))
		assert.Equal(t, reconcilerName, volumes[0].Name)
		assert.Equal(t, reconcilerName, volumes[0].VolumeSource.Secret.SecretName)
	})
}
