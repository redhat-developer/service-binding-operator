package servicebindingrequest

import (
	"context"
	"reflect"
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

const (
	reconcilerNs   = "testing"
	reconcilerName = "binding-request"
)

var (
	deploymentsGVR       = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	deploymentConfigsGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deploymentconfigs"}
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

// TestApplicationSelectorByName tests discovery of application by name
func TestApplicationSelectorByName(t *testing.T) {
	backingServiceResourceRef := "backingServiceRef"
	applicationResourceRef := "applicationRef"
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, applicationResourceRef, deploymentsGVR, nil)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, nil)
	f.AddMockedUnstructuredSecret("db-credentials")

	restMapper := testutils.BuildTestRESTMapper()

	fakeClient := f.FakeClient()
	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{
		RestMapper: restMapper,
		client:     fakeClient,
		dynClient:  fakeDynClient,
		scheme:     f.S,
	}

	t.Run("test-application-selector-by-name", func(t *testing.T) {

		res, err := reconciler.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
		require.NoError(t, err)

		require.True(t,
			conditionsv1.IsStatusConditionPresentAndEqual(
				sbrOutput.Status.Conditions,
				CollectionReady,
				corev1.ConditionTrue,
			),
			"CollectionReady condition should exist and true; existing conditions: %+v",
			sbrOutput.Status.Conditions,
		)

		require.True(t,
			conditionsv1.IsStatusConditionPresentAndEqual(
				sbrOutput.Status.Conditions,
				InjectionReady,
				corev1.ConditionTrue,
			),
			"InjectionReady condition should exist and true; existing conditions: %+v",
			sbrOutput.Status.Conditions,
		)

		require.Equal(t, 1, len(sbrOutput.Status.Applications))
		expectedStatus := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: namespacedName.Name,
			},
		}
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))
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
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedUnstructuredSecret("db-credentials")

	restMapper := testutils.BuildTestRESTMapper()

	fakeClient := f.FakeClient()
	fakeDynClient := f.FakeDynClient()
	reconciler := &Reconciler{
		RestMapper: restMapper,
		client:     fakeClient,
		dynClient:  fakeDynClient,
		scheme:     f.S,
	}

	t.Run("reconcile-using-secret", func(t *testing.T) {
		res, err := reconciler.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		d := appsv1.Deployment{}
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &d))

		containers := d.Spec.Template.Spec.Containers
		require.Equal(t, 1, len(containers))
		require.Equal(t, 1, len(containers[0].EnvFrom))
		require.NotNil(t, containers[0].EnvFrom[0].SecretRef)
		require.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)

		namespacedName = types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
		require.NoError(t, err)

		require.True(t,
			conditionsv1.IsStatusConditionPresentAndEqual(
				sbrOutput.Status.Conditions,
				CollectionReady,
				corev1.ConditionTrue,
			),
			"CollectionReady condition should exist and true; existing conditions: %+v",
			sbrOutput.Status.Conditions,
		)

		require.True(t,
			conditionsv1.IsStatusConditionPresentAndEqual(
				sbrOutput.Status.Conditions,
				InjectionReady,
				corev1.ConditionTrue,
			),
			"InjectionReady condition should exist and true; existing conditions: %+v",
			sbrOutput.Status.Conditions,
		)

		require.Equal(t, reconcilerName, sbrOutput.Status.Secret)

		require.Equal(t, 1, len(sbrOutput.Status.Applications))
		expectedStatus := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: namespacedName.Name,
			},
		}
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))
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
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSVWithVolumeMount("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedUnstructuredSecret("db-credentials")

	restMapper := testutils.BuildTestRESTMapper()

	fakeClient := f.FakeClient()
	reconciler := &Reconciler{
		RestMapper: restMapper,
		client:     fakeClient,
		dynClient:  f.FakeDynClient(),
		scheme:     f.S,
	}

	t.Run("reconcile-using-volume", func(t *testing.T) {
		res, err := reconciler.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		d := appsv1.Deployment{}
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &d))

		containers := d.Spec.Template.Spec.Containers

		require.Equal(t, 1, len(containers[0].VolumeMounts))
		require.Equal(t, "/var/redhat", containers[0].VolumeMounts[0].MountPath)
		require.Equal(t, reconcilerName, containers[0].VolumeMounts[0].Name)

		volumes := d.Spec.Template.Spec.Volumes
		require.Equal(t, 1, len(volumes))
		require.Equal(t, reconcilerName, volumes[0].Name)
		require.Equal(t, reconcilerName, volumes[0].VolumeSource.Secret.SecretName)
	})
}

