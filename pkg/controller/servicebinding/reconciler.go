package servicebinding

import (
	"context"
	"errors"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	// BindingReady indicates that the overall sbr succeeded
	BindingReady conditionsv1.ConditionType = "Ready"
	// CollectionReady indicates readiness for collection and persistance of intermediate manifests
	CollectionReady conditionsv1.ConditionType = "CollectionReady"
	// InjectionReady indicates readiness to change application manifests to use those intermediate manifests
	// If status is true, it indicates that the binding succeeded
	InjectionReady conditionsv1.ConditionType = "InjectionReady"
	// EmptyServiceSelectorsReason is used when the ServiceBinding has empty
	// services.
	EmptyServiceSelectorsReason = "EmptyServiceSelectors"
	// EmptyApplicationReason is used when the ServiceBinding has empty
	// application.
	EmptyApplicationReason = "EmptyApplication"
	// ApplicationNotFoundReason is used when the application is not found.
	ApplicationNotFoundReason = "ApplicationNotFound"
	// ServiceNotFoundReason is used when the service is not found.
	ServiceNotFoundReason = "ServiceNotFound"
)

// Reconciler reconciles a ServiceBinding object
type reconciler struct {
	dynClient       dynamic.Interface // kubernetes dynamic api client
	scheme          *runtime.Scheme   // api scheme
	restMapper      meta.RESTMapper   // restMapper to convert GVK and GVR
	resourceWatcher ResourceWatcher   // ResourceWatcher to add watching for specific GVK/GVR
}

// reconcilerLog local logger instance
var reconcilerLog = log.NewLog("reconciler")

//// validateServiceBinding check for unsupported settings in SBR.
//func (r *Reconciler) validateServiceBinding(sbr *v1alpha1.ServiceBinding) error {
//	// check if application Name and MatchLabels, one of them is required.
//	if sbr.Spec.Application.Name == "" &&
//		sbr.Spec.Application.LabelSelector == nil {
//		return fmt.Errorf("both Name and LabelSelector are not set")
//	}
//	return nil
//}

// getServiceBinding retrieve the SBR object based on namespaced-name.
func (r *reconciler) getServiceBinding(
	namespacedName types.NamespacedName,
) (*v1alpha1.ServiceBinding, error) {
	gr := v1alpha1.SchemeGroupVersion.WithResource(serviceBindingRequestResource)
	resourceClient := r.dynClient.Resource(gr).Namespace(namespacedName.Namespace)
	u, err := resourceClient.Get(namespacedName.Name, metav1.GetOptions{})
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
func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
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
			conditionsv1.Condition{
				Type:    CollectionReady,
				Status:  corev1.ConditionFalse,
				Reason:  EmptyServiceSelectorsReason,
				Message: errEmptyServices.Error(),
			},
			conditionsv1.Condition{
				Type:   InjectionReady,
				Status: corev1.ConditionFalse,
			},
			conditionsv1.Condition{
				Type:   BindingReady,
				Status: corev1.ConditionFalse,
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

	serviceCtxs, err := buildServiceContexts(
		logger.WithName("buildServiceContexts"),
		r.dynClient,
		sbr.GetNamespace(),
		sbr.Spec.Services,
		sbr.Spec.DetectBindingResources,
		r.restMapper,
	)
	if err != nil {
		//handle service not found error
		if k8serrors.IsNotFound(err) {
			err = updateSBRConditions(r.dynClient, sbr,
				conditionsv1.Condition{
					Type:    CollectionReady,
					Status:  corev1.ConditionFalse,
					Reason:  ServiceNotFoundReason,
					Message: err.Error(),
				},
				conditionsv1.Condition{
					Type:   InjectionReady,
					Status: corev1.ConditionFalse,
				},
				conditionsv1.Condition{
					Type:   BindingReady,
					Status: corev1.ConditionFalse,
				},
			)
			if err != nil {
				logger.Error(err, "Failed to update SBR conditions", "sbr", sbr)
			}
		}
		return requeueError(err)
	}
	binding, err := buildBinding(
		r.dynClient,
		sbr.Spec.CustomEnvVar,
		serviceCtxs,
		sbr.Spec.EnvVarPrefix,
	)
	if err != nil {
		return requeueError(err)
	}

	options := &serviceBinderOptions{
		dynClient:              r.dynClient,
		detectBindingResources: *sbr.Spec.DetectBindingResources,
		sbr:                    sbr,
		logger:                 logger,
		objects:                serviceCtxs.getServices(),
		binding:                binding,
		restMapper:             r.restMapper,
	}

	sb, err := buildServiceBinder(ctx, options)
	if err != nil {
		// BuildServiceBinder can return only InvalidOptionsErr, and it is a programmer's error so
		// just bail out without re-queueing nor updating conditions.
		logger.Error(err, "Building ServiceBinder")
		return noRequeue(err)
	}

	if sbr.Spec.Application != nil {
		gvrSpec := sbr.Spec.Application.GroupVersionResource
		gvr := schema.GroupVersionResource{
			Group:    gvrSpec.Group,
			Version:  gvrSpec.Version,
			Resource: gvrSpec.Resource,
		}

		err = r.resourceWatcher.AddWatchForGVR(gvr)
		if err != nil {
			logger.Error(err, "Error add watching application GVR")
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

func updateSBRConditions(dynClient dynamic.Interface, sbr *v1alpha1.ServiceBinding, conditions ...conditionsv1.Condition) error {
	for _, v := range conditions {
		conditionsv1.SetStatusCondition(&sbr.Status.Conditions, v)
	}
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return err
	}

	nsClient := dynClient.
		Resource(groupVersion).
		Namespace(sbr.GetNamespace())

	_, err = nsClient.UpdateStatus(u, metav1.UpdateOptions{})

	return err
}
