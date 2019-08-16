package servicebindingrequest

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	osappsv1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Binder executes the "binding" act of updating different application kinds to use intermediary
// secret. Those secrets should be offered as environment variables.
type Binder struct {
	ctx       context.Context                 // request context
	client    client.Client                   // kubernetes API client
	dynClient dynamic.Interface               // kubernetes dynamic api client
	sbr       *v1alpha1.ServiceBindingRequest // instantiated service binding request
	logger    logr.Logger                     // logger instance
}

func (b *Binder) getResourceKind() string {
	return strings.ToLower(b.sbr.Spec.ApplicationSelector.ResourceKind)
}

func (b *Binder) getListGVR() (schema.GroupVersionResource, error) {
	kind := b.getResourceKind()
	switch kind {
	case "deploymentconfig":
		return osappsv1.Resource("deploymentconfigs").WithVersion("v1"), nil
	case "deployment":
		return appsv1.Resource("deployments").WithVersion("v1"), nil
	default:
		return schema.GroupVersionResource{},
			fmt.Errorf("resource kind '%s' is not supported by this operator", kind)
	}
}

// search objects based in Kind/APIVersion, which contain the labels defined in ApplicationSelector.
func (b *Binder) search() (*ustrv1.UnstructuredList, error) {
	gvr, err := b.getListGVR()
	if err != nil {
		return nil, err
	}

	ns := b.sbr.GetNamespace()
	matchLabels := b.sbr.Spec.ApplicationSelector.MatchLabels
	opts := metav1.ListOptions{
		LabelSelector: labels.Set(matchLabels).String(),
	}

	return b.dynClient.Resource(gvr).Namespace(ns).List(opts)
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

// update the list of objects informed as unstructured, looking for "containers" entry. This method
// loops over each container to inspect "envFrom" and append the intermediary secret, having the same
// name than original ServiceBindingRequest.
func (b *Binder) update(objList *ustrv1.UnstructuredList) error {
	// TODO: does this path work for every k8s resource supported by this operator?
	nestedPath := []string{"spec", "template", "spec", "containers"}
	logger := b.logger.WithValues("nestedPath", nestedPath)

	for _, obj := range objList.Items {
		logger = logger.WithValues("Obj.Name", obj.GetName(), "Obj.Kind", obj.GetKind())
		logger.Info("Inspecting object...")

		containers, found, err := ustrv1.NestedSlice(obj.Object, nestedPath...)
		if err != nil {
			return err
		}
		if !found {
			err = fmt.Errorf("unable to find '%#v' in object kind '%s'", nestedPath, obj.GetKind())
			logger.Error(err, "is this definition supported by this operator?")
			return err
		}

		for i, container := range containers {
			logger = logger.WithValues("Obj.Container.Number", i)
			logger.Info("Inspecting container...")

			c := corev1.Container{}
			u := container.(map[string]interface{})
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, &c); err != nil {
				return err
			}

			// effectively binding the application with intermediary secret
			logger.Info("Binding application!")
			c.EnvFrom = b.appendEnvFrom(c.EnvFrom, b.sbr.GetName())

			bindContainer, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&c)
			if err != nil {
				return err
			}
			containers[i] = bindContainer
		}

		if err = ustrv1.SetNestedSlice(obj.Object, containers, nestedPath...); err != nil {
			return err
		}

		logger.Info("Updating object...")
		if err = b.client.Update(b.ctx, &obj); err != nil {
			return err
		}
	}

	return nil
}

// Bind resources to intermediary secret, by searching informed ResourceKind containing the labels
// in ApplicationSelector, and then updating spec.
func (b *Binder) Bind() error {
	objList, err := b.search()
	if err != nil {
		return err
	}
	return b.update(objList)
}

// NewBinder returns a new Binder instance.
func NewBinder(
	ctx context.Context,
	client client.Client,
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
) *Binder {
	return &Binder{
		ctx:       ctx,
		client:    client,
		dynClient: dynClient,
		sbr:       sbr,
		logger:    logf.Log.WithName("binder"),
	}
}
