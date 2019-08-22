package servicebindingrequest

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	planner "github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/planner"
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
	boolEnvVar 				= false
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
			r.store(path, []byte(pathValue), boolEnvVar)
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
		r.store(fmt.Sprintf("secret_%s", key), value, boolEnvVar)
	}
	return nil
}

// readConfigMap based in configMap name and list of items, read a configMap from the same namespace informed
// in plan instance.
func (r *Retriever) readConfigMap(name string, items []string, boolEnvVar bool) error {
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
		r.store(fmt.Sprintf("configMap_%s", key), []byte(value), boolEnvVar)
	}

	return nil
}

type envVariables map[string]string
var fetchedEnvVars []envVariables

func(r *Retriever) storeEnvVar(){
	for i, val := range fetchedEnvVars{
		for k, v := range val[i]{
			key := k
			value := v
		}
	key = strings.ReplaceAll(key, ":", "_")
	key = strings.ReplaceAll(key, ".", "_")
	key = strings.ToUpper(key)
	r.data[key] = value
	}
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
	if boolEnvVar == true{
			r.storeEnvVar()
		}
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

	r.logger.Info("Looking for environment variables in 'CR'...")
	if err = r.readEnvVar(); err!=nil{
		return err
	}

	return r.saveDataOnSecret()
}

func (r *Retriever) readEnvVar() error {

	//retrieving each r.plan.Sbr.envVar.Name
	for i, envVars := range planner.EnvVars {
		envVarName := envVars.Name
		envVarValue := envVars.Value
		if err = r.parse(envVarValue); err!=nil{
			return err
		}
		}
	}

