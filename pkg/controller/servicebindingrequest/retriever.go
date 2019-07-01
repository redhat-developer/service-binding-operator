package servicebindingrequest

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// Retriever reads all data referred in plan instance, and store in a secret.
type Retriever struct {
	ctx    context.Context   // request context
	client client.Client     // Kubernetes API client
	plan   *Plan             // plan instance
	logger logr.Logger       // logger instance
	data   map[string][]byte // data retrieved
}

const (
	bindingPrefix = "SERVICE_BINDING"
	secretPrefix  = "urn:alm:descriptor:servicebindingrequest:env:object:secret"
)

// getCRDKey retrieve key in section from CRD object, part of the "plan" instance.
func (r *Retriever) getCRDKey(section string, key string) (string, error) {
	obj := r.plan.CRD.Object
	objName := r.plan.CRD.GetName()
	logger := r.logger.WithValues(
		"read.CRD.Name", objName,
		"read.CRD.section", section,
		"read.CRD.key", key,
	)
	logger.Info("Reading CRD attributes...")

	sectionMap, exists := obj[section]
	if !exists {
		return "", fmt.Errorf("Can't find '%s' section in CRD named '%s'", section, objName)
	}

	value, exists := sectionMap.(map[string]interface{})[key]
	if !exists {
		return "", fmt.Errorf("Can't find key '%s' in section '%s', on object named '%s'",
			key, section, objName)
	}

	logger.Info("CRD attribute is found!")
	// making sure we always return a string representation
	return fmt.Sprintf("%v", value), nil
}

// read attributes from CRD, where place means which top level key name contains the "path" actual
// value, and parsing x-descriptors in order to either directly read CRD data, or read items from
// a secret.
func (r *Retriever) read(place, path string, xDescriptors []string) error {
	logger := r.logger.WithValues(
		"CRD.Section", place,
		"CRDDescription.Path", path,
		"CRDDescription.XDescriptors", xDescriptors,
	)
	logger.Info("Reading CRDDescription attributes...")

	// holds the secret name and items
	secrets := make(map[string][]string)

	for _, xDescriptor := range xDescriptors {
		logger = logger.WithValues("CRDDescription.xDescriptor", xDescriptor)
		logger.Info("Inspecting xDescriptor...")
		pathValue, err := r.getCRDKey(place, path)
		if err != nil {
			return err
		}

		if strings.HasPrefix(xDescriptor, secretPrefix) {
			secrets[pathValue] = append(secrets[pathValue], r.extractSecretItemName(xDescriptor))
		} else {
			r.store(path, []byte(pathValue))
		}
	}

	for name, items := range secrets {
		// loading secret items all-at-once
		err := r.readSecret(name, items)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractSecretItemName based in x-descriptor entry, removing prefix in order to keep only the
// secret item name.
func (r *Retriever) extractSecretItemName(xDescriptor string) string {
	return strings.ReplaceAll(xDescriptor, fmt.Sprintf("%s:", secretPrefix), "")
}

// readSecret based in secret name and list of items, read a secret from the same namespace informed
// in plan instance.
func (r *Retriever) readSecret(name string, items []string) error {
	logger := r.logger.WithValues("Secret.Name", name, "Secret.Items", items)
	logger.Info("Reading secret items...")

	secretObj := corev1.Secret{}
	err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.plan.Ns, Name: name}, &secretObj)
	if err != nil {
		return err
	}

	logger.Info("Inspecting secret data...")
	for key, value := range secretObj.Data {
		logger.WithValues("Secret.Key.Name", key, "Secret.Key.Length", len(value)).
			Info("Inspecting secret key...")
		// making sure key name has a secret reference
		r.store(fmt.Sprintf("secret_%s", key), value)
	}

	return nil
}

// store key and value, formatting key to look like an environment variable.
func (r *Retriever) store(key string, value []byte) {
	key = strings.ReplaceAll(key, ":", "_")
	key = strings.ReplaceAll(key, ".", "_")
	key = fmt.Sprintf("%s_%s_%s", bindingPrefix, r.plan.CRD.GetKind(), key)
	key = strings.ToUpper(key)
	r.data[key] = value
}

// saveDataOnSecret create or update secret that will store the data collected.
func (r *Retriever) saveDataOnSecret() error {
	secretObj := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.plan.Name,
			Namespace: r.plan.Ns,
		},
		Data: r.data,
	}

	err := r.client.Create(r.ctx, secretObj)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return r.client.Update(r.ctx, secretObj)
}

// Retrieve loop and read data pointed by the references in plan instance.
func (r *Retriever) Retrieve() error {
	var err error

	r.logger.Info("Looking for spec-descriptors in 'spec'...")
	for _, specDescriptor := range r.plan.CRDDescription.SpecDescriptors {
		if err = r.read("spec", specDescriptor.Path, specDescriptor.XDescriptors); err != nil {
			return err
		}
	}

	r.logger.Info("Looking for spec-descriptors in 'status'...")
	for _, statusDescriptor := range r.plan.CRDDescription.StatusDescriptors {
		if err = r.read("status", statusDescriptor.Path, statusDescriptor.XDescriptors); err != nil {
			return err
		}
	}

	return r.saveDataOnSecret()
}

// NewRetriever instantiate a new retriever instance.
func NewRetriever(ctx context.Context, client client.Client, plan *Plan) *Retriever {
	return &Retriever{
		ctx:    ctx,
		client: client,
		logger: logf.Log.WithName("retriever"),
		plan:   plan,
		data:   make(map[string][]byte),
	}
}
