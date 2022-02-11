package mocks

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
)

func UnstructuredSecretMock(ns, name string) (*unstructured.Unstructured, error) {
	s := SecretMock(ns, name, nil)
	return converter.ToUnstructured(&s)
}

func UnstructuredSecretMockRV(ns, name string) (*unstructured.Unstructured, error) {
	s := SecretMockRV(ns, name)
	return converter.ToUnstructured(&s)
}

//
// Usage of TypeMeta in Mocks
//
// 	Usually TypeMeta should not be explicitly defined in mocked objects, however, on using
//  it via *unstructured.Unstructured it could not find this CR without it.
//

// SecretMock returns a Secret based on PostgreSQL operator usage.
func SecretMock(ns, name string, data map[string][]byte) *corev1.Secret {
	if data == nil {
		data = map[string][]byte{
			"username": []byte("user"),
			"password": []byte("password"),
		}
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: data,
	}
}

// SecretMockRV returns a Secret with a resourceVersion.
func SecretMockRV(ns, name string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ns,
			Name:            name,
			ResourceVersion: "116076",
		},
		Data: map[string][]byte{
			"user":     []byte("user"),
			"password": []byte("password"),
		},
	}
}

// ConfigMapMock returns a dummy config-map object.
func ConfigMapMock(ns, name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: map[string]string{
			"username": "user",
			"password": "password",
		},
	}
}
