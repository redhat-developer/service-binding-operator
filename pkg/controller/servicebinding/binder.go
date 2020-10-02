package servicebinding

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
	knativev1 "knative.dev/serving/pkg/apis/serving/v1"
)

// changeTriggerEnv hijacking environment in order to trigger a change
const changeTriggerEnv = "ServiceBindingOperatorChangeTriggerEnvVar"

// binder executes the "binding" act of updating different application kinds to use intermediary
// secret. Those secrets should be offered as environment variables.
type binder struct {
	ctx        context.Context          // request context
	dynClient  dynamic.Interface        // kubernetes dynamic api client
	sbr        *v1alpha1.ServiceBinding // instantiated service binding request
	volumeKeys []string                 // list of key names used in volume mounts
	modifier   extraFieldsModifier      // extra modifier for CRDs before updating
	restMapper meta.RESTMapper          // RESTMapper to convert GVR from GVK
	logger     *log.Log                 // logger instance
}

// extraFieldsModifier is useful for updating backend service which requires additional changes besides
// env/volumes updating. eg. for knative service we need to remove or update `spec.template.metadata.name`
// from service template before updating otherwise it will be rejected.
type extraFieldsModifier interface {
	ModifyExtraFields(u *unstructured.Unstructured) error
}

// extraFieldsModifierFunc func receiver type for ExtraFieldsModifier
type extraFieldsModifierFunc func(u *unstructured.Unstructured) error

// ModifyExtraFields implements ExtraFieldsModifier interface
func (f extraFieldsModifierFunc) ModifyExtraFields(u *unstructured.Unstructured) error {
	return f(u)
}

// search objects based in Kind/APIVersion, which contain the labels defined in Application.
func (b *binder) search() (*unstructured.UnstructuredList, error) {
	// If Application name is present
	if b.sbr.Spec.Application.Name != "" {
		return b.getApplicationByName()
	} else if b.sbr.Spec.Application.LabelSelector != nil && b.sbr.Spec.Application.LabelSelector.MatchLabels != nil {
		return b.getApplicationByLabelSelector()
	} else {
		return nil, errEmptyApplication
	}
}

func (b *binder) getApplicationByName() (*unstructured.UnstructuredList, error) {
	ns := b.sbr.GetNamespace()
	gvr := schema.GroupVersionResource{
		Group:    b.sbr.Spec.Application.GroupVersionResource.Group,
		Version:  b.sbr.Spec.Application.GroupVersionResource.Version,
		Resource: b.sbr.Spec.Application.GroupVersionResource.Resource,
	}
	object, err := b.dynClient.Resource(gvr).Namespace(ns).
		Get(b.sbr.Spec.Application.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errApplicationNotFound
		}
		return nil, err
	}

	return &unstructured.UnstructuredList{Items: []unstructured.Unstructured{*object}}, nil
}

func (b *binder) getApplicationByLabelSelector() (*unstructured.UnstructuredList, error) {
	ns := b.sbr.GetNamespace()
	gvr := schema.GroupVersionResource{
		Group:    b.sbr.Spec.Application.GroupVersionResource.Group,
		Version:  b.sbr.Spec.Application.GroupVersionResource.Version,
		Resource: b.sbr.Spec.Application.GroupVersionResource.Resource,
	}
	matchLabels := b.sbr.Spec.Application.LabelSelector.MatchLabels
	opts := metav1.ListOptions{
		LabelSelector: labels.Set(matchLabels).String(),
	}

	objList, err := b.dynClient.Resource(gvr).Namespace(ns).List(opts)
	if err != nil {
		return nil, err
	}

	if len(objList.Items) == 0 {
		return nil, errApplicationNotFound
	}

	return objList, nil
}

