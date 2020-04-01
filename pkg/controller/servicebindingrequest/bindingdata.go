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

// BindingDataHandler represents the binding data accumulated from backing
// services, and later persisted as a secret or a configMap.
type BindingDataHandler struct {
	logger *log.Log          // logger instance
	client dynamic.Interface // Kubernetes API client
	plan   *Plan             // plan instance
}

// buildResourceClient creates a resource client to handle corev1/secret resource.
func (s *BindingDataHandler) buildResourceClient(resource string) dynamic.ResourceInterface {
	gvr := corev1.SchemeGroupVersion.WithResource(resource)
	return s.client.Resource(gvr).Namespace(s.plan.Ns)
}

// createOrUpdate will take informed payload and create/update the binding secret/configmap
// one.
// If the spec.BindingReference is nil,
// Returns an error when the Kubernetes client returns an error.
func (s *BindingDataHandler) createOrUpdate(payload map[string][]byte) (*unstructured.Unstructured, error) {
	ns := s.plan.Ns
	name := s.plan.Name
	logger := s.logger.WithValues("Namespace", ns, "Name", name)

	var bindingRef interface{}

	// Default binding resource type is Secret
	resource := SecretResource
	gvk := corev1.SchemeGroupVersion.WithKind(SecretKind)

	// Only if otherwise specified as configmap, we'll use a binding
	// configmaps
	if s.plan.SBR.Spec.Binding != nil &&
		s.plan.SBR.Spec.Binding.Kind == ConfigMapKind {
		gvk = corev1.SchemeGroupVersion.WithKind(s.plan.SBR.Spec.Binding.Kind)
		resource = ConfigMapResource
	}

	objectMeta := metav1.ObjectMeta{
		Namespace: ns,
		Name:      name,
	}

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
func (s *BindingDataHandler) Commit(payload map[string][]byte) (*unstructured.Unstructured, error) {
	return s.createOrUpdate(payload)
}

// Get an unstructured object from the secret handled by this component. It can return errors in case
// the API server does.
func (s *BindingDataHandler) Get() (*unstructured.Unstructured, bool, error) {
	resource := SecretResource
	if s.plan.SBR.Spec.Binding != nil {
		if s.plan.SBR.Spec.Binding.Kind == ConfigMapKind {
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
func (s *BindingDataHandler) Delete() error {
	resource := SecretResource
	if s.plan.SBR.Spec.Binding != nil {
		if s.plan.SBR.Spec.Binding.Kind == ConfigMapKind {
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

// NewBindingDataHandler instantiates a new referrence to a BindingDataHandler object
// that would be used for operations on the binding configmap/secret.
func NewBindingDataHandler(client dynamic.Interface, plan *Plan) *BindingDataHandler {
	return &BindingDataHandler{
		logger: log.NewLog("secret"),
		client: client,
		plan:   plan,
	}
}