func TestApplicationNotFound(t *testing.T) {
	ctx := context.TODO()
	backingServiceResourceRef := "backingService1"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredSecret("db-credentials")

	restMapper := testutils.BuildTestRESTMapper()

	fakeClient := f.FakeClient()
	reconciler := &Reconciler{
		RestMapper: restMapper,
		client:     fakeClient,
		dynClient:  f.FakeDynClient(),
		scheme:     f.S,
	}

	// Reconcile without deployment
	res, err := reconciler.Reconcile(reconcileRequest())
	require.EqualError(t, err, ErrApplicationNotFound.Error())
	require.False(t, res.Requeue)

	namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
	sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
	require.NoError(t, err)

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput.Status.Conditions,
			CollectionReady,
			corev1.ConditionTrue,
		),
		"CollectionReady condition should exist and true; existing conditions: %+v",
		sbrOutput.Status.Conditions,
	)
	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput.Status.Conditions,
			InjectionReady,
			corev1.ConditionFalse,
		),
		"InjectionReady condition should exist and false; existing conditions: %+v",
		sbrOutput.Status.Conditions,
	)
	require.Len(t, sbrOutput.Status.Applications, 0)

	// Reconcile with deployment
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)

	fakeClient = f.FakeClient()
	reconciler = &Reconciler{
		RestMapper: restMapper,
		client:     fakeClient,
		dynClient:  f.FakeDynClient(),
		scheme:     f.S,
	}
	res, err = reconciler.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	d := appsv1.Deployment{}
	require.NoError(t, fakeClient.Get(ctx, namespacedName, &d))

	sbrOutput2, err := reconciler.getServiceBindingRequest(namespacedName)
	require.NoError(t, err)

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput2.Status.Conditions,
			CollectionReady,
			corev1.ConditionTrue,
		),
		"CollectionReady condition should exist and true; existing conditions: %+v",
		sbrOutput2.Status.Conditions,
	)

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput2.Status.Conditions,
			InjectionReady,
			corev1.ConditionTrue,
		),
		"InjectionReady condition should exist and true; existing conditions: %+v",
		sbrOutput2.Status.Conditions,
	)

	require.Equal(t, reconcilerName, sbrOutput2.Status.Secret)
	require.Equal(t, 1, len(sbrOutput2.Status.Applications))

	// Update Credentials
	s := corev1.Secret{}
	secretNamespaced := types.NamespacedName{Namespace: reconcilerNs, Name: "db-credentials"}
	require.NoError(t, fakeClient.Get(ctx, secretNamespaced, &s))
	s.Data["password"] = []byte("abc123")
	require.NoError(t, fakeClient.Update(ctx, &s))

	reconciler = &Reconciler{
		RestMapper: restMapper,
		client:     fakeClient,
		dynClient:  f.FakeDynClient(),
		scheme:     f.S,
	}

	res, err = reconciler.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	sbrOutput3, err := reconciler.getServiceBindingRequest(namespacedName)
	require.NoError(t, err)

	d = appsv1.Deployment{}
	require.NoError(t, fakeClient.Get(ctx, namespacedName, &d))

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput3.Status.Conditions,
			CollectionReady,
			corev1.ConditionTrue,
		),
		"CollectionReady condition should exist and true; existing conditions: %+v",
		sbrOutput3.Status.Conditions,
	)

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput3.Status.Conditions,
			InjectionReady,
			corev1.ConditionTrue,
		),
		"InjectionReady condition should exist and true; existing conditions: %+v",
		sbrOutput3.Status.Conditions,
	)

	require.Equal(t, reconcilerName, sbrOutput3.Status.Secret)
	require.Equal(t, s.Data["password"], []byte("abc123"))
	require.Equal(t, 1, len(sbrOutput3.Status.Applications))
}

