package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

var secretOwnerReference = v1.OwnerReference{
	Name:       "binding-request",
	UID:        "c77ca1ae-72d0-4fdd-809f-58fdd37facf3",
	Kind:       "ServiceBindingRequest",
	APIVersion: "apps.openshift.io/v1alpha1",
}

func assertSecretNamespacedName(t *testing.T, u *unstructured.Unstructured, ns, name string) {
	assert.Equal(t, ns, u.GetNamespace())
	assert.Equal(t, name, u.GetName())
	ownerReference := u.GetOwnerReferences()
	assert.Equal(t, secretOwnerReference, ownerReference[0])
}

func TestSecretNew(t *testing.T) {
	ns := "secret"
	name := "test-secret"

	f := mocks.NewFake(t, ns)

	data := map[string][]byte{"key": []byte("value")}

	s := newSecret(
		f.FakeDynClient(),
		ns,
		name,
	)

	t.Run("createOrUpdate", func(t *testing.T) {
		u, err := s.createOrUpdate(data, secretOwnerReference)
		assert.NoError(t, err)
		assertSecretNamespacedName(t, u, ns, name)
	})

	t.Run("Get", func(t *testing.T) {
		u, found, err := s.get()
		assert.NoError(t, err)
		assert.True(t, found)
		assertSecretNamespacedName(t, u, ns, name)
	})
}
