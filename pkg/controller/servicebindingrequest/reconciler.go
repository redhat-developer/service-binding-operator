package servicebindingrequest

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	osappsv1 "github.com/openshift/api/apps/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

const (
	connectsToLabel                = "connects-to"
	almDescriptorPrefix            = "urn:alm:descriptor"
	almKubernetesPrefix            = "io.kubernetes"
	almServiceBindingRequestPrefix = "io.servicebindingrequest"
)

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

// decodeString encoded with base64.
func decodeString(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// pathValue read value of a given "path" inside of informed custom resource definition instance.
func (r *Reconciler) pathValue(crd *unstructured.Unstructured, path string) (string, error) {
	object := crd.Object
	status, exists := object["status"].(map[string]interface{})
	if !exists {
		return "", fmt.Errorf("unable to find 'status' field in object '%s'", crd.GetName())
	}

	value, exists := status[path]
	if !exists {
		return "", fmt.Errorf("unable to find attribute for path '%s'", path)
	}

	return fmt.Sprintf("%v", value), nil
}

// parseXDescriptor parse's OLM CustomResourceDefinition descripton in order to return it's kind
// and attribute.
func (r *Reconciler) parseXDescriptor(xDescriptor string) (string, string) {
	if !strings.HasPrefix(xDescriptor, almDescriptorPrefix) {
		return "", ""
	}

	prefix := fmt.Sprintf("%s:%s", almDescriptorPrefix, almServiceBindingRequestPrefix)
	if !strings.HasPrefix(xDescriptor, prefix) {
		return "", ""
	}

	sections := strings.Split(xDescriptor, ":")
	kind := sections[len(sections)-2]
	attribute := sections[len(sections)-1]

	return kind, attribute
}

// generateFieldName based on the x-descriptor value, generating a string with the same standards
// than environment variable naming.
func (r *Reconciler) generateFieldName(xDescriptor string) string {
	sections := strings.Split(xDescriptor, ":")
	name := strings.Join(sections, "_")
	name = strings.ReplaceAll(name, ".", "_")
	return strings.ToUpper(name)
}

// readObjectAttributes read a unstructured object in order to read the attributes informed.
func (r *Reconciler) readObjectAttributes(
	logger logr.Logger,
	ns string,
	kind string,
	name string,
	attributes []string,
) (map[string]string, error) {
	logger = logger.WithValues("binding.Kind", kind, "binding.Name", name)
	logger.Info("Searching for object to extract values...")

	// template to find a object of a given kind
	unstructuredTemplateObj := map[string]interface{}{"kind": kind}
	unstructuredObjList := &unstructured.UnstructuredList{Object: unstructuredTemplateObj}
	err := r.client.List(context.TODO(), &client.ListOptions{Namespace: ns}, unstructuredObjList)
	if err != nil {
		return nil, err
	}

	// selecting object based on name
	var unstructuredObj *unstructured.Unstructured
	for _, object := range unstructuredObjList.Items {
		if name == object.GetName() {
			unstructuredObj = &object
		}
	}

	if unstructuredObj == nil {
		logger.Info("Warning: object is not found!")
		// FIXME: won't work in anything that's not part of the core;
		return nil, errors.NewNotFound(corev1.Resource(kind), name)
	}

	logger.Info("Object found, extracting attributes...")
	rawData, exists := unstructuredObj.Object["data"].(map[string]interface{})
	if !exists {
		logger.Info("Warning: unable to find 'data' field in object!")
		return nil, errors.NewNotFound(corev1.Resource(kind), name)
	}

	data := make(map[string]string)
	for _, attribute := range attributes {
		logger = logger.WithValues("attribute", attribute)
		logger.Info("Inspecting attribute...")

		value, exists := rawData[attribute]
		if !exists {
			logger.Info("Warning: attribute could not be found!")
			continue
		}
		logger.Info("Reading attribute value")
		if kind == "secret" {
			data[attribute], _ = decodeString(value.(string))
		} else {
			data[attribute] = value.(string)
		}
	}

	return data, nil
}

// retrieveBindingData using CRD-Description and CRD instantiated object, using description to find
// attributes in different objects that belong to binding process and return those fields as a map.
func (r *Reconciler) retrieveBindingData(
	logger logr.Logger,
	ns string,
	crdDescription *olmv1alpha1.CRDDescription,
	crd *unstructured.Unstructured,
) (map[string][]byte, error) {
	data := make(map[string][]byte)

	logger.Info("Inspecting custom resource definition status descriptor object...")

	// TODO: way too long loop, should be extracted;
	for _, statusDescriptor := range crdDescription.StatusDescriptors {
		logger = logger.WithValues(
			"StatusDescriptor.DisplayName", statusDescriptor.DisplayName,
			"StatusDescriptor.Path", statusDescriptor.Path)
		logger.Info("Descriptor entry")

		// store object kind as key and list of attributes as value
		kindAttributes := make(map[string][]string)
		// store the relationship of attribute name and field name, used to return
		attributeFieldName := make(map[string]string)

		// retrieve the path value from the CRD
		objectName, err := r.pathValue(crd, statusDescriptor.Path)
		if err != nil {
			return nil, err
		}
		if objectName == "" {
			logger.Info("Warning: Unable to extract object-name from CRD, skipping!")
			continue
		}

		logger = logger.WithValues("binding.CRD.Name", objectName)
		logger.Info("Found object name to inspect...")

		for _, xDescriptor := range statusDescriptor.XDescriptors {
			kind, attribute := r.parseXDescriptor(xDescriptor)
			if kind == "" || attribute == "" {
				logger.Info("Unable to parse kind and attribute, skipping!")
				continue
			}

			// appending the attributes per kind of object
			kindAttributes[kind] = append(kindAttributes[kind], attribute)
			// applying convention to transform attribute names
			attributeFieldName[attribute] = r.generateFieldName(xDescriptor)

			logger.WithValues(
				"xDescriptor.Kind", kind,
				"xDescriptor.Attribute", attribute,
				"binding.Field.Name", attributeFieldName[attribute],
			).Info("Found kind in xDescriptor entry")
		}

		for kind, attributes := range kindAttributes {
			// read all keys from object that we intent to bind
			descriptorData, err := r.readObjectAttributes(logger, ns, kind, objectName, attributes)
			if err != nil {
				return nil, err
			}

			// storing object data into final data structure, used in return
			for k, v := range descriptorData {
				data[attributeFieldName[k]] = []byte(v)
			}
		}
	}

	return data, nil
}

// appendEnvFrom based on secret name and list of EnvFromSource instances, making sure secret is
// part of the list or appended.
func (r *Reconciler) appendEnvFrom(envList []corev1.EnvFromSource, secret string) []corev1.EnvFromSource {
	for _, env := range envList {
		if env.SecretRef.Name == secret {
			// secret name is already referenced
			return envList
		}
	}

	return append(envList, corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret,
			},
		},
	})
}

