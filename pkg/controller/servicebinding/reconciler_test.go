package servicebinding

import (
	"reflect"
	"testing"
	"time"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
	secretsGVR           = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
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

func requireConditionPresentAndTrue(t *testing.T, condition conditionsv1.ConditionType, sbrConditions []conditionsv1.Condition) {
	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrConditions,
			condition,
			corev1.ConditionTrue,
		),
		"%+v should exist and be true; existing conditions: %+v",
		condition,
		sbrConditions,
	)
}

func requireConditionPresentAndFalse(t *testing.T, condition conditionsv1.ConditionType, sbrConditions []conditionsv1.Condition) {
	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbrConditions,
			condition,
			corev1.ConditionFalse,
		),
		"%+v should exist and be false; existing conditions: %+v",
		condition,
		sbrConditions,
	)
}

type fakeResourceWatcher struct {
	record map[schema.GroupVersionKind]bool
	mapper meta.RESTMapper
}

var _ ResourceWatcher = (*fakeResourceWatcher)(nil)

func (f *fakeResourceWatcher) AddWatchForGVR(gvr schema.GroupVersionResource) error {
	gvk, err := f.mapper.KindFor(gvr)
	if err != nil {
		return err
	}
	f.record[gvk] = true
	return nil
}

func (f *fakeResourceWatcher) AddWatchForGVK(gvk schema.GroupVersionKind) error {
	f.record[gvk] = true
	return nil
}

func newFakeResourceWatcher(mapper meta.RESTMapper) *fakeResourceWatcher {
	return &fakeResourceWatcher{
		record: make(map[schema.GroupVersionKind]bool),
		mapper: mapper,
	}
}

// TestApplicationByName tests discovery of application by name
func TestApplicationByName(t *testing.T) {
	backingServiceResourceRef := "backingServiceRef"
	applicationResourceRef := "applicationRef"
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBinding(reconcilerName, backingServiceResourceRef, applicationResourceRef, deploymentsGVR, nil)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(applicationResourceRef, nil)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	mapper := testutils.BuildTestRESTMapper()
	r := &reconciler{dynClient: fakeDynClient, restMapper: mapper, scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)

	t.Run("test-application-selector-by-name", func(t *testing.T) {

		res, err := r.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := r.getServiceBinding(namespacedName)
		require.NoError(t, err)

		requireConditionPresentAndTrue(t, CollectionReady, sbrOutput.Status.Conditions)
		requireConditionPresentAndTrue(t, InjectionReady, sbrOutput.Status.Conditions)

		require.Equal(t, 1, len(sbrOutput.Status.Applications))
		expectedStatus := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: applicationResourceRef,
			},
		}
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))
	})
}

