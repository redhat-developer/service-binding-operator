package controllers

import (
	"context"
	"errors"
	"github.com/redhat-developer/service-binding-operator/pkg/naming"

	"k8s.io/apimachinery/pkg/api/meta"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
)

// getServiceBinding retrieve the SBR object based on namespaced-name.
func (r *ServiceBindingReconciler) getServiceBinding(
	namespacedName types.NamespacedName,
) (*v1alpha1.ServiceBinding, error) {
	gr := v1alpha1.GroupVersionResource
	resourceClient := r.dynClient.Resource(gr).Namespace(namespacedName.Namespace)
	u, err := resourceClient.Get(context.TODO(), namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	sbr := &v1alpha1.ServiceBinding{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}
	if sbr.Spec.DetectBindingResources == nil {
		falseBool := false
		sbr.Spec.DetectBindingResources = &falseBool
	}
	return sbr, nil
}

// Reconcile a ServiceBinding by the following steps:
// 1. Inspecting SBR in order to identify backend service. The service is composed by a CRD name and
//    kind, and by inspecting "connects-to" label identify the name of service instance;
// 2. Using OperatorLifecycleManager standards, identifying which items are intersting for binding
//    by parsing CustomResourceDefinitionDescripton object. Alternatively, this informmation may
// 	  also come from special annotations in the CR/CRD;
// 3. Search and read contents identified in previous step, creating an intermediary secret to hold
//    data formatted as environment variables key/value;
// 4. Search applications that are interested to bind with given service, by inspecting labels. The
//    Deployment (and other kinds) will be updated in "spec" level.
func (r *ServiceBindingReconciler) doReconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := reconcilerLog.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name,
	)

	logger.Info("Reconciling ServiceBinding...")

	// fetch and validate namespaced ServiceBinding instance
	sbr, err := r.getServiceBinding(request.NamespacedName)
	if err != nil {
		if errors.Is(err, errApplicationNotFound) {
			logger.Info("SBR deleted after application deletion")
			return done()
		}
		logger.Error(err, "On retrieving service-binding instance.")
		return doneOnNotFound(err)
	}

	// validate namespaced ServiceBinding instance (this check has been disabled until test data has been
	// adjusted to reflect the validation)
	//
	//if err = r.validateServiceBinding(sbr); err != nil {
	//	logger.Error(err, "On validating service-binding instance.")
	//	return Done()
	//}

	logger = logger.WithValues("ServiceBinding.Name", sbr.Name)
	logger.Debug("Found service binding request to inspect")

	ctx := context.Background()

	if len(sbr.Spec.Services) == 0 {
		_, updateErr := updateServiceBindingStatus(
			r.dynClient,
			sbr,
			metav1.Condition{
				Type:    v1alpha1.CollectionReady,
				Status:  metav1.ConditionFalse,
				Reason:  v1alpha1.EmptyServiceSelectorsReason,
				Message: errEmptyServices.Error(),
			},
			metav1.Condition{
				Type:   v1alpha1.InjectionReady,
				Reason: v1alpha1.EmptyServiceSelectorsReason,
				Status: metav1.ConditionFalse,
			},
			metav1.Condition{
				Type:   v1alpha1.BindingReady,
				Reason: v1alpha1.EmptyServiceSelectorsReason,
				Status: metav1.ConditionFalse,
			},
		)
		if updateErr == nil {
			return done()
		}
		// TODO: do not requeue here
		//
		// Since there are nothing to recover from in the case service selectors is empty, it is
		// still required to requeue due to some watches not being implemented. This is known issue
		// being worked in https://github.com/redhat-developer/service-binding-operator/pull/442.
		return requeueError(errEmptyServices)
	}
	serviceCtxs := serviceContextList{}
	// we only need to build the service context if SBR is not deleted
	if sbr.GetDeletionTimestamp() == nil {
		serviceCtxs, err = buildServiceContexts(
			logger.WithName("buildServiceContexts"),
			r.dynClient,
			sbr.GetNamespace(),
			sbr.Spec.Services,
			sbr.Spec.DetectBindingResources,
			sbr.Spec.BindAsFiles,
			r,
			sbr.Spec.NamingTemplate(),
		)
		if err != nil {
			//handle service not found error
			if k8serrors.IsNotFound(err) {
				err = updateSBRConditions(r.dynClient, sbr,
					metav1.Condition{
						Type:    v1alpha1.CollectionReady,
						Status:  metav1.ConditionFalse,
						Reason:  v1alpha1.ServiceNotFoundReason,
						Message: err.Error(),
					},
					metav1.Condition{
						Type:   v1alpha1.InjectionReady,
						Status: metav1.ConditionFalse,
					},
					metav1.Condition{
						Type:   v1alpha1.BindingReady,
						Status: metav1.ConditionFalse,
					},
				)
				if err != nil {
					logger.Error(err, "Failed to update SBR conditions", "sbr", sbr)
				}
			}
			return requeueError(err)
		}
	}
	binding, err := buildBinding(
		r.dynClient,
		sbr.Spec.Mappings,
		serviceCtxs,
		sbr.Spec.NamingTemplate(),
	)
	if err != nil {
		if errors.Is(err, naming.TemplateError) {
			err = updateSBRConditions(r.dynClient, sbr,
				metav1.Condition{
					Type:    v1alpha1.CollectionReady,
					Status:  metav1.ConditionFalse,
					Reason:  v1alpha1.NamingStrategyError,
					Message: err.Error(),
				},
			)
			if err != nil {
				logger.Error(err, "Failed to update SBR conditions", "sbr", sbr)
			}
			return done()
		}
		return requeueError(err)
	}

	options := &serviceBinderOptions{
		dynClient:              r.dynClient,
		detectBindingResources: *sbr.Spec.DetectBindingResources,
		sbr:                    sbr,
		logger:                 logger,
		objects:                serviceCtxs.getServices(),
		binding:                binding,
		typeLookup:             r,
	}

	sb, err := buildServiceBinder(ctx, options)
	if err != nil {
		// BuildServiceBinder can return only InvalidOptionsErr, and it is a programmer's error so
		// just bail out without re-queueing nor updating conditions.
		logger.Error(err, "Building ServiceBinder")
		return noRequeue(err)
	}

	if sbr.Spec.Application != nil {
		gvr, err := r.ResourceForReferable(sbr.Spec.Application)
		if err != nil {
			logger.Error(err, "Error getting application GVR")
		} else {
			err = r.resourceWatcher.AddWatchForGVR(*gvr)
			if err != nil {
				logger.Error(err, "Error add watching application GVR")
			}
		}
	}

	if sbr.Spec.Services != nil {
		for _, service := range sbr.Spec.Services {
			gvr, err := r.ResourceForReferable(&service)
			if err != nil {
				logger.Error(err, "Error getting backing service GVR")
			}

			err = r.resourceWatcher.AddWatchForGVR(*gvr)
			if err != nil {
				logger.Error(err, "Error add watching backing service GVK")
			}
		}
	}

	if sbr.GetDeletionTimestamp() != nil && sbr.GetOwnerReferences() != nil {
		logger := logger.WithName("Deleting SBR when it has ownerReference")
		logger.Debug("Removing resource finalizers...")
		removeFinalizer(sbr)
		if _, err := updateServiceBinding(r.dynClient, sbr); err != nil {
			return requeueError(err)
		}
		return done()
	}
	if sbr.GetDeletionTimestamp() != nil {
		logger := logger.WithName("unbind")
		logger.Info("Executing unbinding steps...")
		return sb.unbind()
	}

	logger.Info("Binding applications with intermediary secret...")
	return sb.bind()
}

func updateSBRConditions(dynClient dynamic.Interface, sbr *v1alpha1.ServiceBinding, conditions ...metav1.Condition) error {
	for _, v := range conditions {
		meta.SetStatusCondition(&sbr.Status.Conditions, v)
	}
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return err
	}

	nsClient := dynClient.
		Resource(v1alpha1.GroupVersionResource).
		Namespace(sbr.GetNamespace())

	_, err = nsClient.UpdateStatus(context.TODO(), u, metav1.UpdateOptions{})

	return err
}
