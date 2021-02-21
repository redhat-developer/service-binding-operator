package secret

import (
	"context"
	"fmt"
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var ownerRefController bool = true
var secretOwnerReference = v1.OwnerReference{
	Name:       "binding-request",
	UID:        "c77ca1ae-72d0-4fdd-809f-58fdd37facf3",
	Kind:       "ServiceBinding",
	APIVersion: "operators.coreos.com/v1alpha1",
	Controller: &ownerRefController,
}

var namespace string = "secret-namespace"

var secretsGVR = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}

func assertSecretMetadata(t *testing.T, secret *corev1.Secret) {
	assert.Equal(t, namespace, secret.GetNamespace())
	ownerReference := secret.GetOwnerReferences()
	assert.Equal(t, secretOwnerReference, ownerReference[0])
}

func isSecretPresent(t *testing.T, f *fake.FakeDynamicClient, secret *corev1.Secret, len int) {
	secretList, err := f.Resource(secretsGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	assert.NoError(t, err)

	// Ensure length of List
	assert.Len(t, secretList.Items, len)

	// Ensure secret is present in List
	s, err := converter.ToUnstructured(secret)
	assert.NoError(t, err)
	assert.Contains(t, secretList.Items, *s)
}

// TestSecretWriteWithSamePrefixAndData tests secret write() and returns existing secret if both prefix and data of existing and current secrets are same
func TestSecretWriteWithSamePrefixAndData(t *testing.T) {
	f := mocks.NewFake(t, namespace)
	fakeClient := f.FakeDynClient()
	resourceClient := buildResourceClient(fakeClient, namespace)

	prefix := "prefix-1"
	data := map[string][]byte{"key1": []byte("value1")}
	secretHash, err := buildSecretHash(data)
	assert.NoError(t, err)

	readSecret, err := getServiceBindingSecret(resourceClient, prefix+"-"+secretHash, namespace)
	// read non existing secret
	assert.EqualError(t, err, fmt.Sprintf("secrets \"%s\" not found", prefix+"-"+secretHash))
	assert.Nil(t, readSecret)

	secret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		prefix,
		data,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, secret)
	assert.Equal(t, secret.GetName(), prefix+"-"+secretHash)

	// Return existing secret if both prefix and data are same
	existingSecret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		prefix,
		data,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, existingSecret)

	// Secret name should be same if both prefixes and data are same for two consecutive invocations
	assert.Equal(t, secret.GetName(), existingSecret.GetName())

	// Ensure only one secret is present
	isSecretPresent(t, fakeClient, secret, 1)
}

// TestSecretWriteWithDifferentPrefixSameData tests secret write() and creates new secret if prefix of current secret is different
func TestSecretWriteWithDifferentPrefixSameData(t *testing.T) {
	f := mocks.NewFake(t, namespace)
	fakeClient := f.FakeDynClient()

	prefix := "prefix-1"
	data := map[string][]byte{"key1": []byte("value1")}
	secretHash, err := buildSecretHash(data)
	assert.NoError(t, err)

	secret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		prefix,
		data,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, secret)
	assert.Equal(t, secret.GetName(), prefix+"-"+secretHash)

	prefix2 := "prefix-2"
	differentPrefixSecret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		prefix2,
		data,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, differentPrefixSecret)
	assert.NotEqual(t, differentPrefixSecret.GetName(), secret.GetName())
	assert.Equal(t, differentPrefixSecret.GetName(), prefix2+"-"+secretHash)

	// Ensure two secrets are present
	isSecretPresent(t, fakeClient, differentPrefixSecret, 2)
}

// TestSecretWriteWithSamePrefixDifferentData tests secret write() and creates new secret if data of current secret is different
func TestSecretWriteWithSamePrefixDifferentData(t *testing.T) {
	f := mocks.NewFake(t, namespace)
	fakeClient := f.FakeDynClient()

	prefix := "prefix-1"
	data := map[string][]byte{"key1": []byte("value1")}
	secretHash, err := buildSecretHash(data)
	assert.NoError(t, err)

	secret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		prefix,
		data,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, secret)
	assert.Equal(t, secret.GetName(), prefix+"-"+secretHash)

	differentData := map[string][]byte{"differentKey": []byte("differentValue")}
	differentSecretHash, err := buildSecretHash(differentData)
	assert.NoError(t, err)

	differentDataSecret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		prefix,
		differentData,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, differentDataSecret)
	assert.NotEqual(t, differentDataSecret.GetName(), secret.GetName())
	assert.Equal(t, differentDataSecret.GetName(), prefix+"-"+differentSecretHash)

	// Ensure two secrets are present
	isSecretPresent(t, fakeClient, differentDataSecret, 2)
}