// TestReconcilerReconcileUsingSecret test the reconciliation process using a secret, expected to be
// the regular approach.
func TestReconcilerReconcileUsingSecret(t *testing.T) {
	backingServiceResourceRef := "test-using-secret"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBinding(reconcilerName, backingServiceResourceRef, reconcilerName, deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	mapper := testutils.BuildTestRESTMapper()
	r := &reconciler{dynClient: fakeDynClient, restMapper: mapper, scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)

	t.Run("reconcile-using-secret", func(t *testing.T) {
		res, err := r.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}

		u, err := fakeDynClient.Resource(deploymentsGVR).Get(reconcilerName, metav1.GetOptions{})
		require.NoError(t, err)

		d := appsv1.Deployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &d)
		require.NoError(t, err)

		containers := d.Spec.Template.Spec.Containers
		require.Equal(t, 1, len(containers))
		require.Equal(t, 1, len(containers[0].EnvFrom))
		require.NotNil(t, containers[0].EnvFrom[0].SecretRef)
		require.Equal(t, reconcilerName, containers[0].EnvFrom[0].SecretRef.Name)

		namespacedName = types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := r.getServiceBinding(namespacedName)
		require.NoError(t, err)

		requireConditionPresentAndTrue(t, CollectionReady, sbrOutput.Status.Conditions)
		requireConditionPresentAndTrue(t, InjectionReady, sbrOutput.Status.Conditions)

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
	t.Skip("there is not an equivalent yet for volume mounting")

	backingServiceResourceRef := "test-using-volumes"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBinding(reconcilerName, backingServiceResourceRef, reconcilerName, deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSVWithVolumeMount("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	mapper := testutils.BuildTestRESTMapper()
	r := &reconciler{dynClient: fakeDynClient, restMapper: mapper, scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)

	t.Run("reconcile-using-volume", func(t *testing.T) {
		res, err := r.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		u, err := fakeDynClient.Resource(deploymentsGVR).Get(reconcilerName, metav1.GetOptions{})
		require.NoError(t, err)

		d := appsv1.Deployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &d)
		require.NoError(t, err)

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
	backingServiceResourceRef := "backingService1"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBinding(reconcilerName, backingServiceResourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	mapper := testutils.BuildTestRESTMapper()
	r := &reconciler{dynClient: fakeDynClient, restMapper: mapper, scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)

	// Reconcile without deployment
	res, err := r.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
	sbrOutput, err := r.getServiceBinding(namespacedName)
	require.NoError(t, err)

	requireConditionPresentAndTrue(t, CollectionReady, sbrOutput.Status.Conditions)
	requireConditionPresentAndFalse(t, InjectionReady, sbrOutput.Status.Conditions)
	require.Len(t, sbrOutput.Status.Applications, 0)

	// Reconcile with deployment
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)
	fakeDynClient = f.FakeDynClient()
	r = &reconciler{dynClient: fakeDynClient, restMapper: testutils.BuildTestRESTMapper(), scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)
	res, err = r.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	u, err := fakeDynClient.Resource(deploymentsGVR).Get(reconcilerName, metav1.GetOptions{})
	require.NoError(t, err)

	d := appsv1.Deployment{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &d)
	require.NoError(t, err)
	sbrOutput2, err := r.getServiceBinding(namespacedName)
	require.NoError(t, err)

	requireConditionPresentAndTrue(t, CollectionReady, sbrOutput2.Status.Conditions)
	requireConditionPresentAndTrue(t, InjectionReady, sbrOutput2.Status.Conditions)

	require.Equal(t, reconcilerName, sbrOutput2.Status.Secret)
	require.Equal(t, 1, len(sbrOutput2.Status.Applications))
}

func TestReconcilerUpdateCredentials(t *testing.T) {
	backingServiceResourceRef := "backingService"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBinding(reconcilerName, backingServiceResourceRef, "", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredCSV("cluster-service-version-list")
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredSecret("db-credentials")
	f.AddMockedUnstructuredDeployment(reconcilerName, matchLabels)

	fakeDynClient := f.FakeDynClient()
	mapper := testutils.BuildTestRESTMapper()
	r := &reconciler{dynClient: fakeDynClient, restMapper: mapper, scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)

	u, err := fakeDynClient.Resource(secretsGVR).Get("db-credentials", metav1.GetOptions{})
	require.NoError(t, err)
	s := corev1.Secret{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &s)
	require.NoError(t, err)

	// Update Credentials
	s.Data["password"] = []byte("abc123")
	// Update resourceVersion for postgresdb
	s.ObjectMeta.ResourceVersion = "112200"
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&s)
	require.NoError(t, err)
	updated := unstructured.Unstructured{Object: obj}
	_, err = fakeDynClient.Resource(secretsGVR).Namespace(updated.GetNamespace()).Update(&updated, metav1.UpdateOptions{})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
	res, err := r.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
	sbrOutput3, err := r.getServiceBinding(namespacedName)
	require.NoError(t, err)

	requireConditionPresentAndTrue(t, CollectionReady, sbrOutput3.Status.Conditions)
	requireConditionPresentAndTrue(t, InjectionReady, sbrOutput3.Status.Conditions)

	require.Equal(t, reconcilerName, sbrOutput3.Status.Secret)
	require.Equal(t, s.Data["password"], []byte("abc123"))
	require.Equal(t, 1, len(sbrOutput3.Status.Applications))
	require.Equal(t, reconcilerName, sbrOutput3.Status.Secret)
	fetchSecret, err := fakeDynClient.Resource(secretsGVR).Namespace(updated.GetNamespace()).Get(reconcilerName, metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, fetchSecret.GetName(), reconcilerName)
	require.Equal(t, fetchSecret.GetResourceVersion(), "")
	require.Equal(t, 1, len(sbrOutput3.Status.Applications))
}

//TestReconcilerReconcileWithConflictingAppSelc tests when sbr has conflicting ApplicationSel such as MatchLabels=App1 and Name=App2 it should prioritise the Name
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
	f.AddMockedUnstructuredServiceBinding(reconcilerName, backingServiceResourceRef, applicationResourceRef2, deploymentsGVR, matchLabels1)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)
	f.AddMockedUnstructuredSecret("db-credentials")

	fakeDynClient := f.FakeDynClient()
	mapper := testutils.BuildTestRESTMapper()
	r := &reconciler{dynClient: fakeDynClient, restMapper: mapper, scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)

	t.Run("test-reconciler-reconcile-with-conflicting-application-selector", func(t *testing.T) {

		res, err := r.Reconcile(reconcileRequest())
		require.NoError(t, err)
		require.False(t, res.Requeue)

		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbrOutput, err := r.getServiceBinding(namespacedName)
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

		requireConditionPresentAndTrue(t, CollectionReady, sbrOutput.Status.Conditions)
		requireConditionPresentAndTrue(t, InjectionReady, sbrOutput.Status.Conditions)

		require.Equal(t, reconcilerName, sbrOutput.Status.Secret)
		require.Len(t, sbrOutput.Status.Applications, 1)
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))
	})
}

