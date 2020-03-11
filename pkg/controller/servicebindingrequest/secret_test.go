package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func assertNamespacedName(t *testing.T, u *unstructured.Unstructured, ns, name string) {
	assert.Equal(t, ns, u.GetNamespace())
	assert.Equal(t, name, u.GetName())
}

func TestSecretNew(t *testing.T) {
	ns := "my-secret"
	name := "test-secret"

	f := mocks.NewFake(t, ns)

	matchLabels := map[string]string{}
	sbr := mocks.ServiceBindingRequestMock(ns, name, nil, "", "", deploymentsGVR, matchLabels)

	plan := &Plan{
		Ns:   ns,
		Name: name,
		SBR:  *sbr,
	}
	data := map[string][]byte{"key": []byte("value")}

	s := NewBindingDataHandler(f.FakeDynClient(), plan)

	t.Run("customEnvParser", func(t *testing.T) {
		customData, err := s.customEnvParser(data)
		assert.NoError(t, err)
		assert.NotNil(t, customData)
	})

	t.Run("createOrUpdate", func(t *testing.T) {
		u, err := s.createOrUpdate(data)
		assert.NoError(t, err)
		assertNamespacedName(t, u, ns, name)
		assert.Equal(t, SecretKind, u.GetKind())
	})

	t.Run("Delete", func(t *testing.T) {
		err := s.Delete()
		assert.NoError(t, err)
	})

	t.Run("Commit", func(t *testing.T) {
		u, err := s.Commit(data)
		assert.NoError(t, err)
		assertNamespacedName(t, u, ns, name)
	})

	t.Run("Get", func(t *testing.T) {
		u, found, err := s.Get()
		assert.NoError(t, err)
		assert.True(t, found)
		assertNamespacedName(t, u, ns, name)
	})

	// if BindingReference is nil, the createOrUpdate(..)
	// ensures that we default to "Secret"

	sbr.Spec.BindingReference = nil

	plan = &Plan{
		Ns:   ns,
		Name: name,
		SBR:  *sbr,
	}

	s = NewBindingDataHandler(f.FakeDynClient(), plan)
	t.Run("createOrUpdate", func(t *testing.T) {
		u, err := s.createOrUpdate(data)
		assert.NoError(t, err)
		assertNamespacedName(t, u, ns, name)
		assert.Equal(t, SecretKind, u.GetKind())
	})
}

func TestConfigMapNew(t *testing.T) {
	ns := "my-configmap"
	name := "test-configmap"

	f := mocks.NewFake(t, ns)

	matchLabels := map[string]string{}
	sbr := mocks.ServiceBindingRequestMock(ns, name, nil, "", "", deploymentsGVR, matchLabels)
	sbr.Spec.BindingReference = &v1alpha1.BindingReference{
		ObjectType: metav1.GroupVersionKind{
			Version: "v1",
			Kind:    "ConfigMap",
		},
	}

	plan := &Plan{
		Ns:   ns,
		Name: name,
		SBR:  *sbr,
	}
	data := map[string][]byte{"key": []byte("value")}

	s := NewBindingDataHandler(f.FakeDynClient(), plan)

	t.Run("customEnvParser", func(t *testing.T) {
		customData, err := s.customEnvParser(data)
		assert.NoError(t, err)
		assert.NotNil(t, customData)
	})

	t.Run("createOrUpdate", func(t *testing.T) {
		u, err := s.createOrUpdate(data)
		assert.NoError(t, err)
		assertNamespacedName(t, u, ns, name)
		assert.Equal(t, ConfigMapKind, u.GetKind())
	})

	t.Run("Delete", func(t *testing.T) {
		err := s.Delete()
		assert.NoError(t, err)
	})

	t.Run("Commit", func(t *testing.T) {
		u, err := s.Commit(data)
		assert.NoError(t, err)
		assertNamespacedName(t, u, ns, name)
	})

	t.Run("Get", func(t *testing.T) {
		u, found, err := s.Get()
		assert.NoError(t, err)
		assert.True(t, found)
		assertNamespacedName(t, u, ns, name)
	})
}
