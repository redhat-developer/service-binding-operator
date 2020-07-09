package servicebindingrequest

import (
	"encoding/base64"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// secret represents the data collected by this operator, and later handled as a secret.
type secret struct {
	logger *log.Log          // logger instance
	client dynamic.Interface // Kubernetes API client
	ns     string
	name   string
}

// buildResourceClient creates a resource client to handle corev1/secret resource.
func (s *secret) buildResourceClient() dynamic.ResourceInterface {
	gvr := corev1.SchemeGroupVersion.WithResource(secretResource)
	return s.client.Resource(gvr).Namespace(s.ns)
}

// createOrUpdate will take informed payload and either create a new secret or update an existing
// one. It can return error when Kubernetes client does.
func (s *secret) createOrUpdate(payload map[string][]byte, ownerReference metav1.OwnerReference) (*unstructured.Unstructured, error) {
	logger := s.logger.WithValues("Namespace", s.ns, "Name", s.name)
	secretObj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       s.ns,
			Name:            s.name,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: payload,
	}

	gvk := corev1.SchemeGroupVersion.WithKind(secretKind)
	u, err := converter.ToUnstructuredAsGVK(secretObj, gvk)
	if err != nil {
		return nil, err
	}

	resourceClient := s.buildResourceClient()

	logger.Debug("Attempt to create secret...")
	existingSecret, err := resourceClient.Get(s.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err := resourceClient.Create(u, metav1.CreateOptions{})
			return u, err
		}
		return nil, err
	}
	existingSecretData, _, _ := unstructured.NestedMap(existingSecret.Object, "data")
	payloadStr := make(map[string]string)
	for k, v := range payload {
		payloadStr[k] = base64.StdEncoding.EncodeToString(v)
	}
	payloadInterim := make(map[string]interface{})
	for k, v := range payloadStr {
		payloadInterim[k] = reflect.ValueOf(v).Interface()
	}
	eq := nestedMapComparison(existingSecretData, payloadInterim)
	if eq {
		logger.Debug("Secret data is same. Skip Update")
	} else {
		logger.Debug("Secret data is different. Update Secret")
		_, err = resourceClient.Update(u, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
	}
	return u, nil
}

// get an unstructured object from the secret handled by this component. It can return errors in case
// the API server does.
func (s *secret) get() (*unstructured.Unstructured, bool, error) {
	resourceClient := s.buildResourceClient()
	u, err := resourceClient.Get(s.name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, false, err
	}
	return u, u != nil, nil
}

// newSecret instantiate a new Secret.
func newSecret(
	client dynamic.Interface,
	ns string,
	name string,
) *secret {
	return &secret{
		logger: log.NewLog("secret"),
		client: client,

		name: name,
		ns:   ns,
	}
}
