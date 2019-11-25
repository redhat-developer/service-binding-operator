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

type Secret struct {
	logger *log.Log          // logger instance
	client dynamic.Interface // Kubernetes API client
	plan   *Plan             // plan instance
}

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

func (s *Secret) buildResourceClient() dynamic.ResourceInterface {
	gvr := corev1.SchemeGroupVersion.WithResource(SecretResource)
	return s.client.Resource(gvr).Namespace(s.plan.Ns)
}

func (s *Secret) createOrUpdate(payload map[string][]byte) (*unstructured.Unstructured, error) {
	ns := s.plan.Ns
	name := s.plan.Name
	logger := s.logger.WithValues("Namespace", ns, "Name", name)
	secretObj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: payload,
	}

	gvk := corev1.SchemeGroupVersion.WithKind(SecretKind)
	u, err := converter.ToUnstructuredAsGVK(secretObj, gvk)
	if err != nil {
		return nil, err
	}

	resourceClient := s.buildResourceClient()

	logger.Debug("Attempt to create secret...")
	_, err = resourceClient.Create(u, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}

	logger.Debug("Secret already exists, updating contents instead...")
	_, err = resourceClient.Update(u, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Secret) Commit(data map[string][]byte) (*unstructured.Unstructured, error) {
	payload, err := s.customEnvParser(data)
	if err != nil {
		return nil, err
	}
	return s.createOrUpdate(payload)
}

func (s *Secret) Get() (*unstructured.Unstructured, error) {
	resourceClient := s.buildResourceClient()
	return resourceClient.Get(s.plan.Name, metav1.GetOptions{})
}

func (s *Secret) Delete() error {
	resourceClient := s.buildResourceClient()
	err := resourceClient.Delete(s.plan.Name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func NewSecret(client dynamic.Interface, plan *Plan) *Secret {
	return &Secret{
		logger: log.NewLog("secret"),
		client: client,
		plan:   plan,
	}
}