// extractSpecVolumes based on volume path, extract it unstructured. It can return error on trying
// to find data in informed Unstructured object.
func (b *binder) extractSpecVolumes(obj *unstructured.Unstructured) ([]interface{}, error) {
	log := b.logger.WithValues("Volumes.NestedPath", b.getVolumesPath())
	log.Debug("Reading volumes definitions...")
	volumes, _, err := unstructured.NestedSlice(obj.Object, b.getVolumesPath()...)
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

// updateSpecVolumes execute the inspection and update "volumes" entries in informed spec.
func (b *binder) updateSpecVolumes(obj *unstructured.Unstructured) error {
	volumes, err := b.extractSpecVolumes(obj)
	if err != nil {
		return err
	}

	volumes, err = b.updateVolumes(volumes)
	if err != nil {
		return err
	}
	if err = unstructured.SetNestedSlice(obj.Object, volumes, b.getVolumesPath()...); err != nil {
		return err
	}
	return nil
}

// removeSpecVolumes based on extract volume subset, removing volume bind volume entry. It can return
// error on navigating though unstructured object, or in the case of having issues to edit
// unstructured resource.
func (b *binder) removeSpecVolumes(
	obj *unstructured.Unstructured,
) (*unstructured.Unstructured, error) {
	volumes, err := b.extractSpecVolumes(obj)
	if err != nil {
		return nil, err
	}
	volumes = b.removeVolumes(volumes)
	if err = unstructured.SetNestedSlice(obj.Object, volumes, b.getVolumesPath()...); err != nil {
		return nil, err
	}
	return obj, nil
}

// updateVolumes inspect informed list assuming as []corev1.Volume, and if binding volume is already
// defined just return the same list, otherwise, appending the binding volume.
func (b *binder) updateVolumes(volumes []interface{}) ([]interface{}, error) {
	name := b.sbr.GetName()
	log := b.logger

	// FIXME(isuttonl): update should not bail out here since b.volumeKeys might change
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
	bindVolume := &corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: name,
				Items:      items,
			},
		},
	}

	// making sure tranforming it back to unstructured before returning
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(bindVolume)
	if err != nil {
		return nil, err
	}
	return append(volumes, u), nil
}

// removeVolumes remove the bind volumes from informed list of unstructured volumes.
func (b *binder) removeVolumes(volumes []interface{}) []interface{} {
	name := b.sbr.GetName()
	var cleanVolumes []interface{}
	for _, v := range volumes {
		volume := v.(corev1.Volume)
		if name != volume.Name {
			cleanVolumes = append(cleanVolumes, v)
		}
	}
	return cleanVolumes
}

// extractSpecContainers search for
func (b *binder) extractSpecContainers(obj *unstructured.Unstructured) ([]interface{}, error) {
	log := b.logger.WithValues("Containers.NestedPath", b.getContainersPath())

	containers, found, err := unstructured.NestedSlice(obj.Object, b.getContainersPath()...)
	if err != nil {
		return nil, err
	}
	if !found {
		err = fmt.Errorf("unable to find '%#v' in object kind '%s'", b.getContainersPath(), obj.GetKind())
		log.Error(err, "is this definition supported by this operator?")
		return nil, err
	}
	return containers, nil
}

// updateSecretField extract the specific secret field from
// the object, and triggers an update.
func (b *binder) updateSecretField(obj *unstructured.Unstructured) error {
	return unstructured.SetNestedField(obj.Object, b.sbr.GetName(), b.getSecretFieldPath()...)
}

// updateSpecContainers extract containers from object, and trigger update.
func (b *binder) updateSpecContainers(obj *unstructured.Unstructured) error {
	containers, err := b.extractSpecContainers(obj)
	if err != nil {
		return err
	}
	if containers, err = b.updateContainers(containers); err != nil {
		return err
	}
	if err = unstructured.SetNestedSlice(obj.Object, containers, b.getContainersPath()...); err != nil {
		return err
	}
	return nil
}

func getContainersPath(applicationSelector *v1alpha1.Application) []string {
	return strings.Split(applicationSelector.BindingPath.ContainersPath, ".")
}

func getSecretFieldPath(applicationSelector *v1alpha1.Application) []string {
	return strings.Split(applicationSelector.BindingPath.SecretPath, ".")
}

