package collectors

import (
	"context"
	"fmt"
	"strings"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/plan"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	basePrefix              = "binding:env:object"
	secretPrefix            = basePrefix + ":secret"
	configMapPrefix         = basePrefix + ":configmap"
	attributePrefix         = "binding:env:attribute"
	volumeMountSecretPrefix = "binding:volumemount:secret"
)

// CSVCollector reads all data referred in plan instance
type CSVCollector struct {
	ctx           context.Context // request context
	client        client.Client   // Kubernetes API client
	plan          plan.Plan       // plan instance
	secretData    map[string][]byte
	volumeData    []string
	bindingPrefix string
}

// NewCSVCollector creates a new CSVCollector
func NewCSVCollector(ctx context.Context, client client.Client, plan plan.Plan, bindingPrefix string) CSVCollector {
	return CSVCollector{
		ctx:           ctx,
		client:        client,
		plan:          plan.Plan,
		bindingPrefix: bindingPrefix,
	}
}

// Collect returns all bindable metadata
func (c *CSVCollector) Collect() (*BindableMetadata, error) {

	for _, specDescriptor := range c.plan.CRDDescription.SpecDescriptors {
		if err := c.read("spec", specDescriptor.Path, specDescriptor.XDescriptors); err != nil {
			return nil, nil
		}
	}

	return nil, nil
}

func (c *CSVCollector) read(place, path string, xDescriptors []string) error {

	// holds the secret name and items
	secrets := make(map[string][]string)

	// holds the configMap name and items
	configMaps := make(map[string][]string)
	for _, xDescriptor := range xDescriptors {
		pathValue, err := c.getCRKey(place, path)
		if err != nil {
			return err
		}

		if strings.HasPrefix(xDescriptor, secretPrefix) {
			secrets[pathValue] = append(secrets[pathValue], c.extractSecretItemName(xDescriptor))
		} else if strings.HasPrefix(xDescriptor, configMapPrefix) {
			configMaps[pathValue] = append(configMaps[pathValue], c.extractConfigMapItemName(xDescriptor))
		} else if strings.HasPrefix(xDescriptor, volumeMountSecretPrefix) {
			secrets[pathValue] = append(secrets[pathValue], c.extractSecretItemName(xDescriptor))
			c.volumeData = append(c.volumeData, pathValue)
		} else if strings.HasPrefix(xDescriptor, attributePrefix) {
			c.store(path, []byte(pathValue))
		}
	}

	for name, items := range secrets {
		// loading secret items all-at-once
		err := c.readSecret(name, items)
		if err != nil {
			return err
		}
	}
	/* FIXIT

	for name, items := range configMaps {
		// add the function readConfigMap
		err := c.readConfigMap(name, items)
		if err != nil {
			return err
		}
	}
	*/
	return nil
}

func (c *CSVCollector) readSecret(name string, items []string) error {
	secretObj := corev1.Secret{}
	err := c.client.Get(c.ctx, types.NamespacedName{Namespace: c.plan.Ns, Name: name}, &secretObj)
	if err != nil {
		return err
	}
	for key, value := range secretObj.Data {
		c.store(fmt.Sprintf("secret_%s", key), value)
	}
	return nil
}

// getCRKey retrieve key in section from CR object, part of the "plan" instance.
func (c *CSVCollector) getCRKey(section string, key string) (string, error) {
	obj := c.plan.CR.Object
	objName := c.plan.CR.GetName()

	sectionMap, exists := obj[section]
	if !exists {
		return "", fmt.Errorf("Can't find '%s' section in CR named '%s'", section, objName)
	}

	return c.getNestedValue(key, sectionMap)
}

// getNestedValue retrieve value from dotted key path
func (c *CSVCollector) getNestedValue(key string, sectionMap interface{}) (string, error) {
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
	return c.getNestedValue(attrs[1], newSectionMap.(map[string]interface{}))
}

// store stores key and value, formatting key to look like an environment variable.
func (c *CSVCollector) store(key string, value []byte) {
	key = strings.ReplaceAll(key, ":", "_")
	key = strings.ReplaceAll(key, ".", "_")
	if c.bindingPrefix == "" {
		key = fmt.Sprintf("%s_%s", c.plan.CR.GetKind(), key)
	} else {
		key = fmt.Sprintf("%s_%s_%s", c.bindingPrefix, c.plan.CR.GetKind(), key)
	}
	key = strings.ToUpper(key)
	c.secretData[key] = value
}

// extractSecretItemName based in x-descriptor entry, removing prefix in order to keep only the
// secret item name.
func (c *CSVCollector) extractSecretItemName(xDescriptor string) string {
	return strings.ReplaceAll(xDescriptor, fmt.Sprintf("%s:", secretPrefix), "")
}

// extractConfigMapItemName based in x-descriptor entry, removing prefix in order to keep only the
// configMap item name.
func (c *CSVCollector) extractConfigMapItemName(xDescriptor string) string {
	return strings.ReplaceAll(xDescriptor, fmt.Sprintf("%s:", configMapPrefix), "")
}