// Reconcile a ServiceBindingRequest by the following steps:
// 1. Inspecting SBR in order to identify backend service. The service is composed by a CRD name and
//    kind, and by inspecting "connects-to" label identify the name of service instance;
// 2. Using OperatorLifecycleManager standards, identifying which items are intersting for binding
//    by parsing CustomResourceDefinitionDescripton object;
// 3. Search and read contents identified in previous step, creating a intermediary secret to hold
//    data formatted as environment variables key/value.
// 4. Search applications that are interested to bind with given service, by inspecting labels. The
//    Deployment (and other kinds) will be updated in PodTeamplate level updating `envFrom` entry
// 	  to load interdiary secret;
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// TODO: very long method that needs to be extracted;
	ctx := context.TODO()
	logger := logf.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling ServiceBindingRequest")

	// Fetch the ServiceBindingRequest instance
	instance := &v1alpha1.ServiceBindingRequest{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// binding-request is not found, empty result means requeue
			return reconcile.Result{}, nil
		}
		// error on executing the request, requeue informing the error
		return reconcile.Result{}, err
	}

	logger.WithValues("ServiceBindingRequest.Name", instance.Name).
		Info("Found service binding request to inspect")

	planner := NewPlanner(ctx, r.client, request.Namespace, instance)
	plan, err := planner.Plan()
	if err != nil {
		return RequeueOnNotFound(err)
	}

	bindingData, err := r.retrieveBindingData(logger, plan.ns, plan.crdDescription, plan.crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// when underlying objects are not found, simple requeue without error
			return reconcile.Result{}, nil
		}
		return reconcile.Result{Requeue: true}, err
	}

	bindingSecretObj := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName(),
			Namespace: request.Namespace,
		},
		Data: bindingData,
	}

	err = r.client.Create(ctx, bindingSecretObj)
	if err != nil && !errors.IsAlreadyExists(err) {
		logger.Error(err, "Error on creating secret")
		return reconcile.Result{Requeue: true}, err
	}

	//
	// Updating applications to use intermediary secret
	//

	logger = logger.WithValues("MatchLabels", instance.Spec.ApplicationSelector.MatchLabels)
	logger.Info("Searching applications to receive intermediary secret bind...")

	resourceKind := strings.ToLower(instance.Spec.ApplicationSelector.ResourceKind)
	searchByLabelsOpts := client.ListOptions{
		Namespace:     request.Namespace,
		LabelSelector: labels.SelectorFromSet(instance.Spec.ApplicationSelector.MatchLabels),
	}

	switch resourceKind {
	case "deploymentconfig":
		logger.Info("Inspecting DeploymentConfig objects matching labels")

		deploymentConfigListObj := &osappsv1.DeploymentConfigList{}
		err = r.client.List(ctx, &searchByLabelsOpts, deploymentConfigListObj)
		if err != nil {
			if errors.IsNotFound(err) {
				return reconcile.Result{}, nil
			}
			return reconcile.Result{Requeue: true}, err
		}

		for _, deploymentConfigObj := range deploymentConfigListObj.Items {
			logger.WithValues("DeploymentConfig.Name", deploymentConfigObj.GetName()).
				Info("Inspecting DeploymentConfig object...")

			for i, c := range deploymentConfigObj.Spec.Template.Spec.Containers {
				logger.Info("Adding EnvFrom to container")
				deploymentConfigObj.Spec.Template.Spec.Containers[i].EnvFrom = r.appendEnvFrom(
					c.EnvFrom, instance.GetName())
			}

			logger.Info("Updating DeploymentConfig object")
			err = r.client.Update(ctx, &deploymentConfigObj)
			if err != nil {
				logger.Error(err, "Error on updating object!")
				return reconcile.Result{}, err
			}
		}
	default:
		logger.Info("Inspecting Deployment objects matching labels")

		deploymentListObj := &extv1beta1.DeploymentList{}
		err = r.client.List(ctx, &searchByLabelsOpts, deploymentListObj)
		if err != nil {
			if errors.IsNotFound(err) {
				return reconcile.Result{}, nil
			}
			return reconcile.Result{Requeue: true}, err
		}

		for _, deploymentObj := range deploymentListObj.Items {
			logger = logger.WithValues("Deployment.Name", deploymentObj.GetName())
			logger.Info("Inspecting Deploymen object...")

			for i, c := range deploymentObj.Spec.Template.Spec.Containers {
				logger.Info("Adding EnvFrom to container")
				deploymentObj.Spec.Template.Spec.Containers[i].EnvFrom = r.appendEnvFrom(
					c.EnvFrom, instance.GetName())
			}

			logger.Info("Updating Deployment object")
			err = r.client.Update(ctx, &deploymentObj)
			if err != nil {
				logger.Error(err, "Error on updating object!")
				return reconcile.Result{}, err
			}
		}
	}

	logger.Info("All done!")
	return reconcile.Result{}, nil
}
