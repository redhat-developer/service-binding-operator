package controllers

import (
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestBuildServiceContexts(t *testing.T) {
	logger := log.NewLog("testBuildServiceContexts")
	restMapper := testutils.BuildTestRESTMapper()
	falseBool := false

	t.Run("empty selectors", func(t *testing.T) {
		ns := "planner"
		f := mocks.NewFake(t, ns)
		serviceCtxs, err := buildServiceContexts(
			logger, f.FakeDynClient(), ns, nil, &falseBool, restMapper)

		require.NoError(t, err, "buildServiceContexts must execute without errors")
		require.Empty(t, serviceCtxs, "buildServiceContexts must be empty")
	})

	t.Run("declared and existing selectors", func(t *testing.T) {
		sbrName := "service-binding"
		firstResourceRef := "db-testing"
		firstNamespace := "existing-namespace"
		matchLabels := map[string]string{
			"connects-to": "database",
			"environment": "planner",
		}

		f := mocks.NewFake(t, firstNamespace)
		f.AddMockedUnstructuredCSV("cluster-service-version")
		f.AddMockedDatabaseCR(firstResourceRef, firstNamespace)
		f.AddMockedUnstructuredDatabaseCRD()
		f.AddNamespacedMockedSecret("db-credentials", firstNamespace, nil)

		sbr := f.AddMockedServiceBinding(sbrName, nil, firstResourceRef, "", deploymentsGVR, matchLabels)

		serviceCtxs, err := buildServiceContexts(
			logger, f.FakeDynClient(), firstNamespace, sbr.Spec.Services, &falseBool, restMapper)

		require.NoError(t, err, "buildServiceContexts must execute without errors")
		require.Len(t, serviceCtxs, 1, "buildServiceContexts must return only one item")

		serviceCtx := serviceCtxs[0]
		expectedKeys := []string{"status", "dbCredentials"}
		expectedDbCredentials := map[string]interface{}{
			"username": "user",
			"password": "password",
		}

		gotDbCredentials, ok, err :=
			unstructured.NestedFieldCopy(serviceCtx.service.Object, expectedKeys...)
		require.NoError(t, err, "must not return error while copying status.dbCredentials out of the context's service")
		require.True(t, ok, "status.dbCredentials must exist")
		require.Equal(t, expectedDbCredentials, gotDbCredentials, "status.dbCredentials in context must be equal to expected")
	})

	t.Run("services in different namespace", func(t *testing.T) {
		sameNs := "same-ns"
		sameNsResourceRef := "same-ns-database"

		otherNs := "other-ns"
		otherNsResourceRef := "other-ns-database"

		matchLabels := map[string]string{
			"connects-to": "database",
			"environment": "planner",
		}

		f := mocks.NewFake(t, sameNs)
		f.AddMockedUnstructuredDatabaseCRD()
		f.AddMockedDatabaseCR(sameNsResourceRef, sameNs)
		f.AddNamespacedMockedSecret("db-credentials", sameNs, map[string][]byte{
			"username": []byte("same-ns-username"),
			"password": []byte("same-ns-password"),
		})

		f.AddMockedDatabaseCR(otherNsResourceRef, otherNs)
		f.AddNamespacedMockedSecret("db-credentials", otherNs, map[string][]byte{
			"username": []byte("other-ns-username"),
			"password": []byte("other-ns-password"),
		})

		sbrName := "services-in-different-ns"

		sbr := f.AddMockedServiceBinding(sbrName, &sameNs, sameNsResourceRef, "", deploymentsGVR, matchLabels)
		sbr.Spec.Services = []v1alpha1.Service{
			{
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   mocks.CRDName,
					Version: mocks.CRDVersion,
					Kind:    mocks.CRDKind,
				},

				LocalObjectReference: corev1.LocalObjectReference{Name: otherNsResourceRef},
				Namespace:            &otherNs,
			},
			{
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   mocks.CRDName,
					Version: mocks.CRDVersion,
					Kind:    mocks.CRDKind,
				},
				LocalObjectReference: corev1.LocalObjectReference{Name: sameNsResourceRef},
				Namespace:            &sameNs,
			},
		}

		serviceCtxs, err := buildServiceContexts(
			logger, f.FakeDynClient(), sameNs, sbr.Spec.Services, &falseBool, restMapper)

		require.NoError(t, err, "buildServiceContexts must execute without errors")
		require.Len(t, serviceCtxs, 2, "buildServiceContexts must return both service contexts")

		{
			otherNsCtx := serviceCtxs[0]
			expectedKeys := []string{"status", "dbCredentials"}
			expectedDbCredentials := map[string]interface{}{
				"username": "other-ns-username",
				"password": "other-ns-password",
			}
			gotDbCredentials, ok, err :=
				unstructured.NestedFieldCopy(otherNsCtx.service.Object, expectedKeys...)
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, expectedDbCredentials, gotDbCredentials)
		}

		{
			sameNsCtx := serviceCtxs[1]
			expectedKeys := []string{"status", "dbCredentials"}
			expectedDbCredentials := map[string]interface{}{
				"username": "same-ns-username",
				"password": "same-ns-password",
			}
			gotDbCredentials, ok, err :=
				unstructured.NestedFieldCopy(sameNsCtx.service.Object, expectedKeys...)
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, expectedDbCredentials, gotDbCredentials)
		}
	})
}

