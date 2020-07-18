package servicebindingrequest

import (
	"context"
	"errors"
	"sort"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	// CollectionReady indicates readiness for collection and persistance of intermediate manifests
	CollectionReady conditionsv1.ConditionType = "CollectionReady"
	// InjectionReady indicates readiness to change application manifests to use those intermediate manifests
	// If status is true, it indicates that the binding succeeded
	InjectionReady conditionsv1.ConditionType = "InjectionReady"
	// EmptyServiceSelectorsReason is used when the ServiceBindingRequest has empty
	// backingServiceSelectors.
	EmptyServiceSelectorsReason = "EmptyServiceSelectors"
	// EmptyApplicationSelectorReason is used when the ServiceBindingRequest has empty
	// applicationSelector.
	EmptyApplicationSelectorReason = "EmptyApplicationSelector"
	// ApplicationNotFoundReason is used when the application is not found.
	ApplicationNotFoundReason = "ApplicationNotFound"
)

// Reconciler reconciles a ServiceBindingRequest object
type reconciler struct {
	dynClient       dynamic.Interface // kubernetes dynamic api client
	scheme          *runtime.Scheme   // api scheme
	restMapper      meta.RESTMapper   // restMapper to convert GVK and GVR
	resourceWatcher ResourceWatcher   // ResourceWatcher to add watching for specific GVK/GVR
}

// reconcilerLog local logger instance
var reconcilerLog = log.NewLog("reconciler")

//// validateServiceBindingRequest check for unsupported settings in SBR.
//func (r *Reconciler) validateServiceBindingRequest(sbr *v1alpha1.ServiceBindingRequest) error {
//	// check if application ResourceRef and MatchLabels, one of them is required.
//	if sbr.Spec.ApplicationSelector.ResourceRef == "" &&
//		sbr.Spec.ApplicationSelector.LabelSelector == nil {
//		return fmt.Errorf("both ResourceRef and LabelSelector are not set")
//	}
//	return nil
//}

// getServiceBindingRequest retrieve the SBR object based on namespaced-name.
func (r *reconciler) getServiceBindingRequest(
	namespacedName types.NamespacedName,
) (*v1alpha1.ServiceBindingRequest, error) {
	gr := v1alpha1.SchemeGroupVersion.WithResource(serviceBindingRequestResource)
	resourceClient := r.dynClient.Resource(gr).Namespace(namespacedName.Namespace)
	u, err := resourceClient.Get(namespacedName.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	sbr := &v1alpha1.ServiceBindingRequest{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}

	return sbr, nil
}

// distinctServiceSelectors returns the distinct elements from the given service selector slice.
func distinctServiceSelectors(selectors []v1alpha1.BackingServiceSelector) []v1alpha1.BackingServiceSelector {
	distinct := make(map[v1alpha1.BackingServiceSelector]bool)
	for _, v := range selectors {
		distinct[v] = true
	}

	var output []v1alpha1.BackingServiceSelector
	for k := range distinct {
		output = append(output, k)
	}

	return output
}

// extractServiceSelectors returns a list of all BackingServiceSelector items from a
// ServiceBindingRequest.
//
// NOTE(isuttonl): remove this method when spec.backingServiceSelector is deprecated
func extractServiceSelectors(
	sbr *v1alpha1.ServiceBindingRequest,
) []v1alpha1.BackingServiceSelector {
	selector := sbr.Spec.BackingServiceSelector
	inSelectors := sbr.Spec.BackingServiceSelectors
	var selectors []v1alpha1.BackingServiceSelector

	if selector != nil {
		selectors = append(selectors, *selector)
	}
	if inSelectors != nil {
		selectors = append(selectors, *inSelectors...)
	}

	// FIXME(isuttonl): sorting selectors using name and namespace can be more robust.
	sort.Slice(selectors, func(i, j int) bool {
		a := selectors[i].ResourceRef < selectors[j].ResourceRef
		b := selectors[j].Namespace != nil &&
			selectors[i].Namespace != nil &&
			*selectors[i].Namespace < *selectors[j].Namespace
		return a && b
	})

	return distinctServiceSelectors(selectors)
}

// Reconcile a ServiceBindingRequest by the following steps:
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

	logger.Info("Reconciling ServiceBindingRequest...")

	// fetch and validate namespaced ServiceBindingRequest instance
	sbr, err := r.getServiceBindingRequest(request.NamespacedName)
	if err != nil {
		if errors.Is(err, errApplicationNotFound) {
			logger.Info("SBR deleted after application deletion")
			return done()
		}
		logger.Error(err, "On retrieving service-binding-request instance.")
		return doneOnNotFound(err)
	}

	// validate namespaced ServiceBindingRequest instance (this check has been disabled until test data has been
	// adjusted to reflect the validation)
	//
	//if err = r.validateServiceBindingRequest(sbr); err != nil {
	//	logger.Error(err, "On validating service-binding-request instance.")
	//	return Done()
	//}

	logger = logger.WithValues("ServiceBindingRequest.Name", sbr.Name)
	logger.Debug("Found service binding request to inspect")

	ctx := context.Background()

	selectors := extractServiceSelectors(sbr)
	if len(selectors) == 0 {
		_, updateErr := updateServiceBindingRequestStatus(
			r.dynClient,
			sbr,
			conditionsv1.Condition{
				Type:    CollectionReady,
				Status:  corev1.ConditionFalse,
				Reason:  EmptyServiceSelectorsReason,
				Message: errEmptyBackingServiceSelectors.Error(),
			},
			conditionsv1.Condition{
				Type:    InjectionReady,
				Status:  corev1.ConditionFalse,
				Reason:  EmptyServiceSelectorsReason,
				Message: errEmptyBackingServiceSelectors.Error(),
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
		return requeueError(errEmptyBackingServiceSelectors)
	}

	serviceCtxs, err := buildServiceContexts(
		logger.WithName("buildServiceContexts"),
		r.dynClient,
		sbr.GetNamespace(),
		selectors,
		sbr.Spec.DetectBindingResources,
		r.restMapper,
	)
	if err != nil {
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
		detectBindingResources: sbr.Spec.DetectBindingResources,
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

	gvrSpec := sbr.Spec.ApplicationSelector.GroupVersionResource
	gvr := schema.GroupVersionResource{
		Group:    gvrSpec.Group,
		Version:  gvrSpec.Version,
		Resource: gvrSpec.Resource,
	}

	err = r.resourceWatcher.AddWatchForGVR(gvr)
	if err != nil {
		logger.Error(err, "Error add watching application GVR")
	}

	if sbr.GetDeletionTimestamp() != nil && sbr.GetOwnerReferences() != nil {
		logger := logger.WithName("Deleting SBR when it has ownerReference")
		logger.Debug("Removing resource finalizers...")
		removeFinalizer(sbr)
		if _, err := updateServiceBindingRequest(r.dynClient, sbr); err != nil {
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
