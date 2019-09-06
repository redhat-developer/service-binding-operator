package servicebindingrequest

import (
	"context"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Nil(t, err)

	plan := &Plan{Ns: ns, Name: "retriever", CRDDescription: &crdDescription, CR: cr}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(context.TODO(), fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("retrive", func(t *testing.T) {
		objs, err := retriever.Retrieve()
		assert.Nil(t, err)
		assert.NotEmpty(t, retriever.data)
		assert.True(t, len(objs) > 0)
	})

	t.Run("getCRKey", func(t *testing.T) {
		imageName, _, err := retriever.getCRKey("spec", "imageName")
		assert.Nil(t, err)
		assert.Equal(t, "postgres", imageName)
	})

	t.Run("read", func(t *testing.T) {
		// reading from secret, from status attribute
		err := retriever.read("status", "dbCredentials", []string{
			"binding:env:object:secret:user",
			"binding:env:object:secret:password",
		})
		assert.Nil(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_USER")
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_PASSWORD")

		// reading from spec attribute
		err = retriever.read("spec", "image", []string{
			"binding:env:attribute",
		})
		assert.Nil(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_IMAGE")

	})

	t.Run("extractSecretItemName", func(t *testing.T) {
		assert.Equal(t, "user", retriever.extractSecretItemName(
			"binding:env:object:secret:user"))
	})

	t.Run("readSecret", func(t *testing.T) {
		retriever.data = make(map[string][]byte)

		err := retriever.readSecret("db-credentials", []string{"user", "password"}, "spec", "dbConfigMap")
		assert.Nil(t, err)

		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_USER")
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_PASSWORD")
	})

	t.Run("store", func(t *testing.T) {
		retriever.store("test", []byte("test"))
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_TEST")
		assert.Equal(t, []byte("test"), retriever.data["SERVICE_BINDING_DATABASE_TEST"])
	})

	t.Run("saveDataOnSecret", func(t *testing.T) {
		err := retriever.saveDataOnSecret()
		assert.Nil(t, err)
	})

	t.Run("empty prefix", func(t *testing.T) {
		retriever = NewRetriever(context.TODO(), fakeDynClient, plan, "")
		require.NotNil(t, retriever)
		retriever.data = make(map[string][]byte)

		err := retriever.readSecret("db-credentials", []string{"user", "password"}, "spec", "dbConfigMap")
		assert.Nil(t, err)

		assert.Contains(t, retriever.data, "DATABASE_SECRET_USER")
		assert.Contains(t, retriever.data, "DATABASE_SECRET_PASSWORD")
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
	require.Nil(t, err)

	plan := &Plan{Ns: ns, Name: "retriever", CRDDescription: &crdDescription, CR: cr}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(context.TODO(), fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("Second level", func(t *testing.T) {
		imageName, _, err := retriever.getCRKey("spec", "image.name")
		assert.Nil(t, err)
		assert.Equal(t, "postgres", imageName)
	})

	t.Run("Second level error", func(t *testing.T) {
		_, _, err := retriever.getCRKey("spec", "image..name")
		assert.NotNil(t, err)
	})

	t.Run("Third level", func(t *testing.T) {
		something, _, err := retriever.getCRKey("spec", "image.third.something")
		assert.Nil(t, err)
		assert.Equal(t, "somevalue", something)
	})

}

func TestRetrieverWithConfigMap(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddMockedConfigMap(crName)
	f.AddMockedDatabaseCR(crName)

	crdDescription := mocks.CRDDescriptionConfigMapMock()

	cr, err := mocks.UnstructuredDatabaseConfigMapMock(ns, crName, crName)
	require.Nil(t, err)

	plan := &Plan{Ns: ns, Name: "retriever", CRDDescription: &crdDescription, CR: cr}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(context.TODO(), fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("read", func(t *testing.T) {
		// reading from configMap, from status attribute
		err = retriever.read("spec", "dbConfigMap", []string{
			"binding:env:object:configmap:user",
			"binding:env:object:configmap:password",
		})
		assert.Nil(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_CONFIGMAP_USER")
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_CONFIGMAP_PASSWORD")
	})

	t.Run("extractConfigMapItemName", func(t *testing.T) {
		assert.Equal(t, "user", retriever.extractConfigMapItemName(
			"binding:env:object:configmap:user"))
	})

	t.Run("readConfigMap", func(t *testing.T) {
		retriever.data = make(map[string][]byte)

		err := retriever.readConfigMap(crName, []string{"user", "password"}, "spec", "dbConfigMap")
		assert.Nil(t, err)

		assert.Contains(t, retriever.data, ("SERVICE_BINDING_DATABASE_CONFIGMAP_USER"))
		assert.Contains(t, retriever.data, ("SERVICE_BINDING_DATABASE_CONFIGMAP_PASSWORD"))
	})


}

func TestCustomEnvParser(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	f := mocks.NewFake(t, ns)
	f.AddMockedUnstructuredCSV("csv")
	f.AddMockedSecret("db-credentials")

	crdDescription := mocks.CRDDescriptionMock()
	cr, err := mocks.UnstructuredDatabaseCRMock(ns, crName)
	require.Nil(t, err)

	plan := &Plan{Ns: ns, Name: "retriever", CRDDescription: &crdDescription, CR: cr}

	fakeDynClient := f.FakeDynClient()

	retriever = NewRetriever(context.TODO(), fakeDynClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("Should detect custom env values", func(t *testing.T) {
		_, err = retriever.Retrieve()
		assert.Nil(t, err)

		t.Logf("\nCache %+v", retriever.cache)

		envMap := []v1alpha1.EnvMap{
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
		assert.Equal(t, "dXNlcg==_cGFzc3dvcmQ=", values["ANOTHER_STRING"], "Custom env values are not matching")
		assert.Equal(t, "postgres@cGFzc3dvcmQ=", values["JDBC_CONNECTION_STRING"], "Custom env values are not matching")
	})
}
