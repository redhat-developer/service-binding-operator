package webhooks

import (
	"context"
	"encoding/json"

	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/apis/spec/v1beta1"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type MappingValidator struct {
	client dynamic.Interface
	lookup kubernetes.K8STypeLookup
}

// log is for logging in this package.
var log = logf.Log.WithName("WebHook Spec ClusterWorkloadResourceMapping")

func NewMappingValidator(config *rest.Config, mapper meta.RESTMapper) (*MappingValidator, error) {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	validator := &MappingValidator{
		client: client,
		lookup: kubernetes.ResourceLookup(mapper),
	}

	return validator, nil
}

func (validator *MappingValidator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1beta1.ClusterWorkloadResourceMapping{}).
		WithValidator(validator).
		Complete()
}

var _ webhook.CustomValidator = &MappingValidator{}

func (validator *MappingValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	mapping := obj.(*v1beta1.ClusterWorkloadResourceMapping)
	err := mapping.ValidateCreate()
	if err != nil {
		log.Error(err, "Error validating mapping (create)", "mapping", mapping.Name)
	}
	return err
}

func (validator *MappingValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	mapping := newObj.(*v1beta1.ClusterWorkloadResourceMapping)
	if err := mapping.ValidateCreate(); err != nil {
		log.Error(err, "Error validating mapping (update)", "mapping", mapping.Name)
		return err
	}
	oldMapping := oldObj.(*v1beta1.ClusterWorkloadResourceMapping)
	err := Serialize(ctx, oldMapping, validator.client, validator.lookup)
	if err != nil {
		return err
	}

	return nil
}

func (validator *MappingValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func Serialize(ctx context.Context, mapping *v1beta1.ClusterWorkloadResourceMapping, client dynamic.Interface, lookup kubernetes.K8STypeLookup) error {
	serialized, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	numItems := 0

	gvr := v1beta1.GroupVersionResource
	data, err := client.Resource(gvr).List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, binding := range data.Items {
		// we should filter out service bindings that the mapping doesn't affect.
		sb := v1beta1.ServiceBinding{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(binding.Object, &sb)
		if err != nil {
			// short circuit, something's gone terribly wrong
			return err
		}

		gvk, _ := sb.Spec.Workload.GroupVersionKind()
		workloadGVR, err := lookup.ResourceForKind(*gvk)
		if err != nil {
			return err
		}
		if !mapping.AcceptsGVR(workloadGVR) {
			// not a relevant binding, skip it
			continue
		}

		annotations := binding.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{
				apis.MappingAnnotationKey: string(serialized),
			}
		} else {
			annotations[apis.MappingAnnotationKey] = string(serialized)
		}
		binding.SetAnnotations(annotations)

		x, err := client.Resource(gvr).Namespace(sb.Namespace).Update(ctx, &binding, v1.UpdateOptions{})
		if err != nil {
			return err
		}
		log.Info("deployed service binding", "annotations", x.GetAnnotations())
		numItems += 1
	}

	gvr = v1alpha1.GroupVersionResource
	data, err = client.Resource(gvr).List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}

	for _, binding := range data.Items {
		sb := v1alpha1.ServiceBinding{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(binding.Object, &sb)
		if err != nil {
			return err
		}

		var workloadGVR *schema.GroupVersionResource
		if sb.Spec.Application.Kind != "" {
			gvk, _ := sb.Spec.Application.GroupVersionKind()
			if workloadGVR, err = lookup.ResourceForKind(*gvk); err != nil {
				return err
			}
		} else {
			workloadGVR, _ = sb.Spec.Application.GroupVersionResource()
		}
		if !mapping.AcceptsGVR(workloadGVR) {
			continue
		}

		annotations := binding.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{
				apis.MappingAnnotationKey: string(serialized),
			}
		} else {
			annotations[apis.MappingAnnotationKey] = string(serialized)
		}
		binding.SetAnnotations(annotations)

		x, err := client.Resource(gvr).Namespace(sb.Namespace).Update(ctx, &binding, v1.UpdateOptions{})
		if err != nil {
			return err
		}
		log.Info("deployed service binding", "annotations", x.GetAnnotations())
		numItems += 1
	}
	log.Info("Rebinding", "num_objects", numItems)
	return nil
}
