package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func SecretMock(ns, name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: map[string][]byte{
			"user":     []byte("user"),
			"password": []byte("password"),
		},
	}
}

func TestUnstructuredToUnstructured(t *testing.T) {
	ns := "ns"
	name := "name"

	secret := SecretMock(ns, name)

	u, err := ToUnstructured(secret)
	assert.NoError(t, err)
	t.Logf("Unstructured: '%#v'", u)

	assert.Equal(t, ns, u.GetNamespace())
	assert.Equal(t, name, u.GetName())
}

func TestUnstructuredToUnstructuredAsGVK(t *testing.T) {
	ns := "ns"
	name := "name"

	secret := SecretMock(ns, name)
	gvk := schema.GroupVersion{Group: "", Version: "v1"}.WithKind("Secret")

	u, err := ToUnstructuredAsGVK(secret, gvk)
	assert.NoError(t, err)
	t.Logf("Unstructured: '%#v'", u)

	assert.Equal(t, "Secret", u.GetKind())
	assert.Equal(t, "v1", u.GetAPIVersion())
}

func TestValidNestedResources(t *testing.T) {
	resource1 := map[string]interface{}{
		"foo": []map[string]interface{}{
			{
				"name":  "bar",
				"value": "baz",
			},
		},
	}
	resource2 := map[string]interface{}{
		"foo": []interface{}{
			map[string]interface{}{
				"name":  "bar",
				"value": "baz",
			},
		},
	}

	result, found, err := NestedResources(&corev1.EnvVar{}, resource1, "foo")
	assert.Equal(t, true, found)
	assert.Equal(t, nil, err)
	assert.Equal(t, []map[string]interface{}{{"name": "bar", "value": "baz"}}, result)

	result, found, err = NestedResources(&corev1.EnvVar{}, resource2, "foo")
	assert.Equal(t, true, found)
	assert.Equal(t, nil, err)
	assert.Equal(t, []map[string]interface{}{{"name": "bar", "value": "baz"}}, result)
}

func TestInvalidNestedResources(t *testing.T) {
	resource1 := map[string]interface{}{
		"foo": map[string]interface{}{"name": "bar", "value": "baz"},
	}
	resource2 := map[string]interface{}{
		"foo": []map[string]interface{}{
			{
				"name":  "bar",
				"value": "baz",
			},
		},
	}
	resource3 := map[string]interface{}{"foo": "bar"}
	resource4 := map[string]interface{}{
		"foo": []interface{}{"bar", "baz"},
	}

	// not a slice
	result, found, err := NestedResources(&corev1.EnvVar{}, resource1, "foo")
	assert.Equal(t, true, found)
	assert.Nil(t, result)
	assert.NotNil(t, err)

	// nonexistant path
	result, found, err = NestedResources(&corev1.EnvVar{}, resource2, "bar")
	assert.Equal(t, false, found)
	assert.Nil(t, result)
	assert.Nil(t, err)

	// not a slice
	result, found, err = NestedResources(&corev1.EnvVar{}, resource3, "foo")
	assert.Equal(t, true, found)
	assert.Nil(t, result)
	assert.Error(t, err)

	// not a slice of maps
	result, found, err = NestedResources(&corev1.EnvVar{}, resource4, "foo")
	assert.Equal(t, true, found)
	assert.Nil(t, result)
	assert.Error(t, err)
}