var trueBool = true

func TestFindOwnedResourcesCtxs_ConfigMap(t *testing.T) {
	ns := "planner"
	name := "service-binding"
	backendService := "db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "planner",
	}
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBinding(name, nil, backendService, "", deploymentsGVR, matchLabels)
	sbr.Spec.DetectBindingResources = &trueBool

	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(name, ns)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredSecret("db-credentials")

	cr := mocks.UnstructuredDatabaseCRMock("test", "test")
	reference := metav1.OwnerReference{
		APIVersion:         cr.GetAPIVersion(),
		Kind:               cr.GetKind(),
		Name:               cr.GetName(),
		UID:                cr.GetUID(),
		Controller:         &trueBool,
		BlockOwnerDeletion: &trueBool,
	}
	configMap, err := converter.ToUnstructured(mocks.ConfigMapMock("test", "test_database"))
	require.NoError(t, err)
	configMap.SetOwnerReferences([]metav1.OwnerReference{reference})
	route := mocks.RouteCRMock("test", "test")
	route.SetOwnerReferences([]metav1.OwnerReference{reference})
	f.S.AddKnownTypeWithName(cr.GroupVersionKind(), &unstructured.Unstructured{})
	f.AddMockResource(cr)
	f.AddMockResource(configMap)
	f.AddMockResource(route)
	logger := log.NewLog("testFindOwnedResourcesCtxs_cm")

	restMapper := testutils.BuildTestRESTMapper()

	t.Run("existing selectors", func(t *testing.T) {
		got, err := findOwnedResourcesCtxs(
			logger,
			f.FakeDynClient(),
			cr.GetNamespace(),
			cr.GetName(),
			cr.GetUID(),
			cr.GroupVersionKind(),
			nil,
			restMapper,
		)
		require.NoError(t, err)
		require.Len(t, got, 2)

		expected := map[string]interface{}{
			"": map[string]interface{}{
				"password": "password",
				"username": "user",
			},
		}
		require.Equal(t, expected, got[0].envVars)

	})
}

func TestFindOwnedResourcesCtxs_Secrets(t *testing.T) {
	testCases := []struct {
		desc    string
		secrets []string
	}{
		{
			desc:    "backend cr creating only one secret should returns only one child",
			secrets: []string{"test_database"},
		},
		{
			desc:    "backend cr creating multiple secrets should returns only multiple children",
			secrets: []string{"test_database", "test_database2"},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			f := mocks.NewFake(t, "test")
			cr := mocks.UnstructuredDatabaseCRMock("test", "test")
			reference := metav1.OwnerReference{
				APIVersion:         cr.GetAPIVersion(),
				Kind:               cr.GetKind(),
				Name:               cr.GetName(),
				UID:                cr.GetUID(),
				Controller:         &trueBool,
				BlockOwnerDeletion: &trueBool,
			}
			f.S.AddKnownTypeWithName(cr.GroupVersionKind(), &unstructured.Unstructured{})
			f.AddMockResource(cr)
			logger := log.NewLog("testFindOwnedResourcesCtxs_secret")

			restMapper := testutils.BuildTestRESTMapper()

			for _, secret := range tC.secrets {
				secret, err := mocks.UnstructuredSecretMock("test", secret)
				require.NoError(t, err)
				secret.SetOwnerReferences([]metav1.OwnerReference{reference})
				f.AddMockResource(secret)
			}

			ownedResourcesCtxs, err := findOwnedResourcesCtxs(
				logger,
				f.FakeDynClient(),
				cr.GetNamespace(),
				cr.GetName(),
				cr.GetUID(),
				cr.GroupVersionKind(),
				nil,
				restMapper,
			)
			require.NoError(t, err)
			require.NotEmpty(t, ownedResourcesCtxs)
			require.EqualValues(t, len(tC.secrets), len(ownedResourcesCtxs))
			for idx, resource := range ownedResourcesCtxs {
				require.Equal(t, tC.secrets[idx], resource.service.GetName())
			}
		})
	}
}