// TestEmptyApplication tests that Status is successfully updated when Application is missing
func TestEmptyApplication(t *testing.T) {
	backingServiceResourceRef := "backingService1"
	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredServiceBindingWithoutApplication(reconcilerName, backingServiceResourceRef)
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef)

	fakeDynClient := f.FakeDynClient()
	mapper := testutils.BuildTestRESTMapper()
	r := &reconciler{dynClient: fakeDynClient, restMapper: mapper, scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)

	res, err := r.Reconcile(reconcileRequest())
	require.NoError(t, err)
	require.False(t, res.Requeue)

	namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
	sbrOutput, err := r.getServiceBinding(namespacedName)
	require.NoError(t, err)

	requireConditionPresentAndTrue(t, CollectionReady, sbrOutput.Status.Conditions)
	requireConditionPresentAndFalse(t, InjectionReady, sbrOutput.Status.Conditions)
}

func TestBindTwoSbrsWithSingleApplication(t *testing.T) {
	applicationResourceRef := "applicationResourceRef"
	backingServiceResourceRef1 := "backingServiceRef1"
	backingServiceResourceRef2 := "backingServiceRef2"

	sbrName1 := "binding-request1"
	sbrName2 := "binding-request2"

	f := mocks.NewFake(t, reconcilerNs)
	f.AddMockedUnstructuredDeployment(applicationResourceRef, nil)
	f.AddMockedUnstructuredServiceBinding(sbrName1, backingServiceResourceRef1, applicationResourceRef, deploymentsGVR, nil)
	f.AddMockedUnstructuredServiceBinding(sbrName2, backingServiceResourceRef2, applicationResourceRef, deploymentsGVR, nil)
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef1)
	f.AddMockedUnstructuredDatabaseCR(backingServiceResourceRef2)

	fakeDynClient := f.FakeDynClient()
	mapper := testutils.BuildTestRESTMapper()
	r := &reconciler{dynClient: fakeDynClient, restMapper: mapper, scheme: f.S}
	r.resourceWatcher = newFakeResourceWatcher(mapper)

	t.Run("test-bind-two-sbrs-with-single-application", func(t *testing.T) {
		// Reconciling first sbr
		// First sbr reconcile request
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: reconcilerNs,
				Name:      sbrName1,
			},
		}

		// Reconcile first sbr
		res, err := r.Reconcile(req)
		require.NoError(t, err)
		require.False(t, res.Requeue)

		// Get sbr after reconciling
		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: sbrName1}
		sbrOutput, err := r.getServiceBinding(namespacedName)
		require.NoError(t, err)

		// expected sbr assertion
		expectedStatus := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: applicationResourceRef,
			},
		}

		requireConditionPresentAndTrue(t, CollectionReady, sbrOutput.Status.Conditions)
		requireConditionPresentAndTrue(t, InjectionReady, sbrOutput.Status.Conditions)
		require.Equal(t, sbrName1, sbrOutput.Status.Secret)
		require.Len(t, sbrOutput.Status.Applications, 1)
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))

		// Reconciling second sbr
		// Second sbr reconcile request
		req = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: reconcilerNs,
				Name:      sbrName2,
			},
		}

		// Reconcile second sbr
		res, err = r.Reconcile(req)
		require.NoError(t, err)
		require.False(t, res.Requeue)

		// Get sbr after reconciling
		namespacedName = types.NamespacedName{Namespace: reconcilerNs, Name: sbrName2}
		sbrOutput, err = r.getServiceBinding(namespacedName)
		require.NoError(t, err)

		// expected sbr assertion
		expectedStatus = v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   deploymentsGVR.Group,
				Version: deploymentsGVR.Version,
				Kind:    "Deployment",
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: applicationResourceRef,
			},
		}

		requireConditionPresentAndTrue(t, CollectionReady, sbrOutput.Status.Conditions)
		requireConditionPresentAndTrue(t, InjectionReady, sbrOutput.Status.Conditions)
		require.Equal(t, sbrName2, sbrOutput.Status.Secret)
		require.Len(t, sbrOutput.Status.Applications, 1)
		require.True(t, reflect.DeepEqual(expectedStatus, sbrOutput.Status.Applications[0]))

		// Get applicationResourceRef deployment
		resourceClient := r.dynClient.Resource(deploymentsGVR).Namespace(namespacedName.Namespace)
		u, err := resourceClient.Get(applicationResourceRef, metav1.GetOptions{})
		require.NoError(t, err)
		dep := &appsv1.Deployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, dep)
		require.NoError(t, err)

		// Assert SBR name with secretRef object in containers env var
		// Expected secretRef=binding-request1
		require.Equal(t, sbrName1, dep.Spec.Template.Spec.Containers[0].EnvFrom[0].SecretRef.LocalObjectReference.Name)
		// Expected secretRef=binding-request2
		require.Equal(t, sbrName2, dep.Spec.Template.Spec.Containers[0].EnvFrom[1].SecretRef.LocalObjectReference.Name)
	})
}