//TestReconcilerReconcileWithConflictingAppSelc tests when sbr has conflicting ApplicationSel such as MatchLabels=App1 and ResourceRef=App2 it should prioritise the ResourceRef
func TestReconcilerReconcileWithConflictingAppSelc(t *testing.T) {
	backingServiceResourceRef := "backingServiceRef"
	applicationResourceRef1 := "applicationResourceRef1"
	matchLabels1 := map[string]string{
		"connects-to": "database",
		"environment": "testing",
	}
	applicationResourceRef2 := "applicationResourceRef2"

	f := mocks.NewFake(t, reconcilerNs)

	f.AddMockedUnstructuredDeployment(applicationResourceRef1, matchLabels1)
	f.AddMockedUnstructuredDeployment(applicationResourceRef2, nil)
	f.AddMockedUnstructuredServiceBindingRequest(reconcilerName, backingServiceResourceRef, applicationResourceRef2, deploymentsGVR, matchLabels1)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeClient := f.FakeClient()
	fakeDynClient := f.FakeDynClient()

	restMapper := testutils.BuildTestRESTMapper()

	reconciler := &Reconciler{
		client:     fakeClient,
		dynClient:  fakeDynClient,
		scheme:     f.S,
		RestMapper: restMapper,
	}

	t.Run("test-reconciler-reconcile-with-conflicting-application-selector", func(t *testing.T) {

		res, err := reconciler.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
		require.NoError(t, err)

		expectedStatus := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: applicationResourceRef2,
			},
		}

		require.True(t,
			conditionsv1.IsStatusConditionPresentAndEqual(
				sbrOutput.Status.Conditions,
				CollectionReady,
				corev1.ConditionTrue,
			),
			"CollectionReady condition should exist and true; existing conditions: %+v",
			sbrOutput.Status.Conditions,
		)

		require.True(t,
			conditionsv1.IsStatusConditionPresentAndEqual(
				sbrOutput.Status.Conditions,
				InjectionReady,
				corev1.ConditionTrue,
			),
			"InjectionReady condition should exist and true; existing conditions: %+v",
			sbrOutput.Status.Conditions,
		)

		require.Equal(t, reconcilerName, sbrOutput.Status.Secret)
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))
	})
}

// TestEmptyApplicationSelector tests that Status is successfully updated when ApplicationSelector is missing
func TestEmptyApplicationSelector(t *testing.T) {
	backingServiceResourceRef := "backingService1"
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingRequestWithoutApplication(reconcilerName, backingServiceResourceRef)
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)

	restMapper := testutils.BuildTestRESTMapper()

	fakeClient := f.FakeClient()
	reconciler := &Reconciler{
		RestMapper: restMapper,
		client:     fakeClient,
		dynClient:  f.FakeDynClient(),
		scheme:     f.S,
	}

	res, err := reconciler.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
	sbrOutput, err := reconciler.getServiceBindingRequest(namespacedName)
	require.NoError(t, err)

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput.Status.Conditions,
			CollectionReady,
			corev1.ConditionTrue,
		),
		"CollectionReady condition should exist and true; existing conditions: %+v",
		sbrOutput.Status.Conditions,
	)

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrOutput.Status.Conditions,
			InjectionReady,
			corev1.ConditionFalse,
		),
		"InjectionReady condition should exist and false; existing conditions: %+v",
		sbrOutput.Status.Conditions,
	)
}
