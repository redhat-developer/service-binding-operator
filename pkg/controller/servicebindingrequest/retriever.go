package servicebindingrequest

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// Retriever reads all data referred in plan instance, and store in a secret.
type Retriever struct {
	logger        logr.Logger                  // logger instance
	data          map[string][]byte            // data retrieved
	objects       []*unstructured.Unstructured // list of objects employed
	client        dynamic.Interface            // Kubernetes API client
	plan          *Plan                        // plan instance
	volumeKeys    []string                     // list of keys found
	bindingPrefix string                       // prefix for variable names
	cache         map[string]interface{}       // store visited paths
}

const (
	basePrefix              = "binding:env:object"
	secretPrefix            = basePrefix + ":secret"
	configMapPrefix         = basePrefix + ":configmap"
	attributePrefix         = "binding:env:attribute"
	volumeMountSecretPrefix = "binding:volumemount:secret"
)

// getNestedValue retrieve value from dotted key path
func (r *Retriever) getNestedValue(key string, sectionMap interface{}) (string, error, interface{}) {
	logger := r.logger.WithValues("Key", key, "SectionMap", sectionMap)
	if !strings.Contains(key, ".") {
		value, exists := sectionMap.(map[string]interface{})[key]
		if !exists {
			return "", fmt.Errorf("Can't find key '%s'", key), sectionMap
		}
		return fmt.Sprintf("%v", value), nil, sectionMap
	}
	attrs := strings.SplitN(key, ".", 2)
	newSectionMap, exists := sectionMap.(map[string]interface{})[attrs[0]]
	logger.Info("Section maps :  ")
	if !exists {
		return "", fmt.Errorf("Can't find '%v' section in CR", attrs), newSectionMap
	}
	return r.getNestedValue(attrs[1], newSectionMap.(map[string]interface{}))
}

// getCRKey retrieve key in section from CR object, part of the "plan" instance.
func (r *Retriever) getCRKey(section string, key string) (string, interface{}, error) {
	obj := r.plan.CR.Object
	objName := r.plan.CR.GetName()
	logger := r.logger.WithValues("CR.Name", objName, "CR.section", section, "CR.key", key)
	logger.Info("Reading CR attributes...")

	sectionMap, exists := obj[section]
	if !exists {
		return "", sectionMap, fmt.Errorf("Can't find '%s' section in CR named '%s'", section, objName)
	}

	v, err, _ := r.getNestedValue(key, sectionMap)
	for k, v := range sectionMap.(map[string]interface{}) {
		if _, ok := r.cache[section]; !ok {
			r.cache[section] = make(map[string]interface{})
		}
		r.cache[section].(map[string]interface{})[k] = v
	}
	return v, sectionMap, err
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
	pathValue, _, err := r.getCRKey(place, path)
	for _, xDescriptor := range xDescriptors {
		logger = logger.WithValues("CRDDescription.xDescriptor", xDescriptor, "cache", r.cache)
		logger.Info("Inspecting xDescriptor...")
		if err != nil {
			return err
		}

		if _, ok := r.cache[place].(map[string]interface{}); !ok {
			r.cache[place] = make(map[string]interface{})
		}
		if strings.HasPrefix(xDescriptor, secretPrefix) {
			secrets[pathValue] = append(secrets[pathValue], r.extractSecretItemName(xDescriptor))
			if _, ok := r.cache[place].(map[string]interface{})[r.extractSecretItemName(xDescriptor)]; !ok {
				r.markVisitedPaths(r.extractSecretItemName(xDescriptor), pathValue, place)
				r.cache[place].(map[string]interface{})[r.extractSecretItemName(xDescriptor)] = make(map[string]interface{})
			}
		} else if strings.HasPrefix(xDescriptor, configMapPrefix) {
			configMaps[pathValue] = append(configMaps[pathValue], r.extractConfigMapItemName(xDescriptor))
			r.markVisitedPaths(r.extractConfigMapItemName(xDescriptor), pathValue, place)
		} else if strings.HasPrefix(xDescriptor, volumeMountSecretPrefix) {
			secrets[pathValue] = append(secrets[pathValue], r.extractSecretItemName(xDescriptor))
			r.markVisitedPaths(r.extractSecretItemName(xDescriptor), pathValue, place)
			r.volumeKeys = append(r.volumeKeys, pathValue)
		} else if strings.HasPrefix(xDescriptor, attributePrefix) {
			r.store(path, []byte(pathValue))
		} else {
			logger.Info("Defaulting....")
		}
	}

	for name, items := range secrets {
		// loading secret items all-at-once
		err := r.readSecret(name, items, place, path)
		if err != nil {
			return err
		}
	}
	for name, items := range configMaps {
		// add the function readConfigMap
		err := r.readConfigMap(name, items, place, path)
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

// markVisitedPaths updates all visited paths in cache, This initializes the cache map
func (r *Retriever) markVisitedPaths(name, keyPath, fromPath string) {
	if _, ok := r.cache[fromPath]; !ok {
		r.cache[fromPath] = make(map[string]interface{})
	}
	if _, ok := r.cache[fromPath].(map[string]interface{})[name]; !ok {
		r.cache[fromPath].(map[string]interface{})[name] = make(map[string]interface{})
	}
	if _, ok := r.cache[fromPath].(map[string]interface{})[name].(map[string]interface{}); !ok {
		r.cache[fromPath].(map[string]interface{})[name] = make(map[string]interface{})
	}
	if _, ok := r.cache[fromPath].(map[string]interface{})[name].(map[string]interface{})[keyPath]; !ok {
		r.cache[fromPath].(map[string]interface{})[name].(map[string]interface{})[keyPath] = make(map[string]interface{})
	}
}

// readSecret based in secret name and list of items, read a secret from the same namespace informed
// in plan instance.
func (r *Retriever) readSecret(
	name string,
	items []string,
	fromPath string,
	path string) error {
	logger := r.logger.WithValues("Secret.Name", name, "Secret.Items", items)
	logger.Info("Reading secret items...")

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	u, err := r.client.Resource(gvr).Namespace(r.plan.Ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	data, exists, err := unstructured.NestedMap(u.Object, []string{"data"}...)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("could not find 'data' in secret")
	}

	for k, v := range data {
		value := v.(string)
		data, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return err
		}
		logger.WithValues("Secret.Key.Name", k, "Secret.Key.Length", len(data)).
			Info("Inspecting secret key...")
		r.markVisitedPaths(path, k, fromPath)
		// update cache after reading configmap/secret in cache
		r.cache[fromPath].(map[string]interface{})[path].(map[string]interface{})[k] = string(value)
		// making sure key name has a secret reference
		r.store(fmt.Sprintf("configMap_%s", k), data)
		r.store(fmt.Sprintf("secret_%s", k), data)
	}

	r.objects = append(r.objects, u)
	return nil
}

// readConfigMap based in configMap name and list of items, read a configMap from the same namespace informed
// in plan instance.
func (r *Retriever) readConfigMap(
	name string,
	items []string,
	fromPath string,
	path string) error {
	logger := r.logger.WithValues("ConfigMap.Name", name, "ConfigMap.Items", items)
	logger.Info("Reading ConfigMap items...")

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	u, err := r.client.Resource(gvr).Namespace(r.plan.Ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	data, exists, err := unstructured.NestedMap(u.Object, []string{"data"}...)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("could not find 'data' in secret")
	}

	logger.Info("Inspecting configMap data...")
	for k, v := range data {
		value := v.(string)
		logger.WithValues("configMap.Key.Name", k, "configMap.Key.Length", len(value)).
			Info("Inspecting configMap key...")
		r.markVisitedPaths(path, k, fromPath)
		// update cache after reading configmap/secret in cache
		r.cache[fromPath].(map[string]interface{})[path].(map[string]interface{})[k] = value
		// making sure key name has a configMap reference
		r.store(fmt.Sprintf("configMap_%s", k), []byte(value))
	}

	r.objects = append(r.objects, u)
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
	gvk := schema.GroupVersion{Group: "", Version: "v1"}.WithKind("Secret")
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	resourceClient := r.client.Resource(gvr).Namespace(r.plan.Ns)
	logger := r.logger.WithValues(
		"Secret.GVK", gvk.String(),
		"Secret.Namespace", r.plan.Ns,
		"Secret.Name", r.plan.Name,
	)
	logger.Info("Retrieving intermediary secret...")

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

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secretObj)
	if err != nil {
		r.logger.Error(err, "Converting secret to unstructured")
		return err
	}
	u := &unstructured.Unstructured{Object: data}
	u.SetGroupVersionKind(gvk)

	logger.Info("Creating intermediary secret...")
	_, err = resourceClient.Create(u, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		logger.Error(err, "on creating intermediary secret")
		return err
	}
	logger.Info("Secret is already found, updating...")
	_, err = resourceClient.Update(u, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "on updating intermediary secret")
		return err
	}

	logger.Info("Intermediary secret created/updated!")
	r.objects = append(r.objects, u)
	return nil
}

