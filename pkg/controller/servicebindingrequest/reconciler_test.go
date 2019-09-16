package servicebindingrequest

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

const (
	reconcilerNs   = "testing"
	reconcilerName = "binding-request"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

// reconcileRequest creates a reconcile.Request object using global variables.
func reconcileRequest() reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: reconcilerNs,
			Name:      reconcilerName,
		},
	}
}

func TestReconcilerReconcileError(t *testing.T) {
	backingServiceResourceRef := "test-using-secret"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, "", matchLabels)

	fakeClient := f.FakeClient()
	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{client: fakeClient, dynClient: fakeDynClient, scheme: f.S}

	res, err := reconciler.Reconcile(reconcileRequest())
	assert.Error(t, err)
	assert.True(t, res.Requeue)
}

// TestApplicationSelectorByName tests discovery of application by name
func TestApplicationSelectorByName(t *testing.T) {
	backingServiceResourceRef := "backingServiceRef"
	applicationResourceRef := "applicationRef"
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, applicationResourceRef, nil)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, nil)
	f.AddMockedSecret("db-credentials")

	fakeClient := f.FakeClient()
	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{client: fakeClient, dynClient: fakeDynClient, scheme: f.S}

	t.Run("reconcile-using-secret", func(t *testing.T) {

		res, err := reconciler.Reconcile(reconcileRequest())
		assert.NoError(t, err)
		assert.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
		require.NoError(t, err)

		require.Equal(t, "Success", sbrOutput.Status.BindingStatus)
		require.Equal(t, 1, len(sbrOutput.Status.ApplicationObjects))
		expectedStatus := fmt.Sprintf("%s/%s", reconcilerNs, reconcilerName)
		assert.Equal(t, expectedStatus, sbrOutput.Status.ApplicationObjects[0])
	})
}

// TestReconcilerReconcileUsingSecret test the reconciliation process using a secret, expected to be
// the regular approach.
func TestReconcilerReconcileUsingSecret(t *testing.T) {
	ctx := context.TODO()
	backingServiceResourceRef := "test-using-secret"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, "", matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedSecret("db-credentials")

	fakeClient := f.FakeClient()
	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{client: fakeClient, dynClient: fakeDynClient, scheme: f.S}

	t.Run("reconcile-using-secret", func(t *testing.T) {
		res, err := reconciler.Reconcile(reconcileRequest())
		assert.Nil(t, err)
		assert.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		d := appsv1.Deployment{}
		require.Nil(t, fakeClient.Get(ctx, namespacedName, &d))

		containers := d.Spec.Template.Spec.Containers
		require.Equal(t, 1, len(containers))
		require.Equal(t, 1, len(containers[0].EnvFrom))
		assert.NotNil(t, containers[0].EnvFrom[0].SecretRef)
		assert.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)

		namespacedName = types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
		require.NoError(t, err)

		require.Equal(t, "Success", sbrOutput.Status.BindingStatus)
		require.Equal(t, reconcilerName, sbrOutput.Status.Secret)

		require.Equal(t, 1, len(sbrOutput.Status.ApplicationObjects))
		expectedStatus := fmt.Sprintf("%s/%s", reconcilerNs, reconcilerName)
		assert.Equal(t, expectedStatus, sbrOutput.Status.ApplicationObjects[0])
	})
}

func TestReconcilerReconcileUsingVolumes(t *testing.T) {
	ctx := context.TODO()
	backingServiceResourceRef := "test-using-volumes"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, "", matchLabels)
	f.AddMockedUnstructuredCSVWithVolumeMount("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedSecret("db-credentials")

	fakeClient := f.FakeClient()
	reconciler := &Reconciler{client: fakeClient, dynClient: f.FakeDynClient(), scheme: f.S}

	t.Run("reconcile-using-volume", func(t *testing.T) {
		res, err := reconciler.Reconcile(reconcileRequest())
		assert.Nil(t, err)
		assert.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		d := appsv1.Deployment{}
		require.Nil(t, fakeClient.Get(ctx, namespacedName, &d))

		containers := d.Spec.Template.Spec.Containers

		require.Equal(t, 1, len(containers[0].VolumeMounts))
		assert.Equal(t, "/var/redhat", containers[0].VolumeMounts[0].MountPath)
		assert.Equal(t, reconcilerName, containers[0].VolumeMounts[0].Name)

		volumes := d.Spec.Template.Spec.Volumes
		require.Equal(t, 1, len(volumes))
		assert.Equal(t, reconcilerName, volumes[0].Name)
		assert.Equal(t, reconcilerName, volumes[0].VolumeSource.Secret.SecretName)
	})
}
