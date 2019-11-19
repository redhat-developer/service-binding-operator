package servicebindingrequest

import (
	"context"
	"fmt"

	"gotest.tools/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

var (
	binderLog = log.NewLog("binder")
)

// Binder executes the "binding" act of updating different application kinds to use intermediary
// secret. Those secrets should be offered as environment variables.
type Binder struct {
	ctx        context.Context                 // request context
	client     client.Client                   // kubernetes API client
	dynClient  dynamic.Interface               // kubernetes dynamic api client
	sbr        *v1alpha1.ServiceBindingRequest // instantiated service binding request
	volumeKeys []string                        // list of key names used in volume mounts
	logger     *log.Log                        // logger instance
}

// search objects based in Kind/APIVersion, which contain the labels defined in ApplicationSelector.
func (b *Binder) search() (*unstructured.UnstructuredList, error) {
	ns := b.sbr.GetNamespace()
	gvr := schema.GroupVersionResource{
		Group:    b.sbr.Spec.ApplicationSelector.Group,
		Version:  b.sbr.Spec.ApplicationSelector.Version,
		Resource: b.sbr.Spec.ApplicationSelector.Resource,
	}

	var opts metav1.ListOptions

	// If Application name is present
	if b.sbr.Spec.ApplicationSelector.ResourceRef != "" {
		fieldName := make(map[string]string)
		fieldName["metadata.name"] = b.sbr.Spec.ApplicationSelector.ResourceRef
		opts = metav1.ListOptions{
			FieldSelector: fields.Set(fieldName).String(),
		}
	} else if b.sbr.Spec.ApplicationSelector.MatchLabels != nil {
		matchLabels := b.sbr.Spec.ApplicationSelector.MatchLabels
		opts = metav1.ListOptions{
			LabelSelector: labels.Set(matchLabels).String(),
		}
	} else {
		return nil, fmt.Errorf("application ResourceRef or MatchLabel not found")
	}

	objList, err := b.dynClient.Resource(gvr).Namespace(ns).List(opts)
	if err != nil {
		return nil, err
	}

	// Return fake NotFound error explicitly to ensure requeue when objList(^) is empty.
	if len(objList.Items) == 0 {
		return nil, errors.NewNotFound(
			gvr.GroupResource(),
			b.sbr.Spec.ApplicationSelector.Resource,
		)
	}
	return objList, err
}

// updateSpecVolumes execute the inspection and update "volumes" entries in informed spec.
func (b *Binder) updateSpecVolumes(
	obj *unstructured.Unstructured,
) (*unstructured.Unstructured, error) {
	volumesPath := []string{"spec", "template", "spec", "volumes"}
	log := b.logger.WithValues("Volumes.NestedPath", volumesPath)

	log.Debug("Reading volumes definitions...")
	volumes, _, err := unstructured.NestedSlice(obj.Object, volumesPath...)
	if err != nil {
		return nil, err
	}
	log.Debug("Amount of volumes in spec.", "Volumes", len(volumes))

	volumes, err = b.updateVolumes(volumes)
	if err != nil {
		return nil, err
	}
	if err = unstructured.SetNestedSlice(obj.Object, volumes, volumesPath...); err != nil {
		return nil, err
	}

	return obj, nil
}

// updateVolumes inspect informed list assuming as []corev1.Volume, and if binding volume is already
// defined just return the same list, otherwise, appending the binding volume.
func (b *Binder) updateVolumes(volumes []interface{}) ([]interface{}, error) {
	name := b.sbr.GetName()
	log := b.logger
	log.Debug("Checking if binding volume is already defined...")
	for _, v := range volumes {
		volume := v.(corev1.Volume)
		if name == volume.Name {
			log.Debug("Volume is already defined!")
			return volumes, nil
		}
	}

	items := []corev1.KeyToPath{}
	for _, k := range b.volumeKeys {
		items = append(items, corev1.KeyToPath{Key: k, Path: k})
	}

	log.Debug("Appending new volume with items.", "Items", items)
	bindVolume := corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: name,
				Items:      items,
			},
		},
	}

	// making sure tranforming it back to unstructured before returning
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&bindVolume)
	if err != nil {
		return nil, err
	}
	return append(volumes, u), nil
}

// updateSpecContainers extract containers from object, and trigger update.
func (b *Binder) updateSpecContainers(
	obj *unstructured.Unstructured,
) (*unstructured.Unstructured, error) {
	containersPath := []string{"spec", "template", "spec", "containers"}
	log := b.logger.WithValues("Containers.NestedPath", containersPath)

	containers, found, err := unstructured.NestedSlice(obj.Object, containersPath...)
	if err != nil {
		return nil, err
	}
	if !found {
		err = fmt.Errorf("unable to find '%#v' in object kind '%s'",
			containersPath, obj.GetKind())
		log.Error(err, "is this definition supported by this operator?")
		return nil, err
	}

	if containers, err = b.updateContainers(containers); err != nil {
		return nil, err
	}
	if err = unstructured.SetNestedSlice(obj.Object, containers, containersPath...); err != nil {
		return nil, err
	}
	return obj, nil
}

// updateContainers execute the update command per container found.
func (b *Binder) updateContainers(
	containers []interface{},
) ([]interface{}, error) {
	var err error

	for i, container := range containers {
		log := b.logger.WithValues("Obj.Container.Number", i)
		log.Debug("Inspecting container...")

		containers[i], err = b.updateContainer(container)
		if err != nil {
			log.Error(err, "during container update.")
			return nil, err
		}
	}

	return containers, nil
}

