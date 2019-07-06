package servicebindingrequest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

var retriever *Retriever

func TestRetrieverNew(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	ns := "testing"
	crdName := "db-testing"

	crdDescription := mocks.CRDDescriptionMock()
	crd := mocks.DatabaseCRDMock(ns, crdName)

	genericCRDObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&crd)
	require.Nil(t, err)

	plan := &Plan{
		Ns:             ns,
		Name:           "retriever",
		CRDDescription: &crdDescription,
		CRD:            &ustrv1.Unstructured{Object: genericCRDObj},
	}

	dbSecret := mocks.SecretMock(ns, "db-credentials")
	objs := []runtime.Object{&dbSecret}
	fakeClient := fake.NewFakeClient(objs...)

	retriever = NewRetriever(context.TODO(), fakeClient, plan)
	require.NotNil(t, retriever)
}

func TestRetrieverGetCRDKey(t *testing.T) {
	imageName, err := retriever.getCRDKey("spec", "imageName")
	assert.Nil(t, err)
	assert.Equal(t, "postgres", imageName)
}

func TestRetrieverRead(t *testing.T) {
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

}

func TestRetrieverExtractSecretItemName(t *testing.T) {
	assert.Equal(t, "user", retriever.extractSecretItemName(
		"urn:alm:descriptor:servicebindingrequest:env:object:secret:user"))
}

func TestRetrieverReadSecret(t *testing.T) {
	retriever.data = make(map[string][]byte)

	err := retriever.readSecret("db-credentials", []string{"user", "password"})
	assert.Nil(t, err)

	assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_USER")
	assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_SECRET_PASSWORD")
}

func TestRetrieverStore(t *testing.T) {
	retriever.store("test", []byte("test"))
	assert.Contains(t, retriever.data, "SERVICE_BINDING_DATABASE_TEST")
	assert.Equal(t, []byte("test"), retriever.data["SERVICE_BINDING_DATABASE_TEST"])
}

func TestRetrieverSaveDataOnSecret(t *testing.T) {
	err := retriever.saveDataOnSecret()
	assert.Nil(t, err)
}

func TestRetrieverRetrieve(t *testing.T) {
	err := retriever.Retrieve()
	assert.Nil(t, err)
	assert.NotEmpty(t, retriever.data)
}
