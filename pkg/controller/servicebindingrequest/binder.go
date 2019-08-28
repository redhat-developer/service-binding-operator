package servicebindingrequest

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

const (
	lastboundparam = "lastbound"
)

// Binder executes the "binding" act of updating different application kinds to use intermediary
// secret. Those secrets should be offered as environment variables.
type Binder struct {
	ctx        context.Context                 // request context
	client     client.Client                   // kubernetes API client
	dynClient  dynamic.Interface               // kubernetes dynamic api client
	sbr        *v1alpha1.ServiceBindingRequest // instantiated service binding request
	volumeKeys []string                        // list of key names used in volume mounts
	logger     logr.Logger                     // logger instance
}

// search objects based in Kind/APIVersion, which contain the labels defined in ApplicationSelector.
func (b *Binder) search() (*ustrv1.UnstructuredList, error) {
	ns := b.sbr.GetNamespace()
	gvr := schema.GroupVersionResource{
		Group:    b.sbr.Spec.ApplicationSelector.Group,
		Version:  b.sbr.Spec.ApplicationSelector.Version,
		Resource: b.sbr.Spec.ApplicationSelector.Resource,
	}
	matchLabels := b.sbr.Spec.ApplicationSelector.MatchLabels
	opts := metav1.ListOptions{
		LabelSelector: labels.Set(matchLabels).String(),
	}

	objList, err := b.dynClient.Resource(gvr).Namespace(ns).List(opts)
	if err != nil {
		return nil, err
	}
	// Return fake NotFound error explicitly to ensure requeue when objList(^) is empty.
	if len(objList.Items) == 0 {
		return nil , errors.NewNotFound(
			gvr.GroupResource(),
			b.sbr.Spec.ApplicationSelector.Resource,
		)
	}
	return objList, err
}

// updateSpecVolumes execute the inspection and update "volumes" entries in informed spec.
func (b *Binder) updateSpecVolumes(
	logger logr.Logger,
	obj *ustrv1.Unstructured,
) (*ustrv1.Unstructured, error) {
	volumesPath := []string{"spec", "template", "spec", "volumes"}
	logger = logger.WithValues("Volumes.NestedPath", volumesPath)

	logger.Info("Reading volumes definitions...")
	volumes, _, err := ustrv1.NestedSlice(obj.Object, volumesPath...)
	if err != nil {
		return nil, err
	}
	logger.WithValues("Volumes", len(volumes)).Info("Amount of volumes in spec.")

	volumes, err = b.updateVolumes(logger, volumes)
	if err != nil {
		return nil, err
	}
	if err = ustrv1.SetNestedSlice(obj.Object, volumes, volumesPath...); err != nil {
		return nil, err
	}

	return obj, nil
}

// updateVolumes inspect informed list assuming as []corev1.Volume, and if binding volume is already
// defined just return the same list, otherwise, appending the binding volume.
func (b *Binder) updateVolumes(logger logr.Logger, volumes []interface{}) ([]interface{}, error) {
	name := b.sbr.GetName()

	logger.Info("Checking if binding volume is already defined...")
	for _, v := range volumes {
		volume := v.(corev1.Volume)
		if name == volume.Name {
			logger.Info("Volume is already defined!")
			return volumes, nil
		}
	}

	items := []corev1.KeyToPath{}
	for _, k := range b.volumeKeys {
		items = append(items, corev1.KeyToPath{Key: k, Path: k})
	}

	logger.WithValues("Items", items).Info("Appending new volume with items.")
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
	logger logr.Logger,
	obj *ustrv1.Unstructured,
) (*ustrv1.Unstructured, error) {
	containersPath := []string{"spec", "template", "spec", "containers"}
	logger = logger.WithValues("Containers.NestedPath", containersPath)

	containers, found, err := ustrv1.NestedSlice(obj.Object, containersPath...)
	if err != nil {
		return nil, err
	}
	if !found {
		err = fmt.Errorf("unable to find '%#v' in object kind '%s'",
			containersPath, obj.GetKind())
		logger.Error(err, "is this definition supported by this operator?")
		return nil, err
	}

	if containers, err = b.updateContainers(logger, containers); err != nil {
		return nil, err
	}
	if err = ustrv1.SetNestedSlice(obj.Object, containers, containersPath...); err != nil {
		return nil, err
	}
	return obj, nil
}

// updateContainers execute the update command per container found.
func (b *Binder) updateContainers(
	logger logr.Logger,
	containers []interface{},
) ([]interface{}, error) {
	var err error

	for i, container := range containers {
		logger := logger.WithValues("Obj.Container.Number", i)
		logger.Info("Inspecting container...")

		containers[i], err = b.updateContainer(container)
		if err != nil {
			logger.Error(err, "during container update.")
			return nil, err
		}
	}

	return containers, nil
}

func (b *Binder) appendEnvVar(envList []corev1.EnvVar, envParam string, envValue string) []corev1.EnvVar {
	return append(envList, corev1.EnvVar{
		Name:  envParam,
		Value: envValue,
	})
}

// appendEnvFrom based on secret name and list of EnvFromSource instances, making sure secret is
// part of the list or appended.
func (b *Binder) appendEnvFrom(envList []corev1.EnvFromSource, secret string) []corev1.EnvFromSource {
	for _, env := range envList {
		if env.SecretRef.Name == secret {
			b.logger.Info("Directive 'envFrom' is already present!")
			// secret name is already referenced
			return envList
		}
	}

	b.logger.Info("Adding 'envFrom' directive...")
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
	c.Env = b.appendEnvVar(c.Env, lastboundparam, time.Now().Format(time.RFC3339))
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

// update the list of objects informed as unstructured, looking for "containers" entry. This method
// loops over each container to inspect "envFrom" and append the intermediary secret, having the same
// name than original ServiceBindingRequest.
func (b *Binder) update(objList *ustrv1.UnstructuredList) ([]string, error) {
	var updatedObjectNames []string

	for _, obj := range objList.Items {
		name := obj.GetName()
		logger := b.logger.WithValues("Obj.Name", name, "Obj.Kind", obj.GetKind())
		logger.Info("Inspecting object...")

		updatedObj, err := b.updateSpecContainers(logger, &obj)
		if err != nil {
			return nil, err
		}

		if len(b.volumeKeys) > 0 {
			updatedObj, err = b.updateSpecVolumes(logger, &obj)
			if err != nil {
				return nil, err
			}
		}

		logger.Info("Updating object in Kube...")
		if err := b.client.Update(b.ctx, updatedObj); err != nil {
			return nil, err
		}

		// recording object as updated
		updatedObjectNames = append(updatedObjectNames, name)
	}
	return updatedObjectNames, nil
}

// Bind resources to intermediary secret, by searching informed ResourceKind containing the labels
// in ApplicationSelector, and then updating spec.
func (b *Binder) Bind() ([]string, error) {
	objList, err := b.search()
	if err != nil {
		return nil, err
	}

	return b.update(objList)
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
		logger:     logf.Log.WithName("binder"),
	}
}
