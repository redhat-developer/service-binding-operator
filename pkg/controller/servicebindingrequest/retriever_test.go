package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestRetriever(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddMockedSecret("db-credentials")

	crdDescription := mocks.CRDDescriptionMock()
	cr, err := mocks.UnstructuredDatabaseCRMock(ns, crName)
	require.NoError(t, err)

	plan := &Plan{Ns: ns, Name: "retriever", CRDDescription: &crdDescription, CR: cr}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("retrive", func(t *testing.T) {
		objs, _, err := retriever.Retrieve()
		require.NoError(t, err)
		require.NotEmpty(t, retriever.data)
		require.True(t, len(objs) > 0)
	})

	t.Run("getCRKey", func(t *testing.T) {
		imageName, _, err := retriever.getCRKey("spec", "imageName")
		require.NoError(t, err)
		require.Equal(t, "postgres", imageName)
	})

	t.Run("read", func(t *testing.T) {
		// reading from secret, from status attribute
		err := retriever.read("status", "dbCredentials", []string{
			"binding:env:object:secret:user",
			"binding:env:object:secret:password",
		})
		require.NoError(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		require.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_USER")
		require.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_PASSWORD")

		// reading from spec attribute
		err = retriever.read("spec", "image", []string{
			"binding:env:attribute",
		})
		require.NoError(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		require.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_IMAGE")

	})

	t.Run("extractSecretItemName", func(t *testing.T) {
		require.Equal(t, "user", retriever.extractSecretItemName(
			"binding:env:object:secret:user"))
	})

	t.Run("readSecret", func(t *testing.T) {
		retriever.data = make(map[string][]byte)

		err := retriever.readSecret("db-credentials", []string{"user", "password"}, "spec", "dbConfigMap")
		require.NoError(t, err)

		require.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_USER")
		require.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_PASSWORD")
	})

	t.Run("store", func(t *testing.T) {
		retriever.store("test", []byte("test"))
		require.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_TEST")
		require.Equal(t, []byte("test"), retriever.data["SERVICE_BINDING_DATABASE_TEST"])
	})

	t.Run("empty prefix", func(t *testing.T) {
		retriever = NewRetriever(fakeDynClient, plan, "")
		require.NotNil(t, retriever)
		retriever.data = make(map[string][]byte)

		err := retriever.readSecret("db-credentials", []string{"user", "password"}, "spec", "dbConfigMap")
		require.NoError(t, err)

		require.Contains(t, retriever.data, "DATABASE_SECRET_USER")
		require.Contains(t, retriever.data, "DATABASE_SECRET_PASSWORD")
	})
}

func TestRetrieverWithNestedCRKey(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddMockedSecret("db-credentials")

	crdDescription := mocks.CRDDescriptionMock()
	cr, err := mocks.UnstructuredNestedDatabaseCRMock(ns, crName)
	require.NoError(t, err)

	plan := &Plan{Ns: ns, Name: "retriever", CRDDescription: &crdDescription, CR: cr}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("Second level", func(t *testing.T) {
		imageName, _, err := retriever.getCRKey("spec", "image.name")
		require.NoError(t, err)
		require.Equal(t, "postgres", imageName)
	})

	t.Run("Second level error", func(t *testing.T) {
		// FIXME: if attribute isn't available in CR we would not throw any error.
		t.Skip()
		_, _, err := retriever.getCRKey("spec", "image..name")
		require.NotNil(t, err)
	})

	t.Run("Third level", func(t *testing.T) {
		something, _, err := retriever.getCRKey("spec", "image.third.something")
		require.NoError(t, err)
		require.Equal(t, "somevalue", something)
	})
}

func TestRetrieverWithConfigMap(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddMockedUnstructuredConfigMap(crName)
	f.AddMockedDatabaseCR(crName)

	crdDescription := mocks.CRDDescriptionConfigMapMock()

	cr, err := mocks.UnstructuredDatabaseConfigMapMock(ns, crName, crName)
	require.NoError(t, err)

	plan := &Plan{Ns: ns, Name: "retriever", CRDDescription: &crdDescription, CR: cr}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("read", func(t *testing.T) {
		// reading from configMap, from status attribute
		err = retriever.read("spec", "dbConfigMap", []string{
			"binding:env:object:configmap:user",
			"binding:env:object:configmap:password",
		})
		require.NoError(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		require.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_CONFIGMAP_USER")
		require.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_CONFIGMAP_PASSWORD")
	})
	t.Run("extractConfigMapItemName", func(t *testing.T) {
		require.Equal(t, "user", retriever.extractConfigMapItemName(
			"binding:env:object:configmap:user"))
	})

	t.Run("readConfigMap", func(t *testing.T) {
		retriever.data = make(map[string][]byte)

		err := retriever.readConfigMap(crName, []string{"user", "password"}, "spec", "dbConfigMap")
		require.NoError(t, err)

		require.Contains(t, retriever.data, ("SERVICE_BINDING_DATABASE_CONFIGMAP_USER"))
		require.Contains(t, retriever.data, ("SERVICE_BINDING_DATABASE_CONFIGMAP_PASSWORD"))
	})
}

func TestCustomEnvParser(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedSecret("db-credentials")

	crdDescription := mocks.CRDDescriptionMock()
	cr, err := mocks.UnstructuredDatabaseCRMock(ns, crName)
	require.NoError(t, err)

	plan := &Plan{Ns: ns, Name: "retriever", CRDDescription: &crdDescription, CR: cr}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("Should detect custom env values", func(t *testing.T) {
		_, _, err = retriever.Retrieve()
		require.NoError(t, err)

		envMap := []corev1.EnvVar{
			{
				Name:  "JDBC_CONNECTION_STRING",
				Value: `{{ .spec.imageName }}@{{ .status.dbCredentials.password }}`,
			},
			{
				Name:  "ANOTHER_STRING",
				Value: `{{ .status.dbCredentials.user }}_{{ .status.dbCredentials.password }}`,
			},
		}

		c := NewCustomEnvParser(envMap, retriever.cache)
		values, err := c.Parse()
		if err != nil {
			t.Error(err)
		}
		require.Equal(t, "user_password", values["ANOTHER_STRING"], "Custom env values are not matching")
		require.Equal(t, "postgres@password", values["JDBC_CONNECTION_STRING"], "Custom env values are not matching")
	})
}