// Retrieve loop and read data pointed by the references in plan instance. It returns a slice of
// Unstructured refering the objects in use by the Retriever, and error when issues reading fields.
func (r *Retriever) Retrieve() ([]*unstructured.Unstructured, error) {
	var err error

	r.logger.Info("Looking for spec-descriptors in 'spec'...")
	for _, specDescriptor := range r.plan.CRDDescription.SpecDescriptors {
		if err = r.read("spec", specDescriptor.Path, specDescriptor.XDescriptors); err != nil {
			return nil, err
		}
	}

	r.logger.Info("Looking for status-descriptors in 'status'...")
	for _, statusDescriptor := range r.plan.CRDDescription.StatusDescriptors {
		if err = r.read("status", statusDescriptor.Path, statusDescriptor.XDescriptors); err != nil {
			return nil, err
		}
	}

	r.logger.WithValues("cache", r.cache).Info("Final cache values...")

	if r.plan.SBR.Spec.BindUnannotated {
		b := NewBindNonBindable(&r.plan.SBR, r.plan.CR, []schema.GroupVersionResource{
			// We can add extra gvrs here
			{Group: "", Version: "v1", Resource: "configmaps"},
			{Group: "", Version: "v1", Resource: "services"},
			{Group: "route.openshift.io", Version: "v1", Resource: "routes"},
		}, r.client)

		vals, err := b.GetBindableVariables()
		if err != nil {
			return nil, err
		}
		for k,v := range vals {
			r.store(k,[]byte(fmt.Sprintf("%v", v)))
		}
	}

	envParser := NewCustomEnvParser(r.plan.SBR.Spec.CustomEnvVar, r.cache)
	values, err := envParser.Parse()
	if err != nil {
		return nil, err
	}
	for k, v := range values {
		r.data[k] = []byte(v.(string))
	}

	r.logger.Info("Saving data on intermediary secret...")
	if err = r.saveDataOnSecret(); err != nil {
		return nil, err
	}
	return r.objects, nil
}

// NewRetriever instantiate a new retriever instance.
func NewRetriever(client dynamic.Interface, plan *Plan, bindingPrefix string) *Retriever {
	return &Retriever{
		logger:        logf.Log.WithName("retriever"),
		data:          make(map[string][]byte),
		objects:       []*unstructured.Unstructured{},
		client:        client,
		plan:          plan,
		volumeKeys:    []string{},
		bindingPrefix: bindingPrefix,
		cache:         make(map[string]interface{}),
	}
}