func (b *binder) getContainersPath() []string {
	return getContainersPath(b.sbr.Spec.Application)
}

func (b *binder) getVolumesPath() []string {
	return []string{"spec", "template", "spec", "volumes"}
}

func (b *binder) getSecretFieldPath() []string {
	return getSecretFieldPath(b.sbr.Spec.Application)
}

// removeSpecContainers find and edit containers resource subset, removing bind related entries
// from the object. It can return error on extracting data, editing steps and final editing of to be
// returned object.
func (b *binder) removeSpecContainers(obj *unstructured.Unstructured) error {
	containers, err := b.extractSpecContainers(obj)
	if err != nil {
		return err
	}
	if containers, err = b.removeContainers(containers); err != nil {
		return err
	}
	if err = unstructured.SetNestedSlice(obj.Object, containers, b.getContainersPath()...); err != nil {
		return err
	}
	return nil
}

// updateContainers execute the update command per container found.
func (b *binder) updateContainers(containers []interface{}) ([]interface{}, error) {
	var err error

	for i, container := range containers {
		log := b.logger.WithValues("Obj.Container.Number", i)
		log.Debug("Inspecting container...")

		containers[i], err = b.updateContainer(container)
		if err != nil {
			log.Error(err, "during container update to add binding items.")
			return nil, err
		}
	}

	return containers, nil
}

// removeContainers execute removal of binding related entries in containers.
func (b *binder) removeContainers(containers []interface{}) ([]interface{}, error) {
	var err error

	for i, container := range containers {
		log := b.logger.WithValues("Obj.Container.Number", i)
		log.Debug("Inspecting container...")

		containers[i], err = b.removeContainer(container)
		if err != nil {
			log.Error(err, "during container update to remove binding items.")
			return nil, err
		}
	}
	return containers, nil
}

// appendEnvVar append a single environment variable onto informed "EnvVar" instance.
func (b *binder) appendEnvVar(
	envList []corev1.EnvVar,
	envParam string,
	envValue string,
) []corev1.EnvVar {
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
func (b *binder) appendEnvFrom(envList []corev1.EnvFromSource, secret string) []corev1.EnvFromSource {
	for _, env := range envList {
		if env.SecretRef != nil && env.SecretRef.Name == secret {
			b.logger.Debug("Directive 'envFrom' is already present!")
			// secret name is already referenced
			return envList
		}
	}

	b.logger.Debug("Adding 'envFrom' directive...")
	return append(envList, corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secret,
			},
		},
	})
}

// removeEnvFrom remove bind related entry from slice of "EnvFromSource".
func (b *binder) removeEnvFrom(envList []corev1.EnvFromSource, secret string) []corev1.EnvFromSource {
	var cleanEnvList []corev1.EnvFromSource
	for _, env := range envList {
		if env.SecretRef != nil && env.SecretRef.Name != secret {
			cleanEnvList = append(cleanEnvList, env)
		}
	}
	return cleanEnvList
}

