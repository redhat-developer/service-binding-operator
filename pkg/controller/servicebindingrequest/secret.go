package servicebindingrequest

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// Secret represents the data collected by this operator, and later handled as a secret.
type Secret struct {
	logger *log.Log          // logger instance
	client dynamic.Interface // Kubernetes API client
	plan   *Plan             // plan instance
}

// customEnvParser parse informed data in order to interpolate with values provided by custom
// environment component.
func (s *Secret) customEnvParser(data map[string][]byte) (map[string][]byte, error) {
	// transforming input into format expected by custom environment parser
	cache := make(map[string]interface{})
	for k, v := range data {
		cache[k] = v
	}

	// interpolating custom environment
	envParser := NewCustomEnvParser(s.plan.SBR.Spec.CustomEnvVar, cache)
	values, err := envParser.Parse()
	if err != nil {
		return nil, err
	}

	for k, v := range values {
		data[k] = []byte(v.(string))
	}
	return data, nil
}

// buildResourceClient creates a resource client to handle corev1/secret resource.
func (s *Secret) buildResourceClient(resource string) dynamic.ResourceInterface {
	gvr := corev1.SchemeGroupVersion.WithResource(resource)
	return s.client.Resource(gvr).Namespace(s.plan.Ns)
}

// createOrUpdate will take informed payload and create/update the binding secret/configmap
// one.
// If the spec.BindingReference is nil,
// Returns an error when the Kubernetes client returns an error.
func (s *Secret) createOrUpdate(payload map[string][]byte) (*unstructured.Unstructured, error) {
	ns := s.plan.Ns
	name := s.plan.Name
	logger := s.logger.WithValues("Namespace", ns, "Name", name)

	var bindingRef interface{}

	// Default binding resource type is Secret
	resource := SecretResource
	gvk := corev1.SchemeGroupVersion.WithKind(SecretKind)

	// Only if otherwise specified as configmap, we'll use a binding
	// configmaps
	if s.plan.SBR.Spec.BindingReference != nil &&
		s.plan.SBR.Spec.BindingReference.ObjectType.Kind == ConfigMapKind {
		gvk = corev1.SchemeGroupVersion.WithKind(s.plan.SBR.Spec.BindingReference.ObjectType.Kind)
		resource = ConfigMapResource
	}

	objectMeta := metav1.ObjectMeta{
		Namespace: ns,
		Name:      name,
	}

	/*
		// panic: cannot deep copy map[string][]uint8 [recovered]
		// panic: cannot deep copy map[string][]uint8

		u := &unstructured.Unstructured{}

		u.Object = map[string]interface{}{
			"data": map[string][]byte{},
		}
		u.SetName(name)
		u.SetNamespace(ns)
		u.SetAPIVersion("v1")
		u.SetKind("Secret")

	*/

	bindingRef = &corev1.Secret{
		ObjectMeta: objectMeta,
		Data:       payload,
	}
	if resource == ConfigMapResource {
		payloadAsString := map[string]string{}
		for index, element := range payload {
			payloadAsString[index] = string(element)
		}
		bindingRef = &corev1.ConfigMap{
			ObjectMeta: objectMeta,
			Data:       payloadAsString,
		}
	}

	u, err := converter.ToUnstructuredAsGVK(bindingRef, gvk)
	if err != nil {
		return nil, err
	}

	resourceClient := s.buildResourceClient(resource)

	logger.Debug("Attempt to create secret/configmap...")
	_, err = resourceClient.Create(u, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}

	logger.Debug("Secret/configmap already exists, updating contents instead...")
	_, err = resourceClient.Update(u, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Commit will store informed data as a secret, commit it against the API server. It can forward
// errors from custom environment parser component, or from the API server itself.
func (s *Secret) Commit(payload map[string][]byte) (*unstructured.Unstructured, error) {
	return s.createOrUpdate(payload)
}

// Get an unstructured object from the secret handled by this component. It can return errors in case
// the API server does.
func (s *Secret) Get() (*unstructured.Unstructured, bool, error) {
	resource := SecretResource
	if s.plan.SBR.Spec.BindingReference != nil {
		if s.plan.SBR.Spec.BindingReference.ObjectType.Kind == ConfigMapKind {
			resource = ConfigMapResource
		}
	}

	resourceClient := s.buildResourceClient(resource)
	u, err := resourceClient.Get(s.plan.Name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, false, err
	}
	return u, u != nil, nil
}

// Delete the secret represented by this component. It can return error when the API server does.
func (s *Secret) Delete() error {
	resource := SecretResource
	if s.plan.SBR.Spec.BindingReference != nil {
		if s.plan.SBR.Spec.BindingReference.ObjectType.Kind == ConfigMapKind {
			resource = ConfigMapResource
		}
	}
	resourceClient := s.buildResourceClient(resource)
	err := resourceClient.Delete(s.plan.Name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// NewSecret instantiate a new Secret.
func NewSecret(client dynamic.Interface, plan *Plan) *Secret {
	return &Secret{
		logger: log.NewLog("secret"),
		client: client,
		plan:   plan,
	}
}
