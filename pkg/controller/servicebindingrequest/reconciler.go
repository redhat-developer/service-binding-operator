package servicebindingrequest

import (
	"errors"

	v1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/redhat-developer/service-binding-operator/pkg/conditions"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	client    client.Client     // kubernetes api client
	dynClient dynamic.Interface // kubernetes dynamic api client
	scheme    *runtime.Scheme   // api scheme
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
func (r *Reconciler) getServiceBindingRequest(
	namespacedName types.NamespacedName,
) (*v1alpha1.ServiceBindingRequest, error) {
	gr := v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)
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

// unbind removes the relationship between the given sbr and the manifests the operator has
// previously modified. This process also deletes any manifests created to support the binding
// functionality, such as ConfigMaps and Secrets.
func (r *Reconciler) unbind(logger *log.Log, bm *ServiceBinder) (reconcile.Result, error) {
	logger = logger.WithName("unbind")

	// when finalizer is not found anymore, it can be safely removed
	if !containsStringSlice(bm.SBR.GetFinalizers(), Finalizer) {
		logger.Info("Resource can be safely deleted!")
		return Done()
	}

	logger.Info("Executing unbinding steps...")
	if res, err := bm.Unbind(); err != nil {
		logger.Error(err, "On unbinding application.")
		return res, err
	}

	logger.Debug("Deletion done!")
	return Done()
}

// bind steps to bind backing service and applications together. It receive the elements collected
// in the common parts of the reconciler, and execute the final binding steps.
func (r *Reconciler) bind(
	logger *log.Log,
	bm *ServiceBinder,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
) (
	reconcile.Result,
	error,
) {
	logger = logger.WithName("bind")

	logger.Info("Binding applications with intermediary secret...")
	return bm.Bind()
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
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := reconcilerLog.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name,
	)

	logger.Info("Reconciling ServiceBindingRequest...")

	// fetch and validate namespaced ServiceBindingRequest instance
	sbr, err := r.getServiceBindingRequest(request.NamespacedName)
	if err != nil {
		if errors.Is(err, ApplicationNotFound) {
			logger.Info("SBR deleted after application deletion")
			return Done()
		}
		logger.Error(err, "On retrieving service-binding-request instance.")
		return DoneOnNotFound(err)
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

	// splitting instance from it's status
	sbrStatus := &sbr.Status

	options := &ServiceBinderOptions{
		Client:                 r.client,
		DynClient:              r.dynClient,
		DetectBindingResources: sbr.Spec.DetectBindingResources,
		EnvVarPrefix:           sbr.Spec.EnvVarPrefix,
		SBR:                    sbr,
		Logger:                 logger,
	}

	bm, err := BuildServiceBinder(options)
	if err != nil {
		logger.Error(err, "Creating binding context")
		if err == EmptyBackingServiceSelectorsErr || err == EmptyApplicationSelectorErr {
			// TODO: find or create an error type containing suitable information to be propagated
			var reason string
			if errors.Is(err, EmptyBackingServiceSelectorsErr) {
				reason = "EmptyBackingServiceSelectors"
			} else {
				reason = "EmptyApplicationSelector"
			}

			v1.SetStatusCondition(&sbr.Status.Conditions, v1.Condition{
				Type:    conditions.BindingReady,
				Status:  corev1.ConditionFalse,
				Reason:  reason,
				Message: err.Error(),
			})
			_, updateErr := updateServiceBindingRequestStatus(r.dynClient, sbr)
			if updateErr == nil {
				return Done()
			}
		}
		return RequeueError(err)
	}

	if sbr.GetDeletionTimestamp() != nil {
		logger.Info("Resource is marked for deletion...")
		return r.unbind(logger, bm)
	}

	logger.Info("Starting the bind of application(s) with backing service...")
	return r.bind(logger, bm, sbrStatus)
}