// containerFromUnstructured based on informed unstructured corev1.Container, convert it back to the
// original type. It can return errors on the process.
func (b *binder) containerFromUnstructured(container interface{}) (*corev1.Container, error) {
	c := &corev1.Container{}
	u := container.(map[string]interface{})
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// updateContainer execute the update of a single container, adding binding items.
func (b *binder) updateContainer(container interface{}) (map[string]interface{}, error) {
	c, err := b.containerFromUnstructured(container)
	if err != nil {
		return nil, err
	}

	// effectively binding the application with intermediary secret
	c.EnvFrom = b.appendEnvFrom(c.EnvFrom, b.sbr.GetName())

	secretRes := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	existingSecret, err := b.dynClient.Resource(secretRes).Namespace(b.sbr.GetNamespace()).Get(b.sbr.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// add a special environment variable that is only used to trigger a change in the declaration,
	// attempting to force a side effect (in case of a Deployment, it would result in its Pods to be
	// restarted)
	c.Env = b.appendEnvVar(c.Env, changeTriggerEnv, existingSecret.GetResourceVersion())

	if len(b.volumeKeys) > 0 {
		// and adding volume mount entries
		c.VolumeMounts = b.appendVolumeMounts(c.VolumeMounts)
	}

	return runtime.DefaultUnstructuredConverter.ToUnstructured(c)
}

// removeContainer execute the update of single container to remove binding items.
func (b *binder) removeContainer(container interface{}) (map[string]interface{}, error) {
	c, err := b.containerFromUnstructured(container)
	if err != nil {
		return nil, err
	}

	// removing intermediary secret, effectively unbinding the application
	c.EnvFrom = b.removeEnvFrom(c.EnvFrom, b.sbr.GetName())

	if len(b.volumeKeys) > 0 {
		// removing volume mount entries
		c.VolumeMounts = b.removeVolumeMounts(c.VolumeMounts)
	}

	return runtime.DefaultUnstructuredConverter.ToUnstructured(c)
}

// appendVolumeMounts append the binding volume in the template level.
func (b *binder) appendVolumeMounts(volumeMounts []corev1.VolumeMount) []corev1.VolumeMount {
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

// removeVolumeMounts from informed slice of corev1.VolumeMount, make sure all binding related
// entries won't be part of returned slice.
func (b *binder) removeVolumeMounts(volumeMounts []corev1.VolumeMount) []corev1.VolumeMount {
	var cleanVolumeMounts []corev1.VolumeMount
	name := b.sbr.GetName()
	for _, v := range volumeMounts {
		if name != v.Name {
			cleanVolumeMounts = append(cleanVolumeMounts, v)
		}
	}
	return cleanVolumeMounts
}

type comparisonResult struct {
	Diff    string
	Success bool
}

// nestedUnstructuredComparison compares a nested field from two objects.
func nestedUnstructuredComparison(a, b *unstructured.Unstructured, fields ...string) (*comparisonResult, error) {
	var (
		aMap map[string]interface{}
		bMap map[string]interface{}
		aOk  bool
		bOk  bool
		err  error
	)

	if aMap, aOk, err = unstructured.NestedMap(a.Object, fields...); err != nil {
		return nil, err
	}

	if bMap, bOk, err = unstructured.NestedMap(b.Object, fields...); err != nil {
		return nil, err
	}

	if aOk != bOk {
		return nil, fmt.Errorf("path should exist in both objects: %v", fields)
	}

	return nestedMapComparison(aMap, bMap), nil
}

func nestedMapComparison(a, b map[string]interface{}) *comparisonResult {
	diff := cmp.Diff(a, b)
	if len(diff) != 0 {
		return &comparisonResult{Success: false, Diff: diff}
	}
	return &comparisonResult{Success: true}
}

// update the list of objects informed as unstructured, looking for "containers" entry. This method
// loops over each container to inspect "envFrom" and append the intermediary secret, having the same
// name than original ServiceBinding.
func (b *binder) update(objs *unstructured.UnstructuredList) ([]*unstructured.Unstructured, error) {
	updatedObjs := []*unstructured.Unstructured{}

	for _, obj := range objs.Items {
		// modify the copy of the original object and use the original one later for comparison
		updatedObj := obj.DeepCopy()
		name := obj.GetName()
		log := b.logger.WithValues("Obj.Name", name, "Obj.Kind", obj.GetKind())
		log.Debug("Inspecting object...")

		sbrNamespacedName := types.NamespacedName{Namespace: b.sbr.GetNamespace(), Name: b.sbr.GetName()}
		updatedObj = setSBRAnnotations(sbrNamespacedName, updatedObj)
		var err error
		if b.sbr.Spec.Application.BindingPath.SecretPath != "" {
			err = b.updateSecretField(updatedObj)
			if err != nil {
				return nil, err
			}
		}

		if b.sbr.Spec.Application.BindingPath.ContainersPath != "" {
			err = b.updateSpecContainers(updatedObj)
			if err != nil {
				return nil, err
			}
		}

		if len(b.volumeKeys) > 0 {
			if err = b.updateSpecVolumes(updatedObj); err != nil {
				return nil, err
			}
		}

		if specsAreEqual, err := nestedUnstructuredComparison(&obj, updatedObj); err != nil {
			log.Error(err, "Error comparing previous and updated object")
			continue
		} else if specsAreEqual.Success {
			log.Debug("Previous and updated object have same spec, skipping")
			continue
		}
		if b.modifier != nil {
			err = b.modifier.ModifyExtraFields(updatedObj)
			if err != nil {
				return nil, err
			}
		}
		log.Debug("Updating object...")
		gk := updatedObj.GroupVersionKind().GroupKind()
		version := updatedObj.GroupVersionKind().Version
		mapping, err := b.restMapper.RESTMapping(gk, version)
		if err != nil {
			return nil, err
		}
		updated, err := b.dynClient.Resource(mapping.Resource).
			Namespace(updatedObj.GetNamespace()).
			Update(updatedObj, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
		updatedObjs = append(updatedObjs, updated)
	}
	return updatedObjs, nil
}

// remove attempts to update each given object without any service binding related information.
func (b *binder) remove(objs *unstructured.UnstructuredList) error {
	for _, obj := range objs.Items {
		name := obj.GetName()
		logger := b.logger.WithValues("Obj.Name", name, "Obj.Kind", obj.GetKind())
		logger.Debug("Inspecting object...")
		updatedObj := obj.DeepCopy()
		err := b.removeSpecContainers(updatedObj)
		if err != nil {
			return err
		}

		if len(b.volumeKeys) > 0 {
			if updatedObj, err = b.removeSpecVolumes(updatedObj); err != nil {
				return err
			}
		}

		gk := updatedObj.GroupVersionKind().GroupKind()
		version := updatedObj.GroupVersionKind().Version
		mapping, err := b.restMapper.RESTMapping(gk, version)
		if err != nil {
			return err
		}

		_, err = b.dynClient.Resource(mapping.Resource).
			Namespace(updatedObj.GetNamespace()).
			Update(updatedObj, metav1.UpdateOptions{})

		if err != nil {
			return err
		}

	}
	return nil
}

// unbind select objects subject to binding, and proceed with "remove", which will unbind objects.
func (b *binder) unbind() error {
	objs, err := b.search()
	if err != nil {
		return err
	}
	return b.remove(objs)
}

// bind resources to intermediary secret, by searching informed ResourceKind containing the labels
// in Application, and then updating spec.
func (b *binder) bind() ([]*unstructured.Unstructured, error) {
	objs, err := b.search()
	if err != nil {
		return nil, err
	}
	return b.update(objs)
}

// newBinder returns a new Binder instance.
func newBinder(
	ctx context.Context,
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBinding,
	volumeKeys []string,
	restMapper meta.RESTMapper,
) *binder {

	logger := log.NewLog("binder")
	modifier := buildExtraFieldsModifier(logger, sbr)

	return &binder{
		ctx:        ctx,
		dynClient:  dynClient,
		sbr:        sbr,
		volumeKeys: volumeKeys,
		modifier:   modifier,
		restMapper: restMapper,
		logger:     logger,
	}
}

func buildExtraFieldsModifier(logger *log.Log, sbr *v1alpha1.ServiceBinding) extraFieldsModifier {
	if sbr.Spec.Application != nil {
		gvr := sbr.Spec.Application.GroupVersionResource
		ksvcgvr := knativev1.SchemeGroupVersion.WithResource("services")
		switch gvr.String() {
		case ksvcgvr.String():
			pathToRevisionName := "spec.template.metadata.name"
			return extraFieldsModifierFunc(func(u *unstructured.Unstructured) error {
				revisionName, ok, err := unstructured.NestedString(u.Object, strings.Split(pathToRevisionName, ".")...)
				if err == nil && ok {
					logger.Info("remove revision in knative service template", "name", revisionName)
					unstructured.RemoveNestedField(u.Object, strings.Split(pathToRevisionName, ".")...)
				}
				return nil
			})
		default:
			return nil
		}
	}
	return nil
}