func (r *Retriever) parse(value string) {
	re := regexp.MustCompile(`\$\{.*?\}`)
	tempStr := re.FindAllString(value, -1)
	for i, element := range tempStr {
		element = strings.Trim(element, "${")
		element = strings.Trim(element, "}")
		value = strings.Replace(value, element, r.fetchEnvVarValue(element), i+1)
	}
}
func (r *Retriever) fetchEnvVarValue(parsedValue string) string{
	var err error
	boolEnvVar = true
	if strings.Count(parsedValue, ".") == 1{
		r.store(path, []byte(pathValue), boolEnvVar) //attribute
	}else{
	//status.dbCredentials.user
	//status.dbCredentials.password
	//status.dbConnectionIP
	elements := strings.Split(parsedValue,".")
	ele1 := elements[0] // either a status or a spec
	ele2 := elements[1] // path value
	ele3 := elements[2] // if a configMap/Secret then the value for that object
	if strings.HasPrefix(parsedValue, "status"){
		r.logger.Info("Looking for status-descriptors in 'status'")
		statusDescriptor := r.plan.CRDDescription.StatusDescriptors.Path.xDescriptors {
			secrets := make(map[string][]string)
			// holds the configMap name and items
			configMaps := make(map[string][]string)
			for _, xDescriptor := range statusDescriptors {
				logger = logger.WithValues("CRDDescription.xDescriptor", xDescriptor)
				logger.Info("Inspecting xDescriptor...")
				pathValue, err := r.getCRKey(place, path)
				if err != nil {
					return err
				}	
				if strings.HasPrefix(xDescriptor, secretPrefix+ele3 {
					logger := r.logger.WithValues("Secret.Name", name, "Secret.Items", items)
					logger.Info("Reading secret items...")
					secretFound := &corev1.Secret{}
					err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.plan.Ns, Name: name}, &secretFound)
					if err != nil {
						return err
					}
					logger.Info("Inspecting secret data...")
					for key, value := range secretFound.Data {
						logger.WithValues("Secret.Key.Name", key, "Secret.Key.Length", len(value)).
							Info("Inspecting secret key...")
					crEnvVar := p.Sbr.Spec.EnvVar
					for i, envMap := range crEnvVar {
						EnvVars[i].Name = envMap.Name
						key := EnvVars[i].Name
						fetchedValue := secretFound.Data[key] 
						fetchedEnvVars[i] = append(fetchedEnvVars, map[key]fetchedValue)
					}
					secrets[pathValue] = append(secrets[pathValue], r.extractSecretItemName(xDescriptor))
				} 
			}else if strings.HasPrefix(xDescriptor, configMapPrefix+ele3{
				logger := r.logger.WithValues("ConfigMap.Name", name, "ConfigMap.Items", items)
					logger.Info("Reading configmap items...")
					configMapFound := &corev1.ConfigMap{}
					err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.plan.Ns, Name: name}, &configMapFound)
					if err != nil {
						return err
					}
					logger.Info("Inspecting Config Map data...")
					for key, value := range configMapFound.Data {
						logger.WithValues("ConfigMap.Key.Name", key, "ConfigMap.Key.Length", len(value)).
							Info("Inspecting configmap ...")
					crEnvVar := p.Sbr.Spec.EnvVar
					for i, envMap := range crEnvVar {
						EnvVars[i].Name = envMap.Name
						key := EnvVars[i].Name
						fetchedValue := configMapFound.Data[key] 
						fetchedEnvVars[i] = append(fetchedEnvVars, map[key]fetchedValue)
					}
					configMaps[pathValue] = append(configMaps[pathValue], r.extractConfigMapItemName(xDescriptor))
				} 
			}
		}
	}

			for name, items := range secrets {
				// loading secret items all-at-once
				err := r.readSecret(name, items, boolEnvVar)
				if err != nil {
					return err
				}
			}
			for name, items := range configMaps {
				// add the function readConfigMap
				err := r.readConfigMap(name, items, boolEnvVar)
				if err != nil {
					return err
				}
			}	
		}else if strings.HasPrefix(parsedValue, "spec"){
			r.logger.Info("Looking for spec-descriptors in 'spec'")
			specDescriptor := r.plan.CRDDescription.SpecDescriptors.Path.xDescriptors {
			//user and password
			//one is secretPrefix
			//other one is configMapPrefix
			// holds the secret name and items
			secrets := make(map[string][]string)
			// holds the configMap name and items
			configMaps := make(map[string][]string)
			for _, xDescriptor := range specDescriptors {
				logger = logger.WithValues("CRDDescription.xDescriptor", xDescriptor)
				logger.Info("Inspecting xDescriptor...")
				pathValue, err := r.getCRKey(place, path)
				if err != nil {
					return err
				}	
				if strings.HasPrefix(xDescriptor, secretPrefix+ele3 {
					// how to get to user??
					// get the exact secretPrefix:user i.e ele[2] value
					// Check if this Secret already exists
					logger := r.logger.WithValues("Secret.Name", name, "Secret.Items", items)
					logger.Info("Reading secret items...")
					secretFound := &corev1.Secret{}
					err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.plan.Ns, Name: name}, &secretFound)
					if err != nil {
						return err
					}
					logger.Info("Inspecting secret data...")
					for key, value := range secretFound.Data {
						logger.WithValues("Secret.Key.Name", key, "Secret.Key.Length", len(value)).
							Info("Inspecting secret key...")
						// making sure key name has a secret reference
					crEnvVar := p.Sbr.Spec.EnvVar
					for i, envMap := range crEnvVar {
						EnvVars[i].Name = envMap.Name
						key := EnvVars[i].Name
						fetchedValue := secretFound.Data[key] 
						fetchedEnvVars[i] = append(fetchedEnvVars, map[key]fetchedValue)
					}
					// path value??
					secrets[pathValue] = append(secrets[pathValue], r.extractSecretItemName(xDescriptor))
				} 
			}else if strings.HasPrefix(xDescriptor, configMapPrefix+ele3{
				logger := r.logger.WithValues("ConfigMap.Name", name, "ConfigMap.Items", items)
					logger.Info("Reading configmap items...")
					configMapFound := &corev1.ConfigMap{}
					err := r.client.Get(r.ctx, types.NamespacedName{Namespace: r.plan.Ns, Name: name}, &configMapFound)
					if err != nil {
						return err
					}
					logger.Info("Inspecting Config Map data...")
					for key, value := range configMapFound.Data {
						logger.WithValues("ConfigMap.Key.Name", key, "ConfigMap.Key.Length", len(value)).
							Info("Inspecting configmap ...")
						// making sure key name has a secret reference
					crEnvVar := p.Sbr.Spec.EnvVar
					for i, envMap := range crEnvVar {
						EnvVars[i].Name = envMap.Name
						key := EnvVars[i].Name
						fetchedValue := configMapFound.Data[key] 
						fetchedEnvVars[i] = append(fetchedEnvVars, map[key]fetchedValue)
					}
					configMaps[pathValue] = append(configMaps[pathValue], r.extractConfigMapItemName(xDescriptor))
				} 
			}
		}
	}

			for name, items := range secrets {
				// loading secret items all-at-once
				err := r.readSecret(name, items, boolEnvVar)
				if err != nil {
					return err
				}
			}
			for name, items := range configMaps {
				// add the function readConfigMap
				err := r.readConfigMap(name, items, boolEnvVar)
				if err != nil {
					return err
				}
			}	
		
		}
	}
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