func (b *Binder) appendEnvVar(envList []corev1.EnvVar, envParam string, envValue string) []corev1.EnvVar {
	var updatedEnvList []corev1.EnvVar

	alreadyPresent := false
	for _, env := range envList {
		if env.Name == envParam {
			env.Value = envValue
			alreadyPresent = true
		}
		updatedEnvList = append(updatedEnvList, env)
	}

	if !alreadyPresent {
		updatedEnvList = append(updatedEnvList, corev1.EnvVar{
			Name:  envParam,
			Value: envValue,
		})
	}
	return updatedEnvList
}

// appendEnvFrom based on secret name and list of EnvFromSource instances, making sure secret is
// part of the list or appended.
func (b *Binder) appendEnvFrom(envList []corev1.EnvFromSource, secret string) []corev1.EnvFromSource {
	log := b.logger
	for _, env := range envList {
		if env.SecretRef.Name == secret {
			log.Debug("Directive 'envFrom' is already present!")
			// secret name is already referenced
			return envList
		}
	}

	log.Debug("Adding 'envFrom' directive...")
	return append(envList, corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret,
			},
		},
	})
}

// updateContainer execute the update of a single container.
func (b *Binder) updateContainer(container interface{}) (map[string]interface{}, error) {
	c := corev1.Container{}
	u := container.(map[string]interface{})
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, &c)
	if err != nil {
		return nil, err
	}
	// effectively binding the application with intermediary secret
	c.EnvFrom = b.appendEnvFrom(c.EnvFrom, b.sbr.GetName())
	// c.Env = b.appendEnvVar(c.Env, lastboundparam, time.Now().Format(time.RFC3339))
	if len(b.volumeKeys) > 0 {
		// and adding volume mount entries
		c.VolumeMounts = b.appendVolumeMounts(c.VolumeMounts)
	}

	return runtime.DefaultUnstructuredConverter.ToUnstructured(&c)
}

// appendVolumeMounts append the binding volume in the template level.
func (b *Binder) appendVolumeMounts(volumeMounts []corev1.VolumeMount) []corev1.VolumeMount {
	name := b.sbr.GetName()
	mountPath := b.sbr.Spec.MountPathPrefix
	if mountPath == "" {
		mountPath = "/var/data"
	}

	for _, v := range volumeMounts {
		if name == v.Name {
			return volumeMounts
		}
	}

	return append(volumeMounts, corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	})
}

// nestedMapComparison compares a nested field from two objects.
func nestedMapComparison(a, b *unstructured.Unstructured, fields ...string) (bool, error) {
	var (
		aMap map[string]interface{}
		bMap map[string]interface{}
		ok   bool
		err  error
	)

	if aMap, ok, err = unstructured.NestedMap(a.Object, fields...); err != nil {
		return false, err
	} else if !ok {
		return false, fmt.Errorf("original object doesn't have a 'spec' field")
	}

	if bMap, ok, err = unstructured.NestedMap(b.Object, fields...); err != nil {
		return false, err
	} else if !ok {
		return false, fmt.Errorf("original object doesn't have a 'spec' field")
	}

	result := cmp.DeepEqual(aMap, bMap)()

	return result.Success(), nil
}

// update the list of objects informed as unstructured, looking for "containers" entry. This method
// loops over each container to inspect "envFrom" and append the intermediary secret, having the same
// name than original ServiceBindingRequest.
func (b *Binder) update(objs *unstructured.UnstructuredList) ([]*unstructured.Unstructured, error) {
	updatedObjs := []*unstructured.Unstructured{}

	for _, obj := range objs.Items {
		// store a copy of the original object to later be used in a comparison
		originalObj := obj.DeepCopy()
		name := obj.GetName()
		log := b.logger.WithValues("Obj.Name", name, "Obj.Kind", obj.GetKind())
		log.Debug("Inspecting object...")

		updatedObj, err := b.updateSpecContainers(&obj)
		if err != nil {
			return nil, err
		}

		if len(b.volumeKeys) > 0 {
			updatedObj, err = b.updateSpecVolumes(&obj)
			if err != nil {
				return nil, err
			}
		}

		if specsAreEqual, err := nestedMapComparison(originalObj, updatedObj, "spec"); err != nil {
			log.Error(err, "")
			continue
		} else if specsAreEqual {
			continue
		}

		log.Debug("Updating object...")
		if err := b.client.Update(b.ctx, updatedObj); err != nil {
			return nil, err
		}

		log.Debug("Reading back updated object...")
		// reading object back again, to comply with possible modifications
		namespacedName := types.NamespacedName{
			Namespace: updatedObj.GetNamespace(),
			Name:      updatedObj.GetName(),
		}
		if err = b.client.Get(b.ctx, namespacedName, updatedObj); err != nil {
			return nil, err
		}

		updatedObjs = append(updatedObjs, updatedObj)
	}

	return updatedObjs, nil
}

// Bind resources to intermediary secret, by searching informed ResourceKind containing the labels
// in ApplicationSelector, and then updating spec.
func (b *Binder) Bind() ([]*unstructured.Unstructured, error) {
	objs, err := b.search()
	if err != nil {
		return nil, err
	}
	return b.update(objs)
}

// NewBinder returns a new Binder instance.
func NewBinder(
	ctx context.Context,
	client client.Client,
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
	volumeKeys []string,
) *Binder {
	return &Binder{
		ctx:        ctx,
		client:     client,
		dynClient:  dynClient,
		sbr:        sbr,
		volumeKeys: volumeKeys,
		logger:     binderLog,
	}
}
