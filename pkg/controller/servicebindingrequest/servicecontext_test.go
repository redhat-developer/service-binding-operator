package servicebindingrequest

import (
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestBuildServiceContexts(t *testing.T) {
	logger := log.NewLog("testBuildServiceContexts")
	restMapper := testutils.BuildTestRESTMapper()

	t.Run("empty selectors", func(t *testing.T) {
		ns := "planner"
		f := mocks.NewFake(t, ns)
		serviceCtxs, err := buildServiceContexts(
			logger, f.FakeDynClient(), ns, nil, false, restMapper)

		require.NoError(t, err, "buildServiceContexts must execute without errors")
		require.Empty(t, serviceCtxs, "buildServiceContexts must be empty")
	})

	t.Run("declared and existing selectors", func(t *testing.T) {
		sbrName := "service-binding-request"
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

		sbr := f.AddMockedServiceBindingRequest(sbrName, nil, firstResourceRef, "", deploymentsGVR, matchLabels)

		serviceCtxs, err := buildServiceContexts(
			logger, f.FakeDynClient(), firstNamespace, extractServiceSelectors(sbr), false, restMapper)

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

		sbr := f.AddMockedServiceBindingRequest(sbrName, &sameNs, sameNsResourceRef, "", deploymentsGVR, matchLabels)
		sbr.Spec.BackingServiceSelectors = &[]v1alpha1.BackingServiceSelector{
			{
				GroupVersionKind: metav1.GroupVersionKind{
					Group:   mocks.CRDName,
					Version: mocks.CRDVersion,
					Kind:    mocks.CRDKind,
				},
				ResourceRef: otherNsResourceRef,
				Namespace:   &otherNs,
			},
		}

		serviceCtxs, err := buildServiceContexts(
			logger, f.FakeDynClient(), sameNs, extractServiceSelectors(sbr), false, restMapper)

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
	sbr.Spec.DetectBindingResources = trueBool

	f.AddMockedUnstructuredCSV("cluster-service-version")
	f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredSecret("db-credentials")

	cr := mocks.DatabaseCRMock("test", "test")
	reference := metav1.OwnerReference{
		APIVersion:         cr.APIVersion,
		Kind:               cr.Kind,
		Name:               cr.Name,
		UID:                cr.UID,
		Controller:         &trueBool,
		BlockOwnerDeletion: &trueBool,
	}
	configMap := mocks.ConfigMapMock("test", "test_database")
	us := &unstructured.Unstructured{}
	uc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&configMap)
	require.NoError(t, err)
	us.Object = uc
	us.SetOwnerReferences([]metav1.OwnerReference{reference})
	route, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mocks.RouteCRMock("test", "test"))
	require.NoError(t, err)
	usRoute := &unstructured.Unstructured{Object: route}
	usRoute.SetOwnerReferences([]metav1.OwnerReference{reference})
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.S.AddKnownTypes(routev1.SchemeGroupVersion, &routev1.Route{})
	f.AddMockResource(cr)
	f.AddMockResource(us)
	f.AddMockResource(&unstructured.Unstructured{Object: route})
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
		require.Len(t, got, 1)

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
			cr := mocks.DatabaseCRMock("test", "test")
			reference := metav1.OwnerReference{
				APIVersion:         cr.APIVersion,
				Kind:               cr.Kind,
				Name:               cr.Name,
				UID:                cr.UID,
				Controller:         &trueBool,
				BlockOwnerDeletion: &trueBool,
			}
			route, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mocks.RouteCRMock("test", "test"))
			require.NoError(t, err)
			usRoute := &unstructured.Unstructured{Object: route}
			usRoute.SetOwnerReferences([]metav1.OwnerReference{reference})
			f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
			f.S.AddKnownTypes(routev1.SchemeGroupVersion, &routev1.Route{})
			f.AddMockResource(cr)
			f.AddMockResource(&unstructured.Unstructured{Object: route})
			logger := log.NewLog("testFindOwnedResourcesCtxs_secret")

			restMapper := testutils.BuildTestRESTMapper()

			for _, secret := range tC.secrets {
				secret := mocks.SecretMock("test", secret, nil)
				us := &unstructured.Unstructured{}
				uc, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&secret)
				require.NoError(t, err)
				us.Object = uc
				us.SetOwnerReferences([]metav1.OwnerReference{reference})
				f.AddMockResource(us)
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