// TestSecretUpdatedExternally tests if binding secret is repopulated on external update
func TestSecretUpdatedExternally(t *testing.T) {
	f := mocks.NewFake(t, namespace)
	fakeClient := f.FakeDynClient()
	resourceClient := buildResourceClient(fakeClient, namespace)

	bindingName := "binding"
	bindingSecretData := map[string][]byte{"key1": []byte("value1")}
	bindingSecretHash, err := buildSecretHash(bindingSecretData)
	assert.NoError(t, err)

	bindingSecret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		bindingName,
		bindingSecretData,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, bindingSecret)
	assert.Equal(t, bindingSecret.GetName(), bindingName+"-"+bindingSecretHash)

	bindingSecret.Data = map[string][]byte{"externallyUpdatedKey": []byte("externallyUpdatedValue")}
	unstructuredSecret, err := runtime.DefaultUnstructuredConverter.ToUnstructured(bindingSecret)
	require.NoError(t, err)

	updated := unstructured.Unstructured{Object: unstructuredSecret}
	externallyUpdatedSecret, err := fakeClient.Resource(secretsGVR).Namespace(updated.GetNamespace()).Update(context.TODO(), &updated, metav1.UpdateOptions{})
	assert.NoError(t, err)

	updatedSecretHash, err := buildSecretHash(bindingSecret.Data)
	assert.NoError(t, err)

	assert.Equal(t, externallyUpdatedSecret.GetName(), bindingSecret.GetName())
	// Externally updated secret name will not contain the new updated secret hash
	assert.NotContains(t, externallyUpdatedSecret.GetName(), updatedSecretHash)
	assert.Contains(t, externallyUpdatedSecret.GetName(), bindingSecretHash)

	correctedSecret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		bindingName,
		bindingSecretData,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, correctedSecret)

	assert.Equal(t, correctedSecret.GetName(), bindingSecret.GetName())

	assert.Equal(t, externallyUpdatedSecret.GetName(), correctedSecret.GetName())
	assert.NotEqual(t, externallyUpdatedSecret.Object, correctedSecret.Data)

	updatedSecretWithCorrectData, err := getServiceBindingSecret(resourceClient, correctedSecret.GetName(), namespace)
	assert.NoError(t, err)

	assert.Equal(t, updatedSecretWithCorrectData.GetName(), correctedSecret.GetName())

	assert.Equal(t, updatedSecretWithCorrectData, correctedSecret)
	isSecretPresent(t, fakeClient, correctedSecret, 1)
}

// TestKeysSortedDifferently tests secret write() and creates new secret if keys are sorted differently
func TestKeysSortedDifferently(t *testing.T) {
	f := mocks.NewFake(t, namespace)
	fakeClient := f.FakeDynClient()

	prefix := "binding"
	data := map[string][]byte{"sameKey": []byte("sameValue")}
	secretHash, err := buildSecretHash(data)
	assert.NoError(t, err)

	secret, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		prefix,
		data,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, secret)
	assert.Equal(t, secret.GetName(), prefix+"-"+secretHash)
	isSecretPresent(t, fakeClient, secret, 1)

	dataDifferentKeySort := map[string][]byte{"sameKey": []byte("differentValue")}
	dataDifferentKeySortHash, err := buildSecretHash(dataDifferentKeySort)
	assert.NoError(t, err)

	secretDifferentSort, err := WriteServiceBindingSecret(
		fakeClient,
		namespace,
		prefix,
		dataDifferentKeySort,
		secretOwnerReference,
	)
	assert.NoError(t, err)
	assertSecretMetadata(t, secretDifferentSort)
	assert.Equal(t, secretDifferentSort.GetName(), prefix+"-"+dataDifferentKeySortHash)
	isSecretPresent(t, fakeClient, secretDifferentSort, 2)
}
