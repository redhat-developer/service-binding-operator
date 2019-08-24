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

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
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

// TestReconcilerReconcileUsingSecret test the reconciliation process using a secret, expected to be
// the regular approach.
func TestReconcilerReconcileUsingSecret(t *testing.T) {
	ctx := context.TODO()
	resourceRef := "test-using-secret"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedServiceBindingRequest(reconcilerName, resourceRef, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedDatabaseCR(resourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedSecret("db-credentials")

	fakeClient := f.FakeClient()
	reconciler := &Reconciler{client: fakeClient, dynClient: f.FakeDynClient(), scheme: f.S}

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

		sbrOutput := v1alpha1.ServiceBindingRequest{}
		require.Nil(t, fakeClient.Get(ctx, namespacedName, &sbrOutput))
		require.Equal(t, "Success", sbrOutput.Status.BindingStatus)
		require.Equal(t, reconcilerName, sbrOutput.Status.Secret)

		require.Equal(t, 1, len(sbrOutput.Status.ApplicationObjects))
		expectedStatus := fmt.Sprintf("%s/%s", reconcilerNs, reconcilerName)
		assert.Equal(t, expectedStatus, sbrOutput.Status.ApplicationObjects[0])
	})
}

// TestReconcilerForForcedTriggeringOfBinding test the reconciliation process using a secret,
// and using TriggerRebind = true, false
func TestReconcilerForForcedTriggeringOfBinding(t *testing.T) {
	ctx := context.TODO()
	resourceRef := "test-for-forced-trigger"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedServiceBindingRequest(reconcilerName, resourceRef, matchLabels)

	f.AddMockedUnstructuredCSV("cluster-service-version-list-forced-trigger")
	f.AddMockedDatabaseCRList(resourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedSecret("db-credentials")

	fakeClient := f.FakeClient()
	reconciler := &Reconciler{client: fakeClient, dynClient: f.FakeDynClient(), scheme: f.S}

	t.Run("reconcile-using-trigger-starting-with-true", func(t *testing.T) {

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput := v1alpha1.ServiceBindingRequest{}
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &sbrOutput))

		triggertrue := true

		sbrOutput.Spec.TriggerRebinding = &triggertrue

		// set to True and reconcile
		require.NoError(t, fakeClient.Update(ctx, &sbrOutput))

		res, err := reconciler.Reconcile(reconcileRequest())
		assert.Nil(t, err)
		assert.False(t, res.Requeue)

		d := appsv1.Deployment{}
		require.Nil(t, fakeClient.Get(ctx, namespacedName, &d))

		containers := d.Spec.Template.Spec.Containers
		assert.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)

		sbrOutput = v1alpha1.ServiceBindingRequest{}
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &sbrOutput))

		// If TRUE was set, this will become FALSE
		require.False(t, *sbrOutput.Spec.TriggerRebinding)
	})

	t.Run("reconcile-using-trigger-starting-with-false", func(t *testing.T) {

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput := v1alpha1.ServiceBindingRequest{}
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &sbrOutput))

		triggerfalse := false

		sbrOutput.Spec.TriggerRebinding = &triggerfalse

		// set to False and reconcile
		require.NoError(t, fakeClient.Update(ctx, &sbrOutput))

		res, err := reconciler.Reconcile(reconcileRequest())
		assert.Nil(t, err)
		assert.False(t, res.Requeue)

		d := appsv1.Deployment{}
		require.Nil(t, fakeClient.Get(ctx, namespacedName, &d))

		containers := d.Spec.Template.Spec.Containers
		assert.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)

		sbrOutput = v1alpha1.ServiceBindingRequest{}
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &sbrOutput))

		// If FALSE was set, this will stay FALSE
		require.False(t, *sbrOutput.Spec.TriggerRebinding)
	})
}

func TestReconcilerReconcileUsingVolumes(t *testing.T) {
	ctx := context.TODO()
	resourceRef := "test-using-volumes"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedServiceBindingRequest(reconcilerName, resourceRef, matchLabels)
	f.AddMockedUnstructuredCSVWithVolumeMount("cluster-service-version-list")
	f.AddMockedDatabaseCR(resourceRef)
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
