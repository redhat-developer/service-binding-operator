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
	ctx           context.Context   // request context
	client        client.Client     // Kubernetes API client
	plan          *Plan             // plan instance
	logger        logr.Logger       // logger instance
	data          map[string][]byte // data retrieved
	volumeKeys    []string
	bindingPrefix string
}

const (
	basePrefix              = "binding:env:object"
	secretPrefix            = basePrefix + ":secret"
	configMapPrefix         = basePrefix + ":configmap"
	attributePrefix         = "binding:env:attribute"
	volumeMountSecretPrefix = "binding:volumemount:secret"
)

// getNestedValue retrieve value from dotted key path
func (r *Retriever) getNestedValue(key string, sectionMap interface{}) (string, error) {
	if !strings.Contains(key, ".") {
		value, exists := sectionMap.(map[string]interface{})[key]
		if !exists {
			return "", fmt.Errorf("Can't find key '%s'", key)
		}
		return fmt.Sprintf("%v", value), nil
	}
	attrs := strings.SplitN(key, ".", 2)
	newSectionMap, exists := sectionMap.(map[string]interface{})[attrs[0]]
	if !exists {
		return "", fmt.Errorf("Can't find '%v' section in CR", attrs)
	}
	return r.getNestedValue(attrs[1], newSectionMap.(map[string]interface{}))
}

// getCRKey retrieve key in section from CR object, part of the "plan" instance.
func (r *Retriever) getCRKey(section string, key string) (string, error) {
	obj := r.plan.CR.Object
	objName := r.plan.CR.GetName()
	logger := r.logger.WithValues("CR.Name", objName, "CR.section", section, "CR.key", key)
	logger.Info("Reading CR attributes...")

	sectionMap, exists := obj[section]
	if !exists {
		return "", fmt.Errorf("Can't find '%s' section in CR named '%s'", section, objName)
	}

	return r.getNestedValue(key, sectionMap)
}

// read attributes from CR, where place means which top level key name contains the "path" actual
// value, and parsing x-descriptors in order to either directly read CR data, or read items from
// a secret.
func (r *Retriever) read(place, path string, xDescriptors []string) error {
	logger := r.logger.WithValues(
		"CR.Section", place,
		"CRDDescription.Path", path,
		"CRDDescription.XDescriptors", xDescriptors,
	)
	logger.Info("Reading CRDDescription attributes...")

	// holds the secret name and items
	secrets := make(map[string][]string)

	// holds the configMap name and items
	configMaps := make(map[string][]string)
	for _, xDescriptor := range xDescriptors {
		logger = logger.WithValues("CRDDescription.xDescriptor", xDescriptor)
		logger.Info("Inspecting xDescriptor...")
		pathValue, err := r.getCRKey(place, path)
		if err != nil {
			return err
		}

		if strings.HasPrefix(xDescriptor, secretPrefix) {
			secrets[pathValue] = append(secrets[pathValue], r.extractSecretItemName(xDescriptor))
		} else if strings.HasPrefix(xDescriptor, configMapPrefix) {
			configMaps[pathValue] = append(configMaps[pathValue], r.extractConfigMapItemName(xDescriptor))
		} else if strings.HasPrefix(xDescriptor, volumeMountSecretPrefix) {
			secrets[pathValue] = append(secrets[pathValue], r.extractSecretItemName(xDescriptor))
			r.volumeKeys = append(r.volumeKeys, pathValue)
		} else if strings.HasPrefix(xDescriptor, attributePrefix) {
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
	for name, items := range configMaps {
		// add the function readConfigMap
		err := r.readConfigMap(name, items)
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

// extractConfigMapItemName based in x-descriptor entry, removing prefix in order to keep only the
// configMap item name.
func (r *Retriever) extractConfigMapItemName(xDescriptor string) string {
	return strings.ReplaceAll(xDescriptor, fmt.Sprintf("%s:", configMapPrefix), "")
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

// readConfigMap based in configMap name and list of items, read a configMap from the same namespace informed
// in plan instance.
func (r *Retriever) readConfigMap(name string, items []string) error {
	logger := r.logger.WithValues("ConfigMap.Name", name, "ConfigMap.Items", items)
	logger.Info("Reading ConfigMap items...")
	configMapObj := corev1.ConfigMap{}
	err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.plan.Ns, Name: name}, &configMapObj)
	if err != nil {
		return err
	}
	logger.Info("Inspecting configMap data...")
	for key, value := range configMapObj.Data {
		logger.WithValues("configMap.Key.Name", key, "configMap.Key.Length", len(value)).
			Info("Inspecting configMap key...")
		// making sure key name has a configMap reference
		// string to byte
		r.store(fmt.Sprintf("configMap_%s", key), []byte(value))
	}

	return nil
}

// store key and value, formatting key to look like an environment variable.
func (r *Retriever) store(key string, value []byte) {
	key = strings.ReplaceAll(key, ":", "_")
	key = strings.ReplaceAll(key, ".", "_")
	if r.bindingPrefix == "" {
		key = fmt.Sprintf("%s_%s", r.plan.CR.GetKind(), key)
	} else {
		key = fmt.Sprintf("%s_%s_%s", r.bindingPrefix, r.plan.CR.GetKind(), key)
	}
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

	r.logger.Info("Looking for status-descriptors in 'status'...")
	for _, statusDescriptor := range r.plan.CRDDescription.StatusDescriptors {
		if err = r.read("status", statusDescriptor.Path, statusDescriptor.XDescriptors); err != nil {
			return err
		}
	}

	return r.saveDataOnSecret()
}

// NewRetriever instantiate a new retriever instance.
func NewRetriever(ctx context.Context, client client.Client, plan *Plan, bindingPrefix string) *Retriever {
	return &Retriever{
		ctx:           ctx,
		client:        client,
		logger:        logf.Log.WithName("retriever"),
		plan:          plan,
		data:          make(map[string][]byte),
		volumeKeys:    []string{},
		bindingPrefix: bindingPrefix,
	}
}
