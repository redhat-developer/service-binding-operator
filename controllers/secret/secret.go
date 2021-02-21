package secret

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
)

const (
	// secretResource defines the resource name for Secrets.
	secretResource = "secrets"
	// secretKind defines the name of Secret kind.
	secretKind = "Secret"
)

// buildResourceClient creates a resource client to handle corev1/secret resource.
func buildResourceClient(client dynamic.Interface, namespace string) dynamic.ResourceInterface {
	gvr := corev1.SchemeGroupVersion.WithResource(secretResource)
	return client.Resource(gvr).Namespace(namespace)
}

// buildSecretHash is a utility function to get a checksum of the resource data
func buildSecretHash(data map[string][]byte) (string, error) {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	hash := sha1.New()
	for _, k := range keys {
		_, err := hash.Write([]byte(k))
		if err != nil {
			return "", err
		}
		_, err = hash.Write(data[k])
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// WriteServiceBindingSecret will take informed payload and create a new secret
// It can return error when Kubernetes client does.
func WriteServiceBindingSecret(
	client dynamic.Interface,
	ns string,
	prefix string,
	payload map[string][]byte,
	ownerReference metav1.OwnerReference,
) (*corev1.Secret, error) {
	secretHash, err := buildSecretHash(payload)
	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ns,
			Name:            prefix + "-" + secretHash,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: payload,
	}

	gvk := corev1.SchemeGroupVersion.WithKind(secretKind)
	u, err := converter.ToUnstructuredAsGVK(secret, gvk)
	if err != nil {
		return nil, err
	}

	resourceClient := buildResourceClient(client, ns)
	expectedSecret, err := getServiceBindingSecret(resourceClient, secret.Name, ns)
	if err != nil {
		if errors.IsNotFound(err) {
			secretNew, err := resourceClient.Create(context.TODO(), u, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
			s := &corev1.Secret{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(secretNew.Object, s)
			if err != nil {
				return nil, err
			}
			return s, nil
		}
		return nil, err
	}

	expectedDataHash, err := buildSecretHash(expectedSecret.Data)
	if err != nil {
		return nil, err
	}

	// compare current and existing secret name and data hash
	if expectedSecret.GetName() != secret.Name || expectedDataHash != secretHash {
		secretNew, err := resourceClient.Update(context.TODO(), u, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
		s := &corev1.Secret{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(secretNew.Object, s)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
	return secret, nil
}

// getServiceBindingSecret an unstructured object from the secret handled by this component. It can return errors in case
// the API server does.
func getServiceBindingSecret(resourceClient dynamic.ResourceInterface, name, namespace string) (*corev1.Secret, error) {
	u, err := resourceClient.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	s := &corev1.Secret{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
