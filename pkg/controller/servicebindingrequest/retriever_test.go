package servicebindingrequest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/planner"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestRetriever(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	crdDescription := mocks.CRDDescriptionMock()
	cr := mocks.DatabaseCRMock(ns, crName)

	genericCR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&cr)
	require.Nil(t, err)

	plan := &planner.Plan{
		Ns:             ns,
		Name:           "retriever",
		CRDDescription: &crdDescription,
		CR:             &ustrv1.Unstructured{Object: genericCR},
	}

	dbSecret := mocks.SecretMock(ns, "db-credentials")
	objs := []runtime.Object{&dbSecret}
	fakeClient := fake.NewFakeClient(objs...)

	retriever = NewRetriever(context.TODO(), fakeClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("retrive", func(t *testing.T) {
		err := retriever.Retrieve()
		assert.Nil(t, err)
		assert.NotEmpty(t, retriever.data)
	})

	t.Run("getCRKey", func(t *testing.T) {
		imageName, err := retriever.getCRKey("spec", "imageName")
		assert.Nil(t, err)
		assert.Equal(t, "postgres", imageName)
	})

	t.Run("read", func(t *testing.T) {
		// reading from secret, from status attribute
		err := retriever.read("status", "dbCredentials", []string{
			"urn:alm:descriptor:servicebindingrequest:env:object:secret:user",
			"urn:alm:descriptor:servicebindingrequest:env:object:secret:password",
		})
		assert.Nil(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_USER")
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_PASSWORD")

		// reading from spec attribute
		err = retriever.read("spec", "image", []string{
			"urn:alm:descriptor:servicebindingrequest:env:attribute",
		})
		assert.Nil(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_IMAGE")

	})

	t.Run("extractSecretItemName", func(t *testing.T) {
		assert.Equal(t, "user", retriever.extractSecretItemName(
			"urn:alm:descriptor:servicebindingrequest:env:object:secret:user"))
	})

	t.Run("readSecret", func(t *testing.T) {
		retriever.data = make(map[string][]byte)

		err := retriever.readSecret("db-credentials", []string{"user", "password"})
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
}

func TestRetrieverNestedCRDKey(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	crdDescription := mocks.CRDDescriptionMock()
	cr := mocks.NestedDatabaseCRMock(ns, crName)

	genericCR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&cr)
	require.Nil(t, err)

	plan := &planner.Plan{
		Ns:             ns,
		Name:           "retriever",
		CRDDescription: &crdDescription,
		CR:             &ustrv1.Unstructured{Object: genericCR},
	}

	dbSecret := mocks.SecretMock(ns, "db-credentials")
	objs := []runtime.Object{&dbSecret}
	fakeClient := fake.NewFakeClient(objs...)

	retriever = NewRetriever(context.TODO(), fakeClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("Second level", func(t *testing.T) {
		imageName, err := retriever.getCRKey("spec", "image.name")
		assert.Nil(t, err)
		assert.Equal(t, "postgres", imageName)
	})

	t.Run("Second level error", func(t *testing.T) {
		_, err := retriever.getCRKey("spec", "image..name")
		assert.NotNil(t, err)
	})

	t.Run("Third level", func(t *testing.T) {
		something, err := retriever.getCRKey("spec", "image.third.something")
		assert.Nil(t, err)
		assert.Equal(t, "somevalue", something)
	})

}

func TestConfigMapRetriever(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))
	var retriever *Retriever

	ns := "testing"
	crName := "db-testing"

	crdDescription := mocks.CRDDescriptionConfigMapMock()

	cr := mocks.DatabaseConfigMapMock(ns, crName)

	genericCR, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&cr)
	require.Nil(t, err)

	plan := &planner.Plan{
		Ns:             ns,
		Name:           "retriever",
		CRDDescription: &crdDescription,
		CR:             &ustrv1.Unstructured{Object: genericCR},
	}

	dbConfigMap := mocks.ConfigMapMock(ns, "db-configmap")
	objs := []runtime.Object{&dbConfigMap}
	fakeClient := fake.NewFakeClient(objs...)

	retriever = NewRetriever(context.TODO(), fakeClient, plan, "SERVICE_BINDING")
	require.NotNil(t, retriever)

	t.Run("read", func(t *testing.T) {

		// reading from configMap, from status attribute
		err = retriever.read("spec", "dbConfigMap", []string{
			"urn:alm:descriptor:servicebindingrequest:env:object:configmap:user",
			"urn:alm:descriptor:servicebindingrequest:env:object:configmap:password",
		})
		assert.Nil(t, err)

		t.Logf("retriever.data '%#v'", retriever.data)
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_CONFIGMAP_USER")
		assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_CONFIGMAP_PASSWORD")
	})

	t.Run("extractConfigMapItemName", func(t *testing.T) {
		assert.Equal(t, "user", retriever.extractConfigMapItemName(
			"urn:alm:descriptor:servicebindingrequest:env:object:configmap:user"))
	})

	t.Run("readConfigMap", func(t *testing.T) {
		retriever.data = make(map[string][]byte)

		err := retriever.readConfigMap("db-configmap", []string{"user", "password"})
		assert.Nil(t, err)

		assert.Contains(t, retriever.data, ("SERVICE_BINDING_DATABASE_CONFIGMAP_USER"))
		assert.Contains(t, retriever.data, ("SERVICE_BINDING_DATABASE_CONFIGMAP_PASSWORD"))
	})

}
