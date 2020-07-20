package servicebindingrequest

import (
	"context"
	"errors"
	"fmt"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"gotest.tools/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	// bindingFail binding has failed
	bindingFail = "BindingFail"
	//finalizer annotation used in finalizer steps
	finalizer = "finalizer.servicebindingrequest.openshift.io"
	// time in seconds to wait before requeuing requests
	requeueAfter int64 = 45
)

// groupVersion represents the service binding request resource's group version.
var groupVersion = v1alpha1.SchemeGroupVersion.WithResource(serviceBindingRequestResource)

// message converts the error to string for the Message field in the Status condition
func (b *serviceBinder) message(err error) string {
	return err.Error()
}

// serviceBinderOptions is BuildServiceBinder arguments.
type serviceBinderOptions struct {
	logger                 *log.Log
	dynClient              dynamic.Interface
	detectBindingResources bool
	sbr                    *v1alpha1.ServiceBindingRequest
	objects                []*unstructured.Unstructured
	binding                *binding
	restMapper             meta.RESTMapper
}

// errInvalidServiceBinderOptions is returned when ServiceBinderOptions contains an invalid value.
type errInvalidServiceBinderOptions string

func (e errInvalidServiceBinderOptions) Error() string {
	return fmt.Sprintf("option %q is missing", string(e))
}

// Valid returns an error if the receiver is invalid, nil otherwise.
func (o *serviceBinderOptions) Valid() error {
	if o.sbr == nil {
		return errInvalidServiceBinderOptions("SBR")
	}

	if o.dynClient == nil {
		return errInvalidServiceBinderOptions("DynClient")
	}

	if o.binding == nil {
		return errInvalidServiceBinderOptions("Binding")
	}

	if o.restMapper == nil {
		return errInvalidServiceBinderOptions("RESTMapper")
	}

	return nil
}

// serviceBinder manages binding for a Service Binding Request and associated objects.
type serviceBinder struct {
	// binder is responsible for interacting with the cluster and apply binding related changes.
	binder *binder
	// envVars contains the environment variables to bind.
	envVars map[string][]byte
	// dynClient is the Kubernetes dynamic client used to interact with the cluster.
	dynClient dynamic.Interface
	// logger provides logging facilities for internal components.
	logger *log.Log
	// objects is a list of additional unstructured objects related to the Service Binding Request.
	objects []*unstructured.Unstructured
	// sbr is the ServiceBindingRequest associated with binding.
	sbr *v1alpha1.ServiceBindingRequest
	// secret is the secret associated with the Service Binding Request.
	secret *secret
}

// updateServiceBindingRequest execute update API call on a SBR request. It can return errors from
// this action.
func updateServiceBindingRequest(
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return nil, err
	}

	nsClient := dynClient.
		Resource(groupVersion).
		Namespace(sbr.GetNamespace())

	u, err = nsClient.Update(u, v1.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}

	return sbr, nil
}

// updateServiceBindingRequest execute update API call on a SBR request. It can return errors from
// this action.
func (b *serviceBinder) updateServiceBindingRequest(
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	return updateServiceBindingRequest(b.dynClient, sbr)
}

// unbind removes the relationship between a Service Binding Request and its related objects.
func (b *serviceBinder) unbind() (reconcile.Result, error) {
	logger := b.logger.WithName("Unbind")

	// when finalizer is not found anymore, it can be safely removed
	if !containsStringSlice(b.sbr.GetFinalizers(), finalizer) {
		logger.Info("Resource can be safely deleted!")
		return done()
	}

	logger.Info("Cleaning related objects from operator's annotations...")
	if err := removeAndUpdateSBRAnnotations(b.dynClient, b.objects); err != nil {
		logger.Error(err, "On removing annotations from related objects.")
		return requeueError(err)
	}

	if err := b.binder.unbind(); err != nil {
		logger.Error(err, "On unbinding related objects")
		return requeueError(err)
	}

	logger.Debug("Removing resource finalizers...")
	removeFinalizer(b.sbr)
	if _, err := b.updateServiceBindingRequest(b.sbr); err != nil {
		return noRequeue(err)
	}

	return done()
}

// UpdateServiceBindingRequestStatus execute update API call on a SBR Status. It can return errors from
// this action.
func updateServiceBindingRequestStatus(
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
	conditions ...conditionsv1.Condition,
) (*v1alpha1.ServiceBindingRequest, error) {
	for _, v := range conditions {
		conditionsv1.SetStatusCondition(&sbr.Status.Conditions, v)
	}
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return nil, err
	}

	nsClient := dynClient.
		Resource(groupVersion).
		Namespace(sbr.GetNamespace())

	u, err = nsClient.UpdateStatus(u, v1.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}

	return sbr, nil
}

// updateStatusServiceBindingRequest updates the Service Binding Request's status field.
func (b *serviceBinder) updateStatusServiceBindingRequest(
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
) (
	*v1alpha1.ServiceBindingRequest,
	error,
) {
	// do not update if both statuses are equal
	if result := cmp.DeepEqual(sbr.Status, sbrStatus)(); result.Success() {
		return sbr, nil
	}

	// coping status over informed object
	sbr.Status = *sbrStatus

	return updateServiceBindingRequestStatus(b.dynClient, sbr)
}

// onError comprise the update of ServiceBindingRequest status to set error flag, and inspect
// informed error to apply a different behavior for not-founds.
func (b *serviceBinder) onError(
	err error,
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) (reconcile.Result, error) {

	if objs != nil {
		b.setApplicationObjects(sbrStatus, objs)
	}
	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:    InjectionReady,
		Status:  corev1.ConditionFalse,
		Reason:  bindingFail,
		Message: b.message(err),
	})
	newSbr, errStatus := b.updateStatusServiceBindingRequest(sbr, sbrStatus)
	if errStatus != nil {
		return requeueError(errStatus)
	}
	b.sbr = newSbr

	return requeueOnNotFound(err, requeueAfter)
}

// isApplicationSelectorEmpty returns true if applicationSelector is not declared in
// the Service Binding Request.
func isApplicationSelectorEmpty(
	application v1alpha1.ApplicationSelector,
) bool {
	var emptyApplication v1alpha1.ApplicationSelector
	return application == emptyApplication ||
		application.ResourceRef == "" &&
			application.LabelSelector.MatchLabels == nil
}

func addFinalizer(sbr *v1alpha1.ServiceBindingRequest) {
	sbr.SetFinalizers(append(removeStringSlice(sbr.GetFinalizers(), finalizer), finalizer))
}

func removeFinalizer(sbr *v1alpha1.ServiceBindingRequest) {
	sbr.SetFinalizers(removeStringSlice(sbr.GetFinalizers(), finalizer))
}

// handleApplicationError handles scenarios when:
// 1. applicationSelector not declared in the Service Binding Request
// 2. application not found
func (b *serviceBinder) handleApplicationError(reason string, applicationError error, sbrStatus *v1alpha1.ServiceBindingRequestStatus) (reconcile.Result, error) {
	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:    InjectionReady,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: applicationError.Error(),
	})

	// updating status of request instance
	sbr, err := b.updateStatusServiceBindingRequest(b.sbr, sbrStatus)
	if err != nil {
		return requeueError(err)
	}

	// appending finalizer, should be later removed upon resource deletion
	addFinalizer(b.sbr)
	if _, err = b.updateServiceBindingRequest(sbr); err != nil {
		return requeueError(err)
	}

	b.logger.Info(applicationError.Error())

	if errors.Is(applicationError, errApplicationNotFound) {
		removeFinalizer(b.sbr)
		if _, err = b.updateServiceBindingRequest(sbr); err != nil {
			return requeueError(err)
		}
		return requeue(applicationError, requeueAfter)

	}

	return done()
}

// bind configures binding between the Service Binding Request and its related objects.
func (b *serviceBinder) bind() (reconcile.Result, error) {
	sbrStatus := b.sbr.Status.DeepCopy()

	b.logger.Info("Saving data on intermediary secret...")

	secretObj, err := b.secret.createOrUpdate(b.envVars, b.sbr.AsOwnerReference())
	if err != nil {
		b.logger.Error(err, "On saving secret data..")
		return b.onError(err, b.sbr, sbrStatus, nil)
	}
	sbrStatus.Secret = secretObj.GetName()

	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:   CollectionReady,
		Status: corev1.ConditionTrue,
	})

	if isApplicationSelectorEmpty(b.sbr.Spec.ApplicationSelector) {
		return b.handleApplicationError(EmptyApplicationSelectorReason, errEmptyApplicationSelector, sbrStatus)
	}
	updatedObjects, err := b.binder.bind()
	if err != nil {
		b.logger.Error(err, "On binding application.")
		if errors.Is(err, errApplicationNotFound) {
			return b.handleApplicationError(ApplicationNotFoundReason, errApplicationNotFound, sbrStatus)
		}
		return b.onError(err, b.sbr, sbrStatus, nil)
	}
	b.setApplicationObjects(sbrStatus, updatedObjects)

	// annotating objects related to binding
	namespacedName := types.NamespacedName{Namespace: b.sbr.GetNamespace(), Name: b.sbr.GetName()}
	if err = setAndUpdateSBRAnnotations(b.dynClient, namespacedName, append(b.objects, secretObj)); err != nil {
		b.logger.Error(err, "On setting annotations in related objects.")
		return b.onError(err, b.sbr, sbrStatus, updatedObjects)
	}

	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:   InjectionReady,
		Status: corev1.ConditionTrue,
	})

	// updating status of request instance
	sbr, err := b.updateStatusServiceBindingRequest(b.sbr, sbrStatus)
	if err != nil {
		return requeueOnConflict(err)
	}

	// appending finalizer, should be later removed upon resource deletion
	addFinalizer(sbr)

	if _, err = b.updateServiceBindingRequest(sbr); err != nil {
		return requeueError(err)
	}

	b.logger.Info("All done!")
	return done()
}

// setApplicationObjects replaces the Status's equivalent field.
func (b *serviceBinder) setApplicationObjects(
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) {
	boundApps := []v1alpha1.BoundApplication{}
	for _, obj := range objs {
		boundApp := v1alpha1.BoundApplication{
			GroupVersionKind: v1.GroupVersionKind{
				Group:   obj.GroupVersionKind().Group,
				Version: obj.GroupVersionKind().Version,
				Kind:    obj.GetKind(),
			},
			LocalObjectReference: corev1.LocalObjectReference{
				Name: obj.GetName(),
			},
		}
		boundApps = append(boundApps, boundApp)
	}
	sbrStatus.Applications = boundApps
}

// buildServiceBinder creates a new binding manager according to options.
func buildServiceBinder(
	ctx context.Context,
	options *serviceBinderOptions,
) (
	*serviceBinder,
	error,
) {
	if err := options.Valid(); err != nil {
		return nil, err
	}

	// FIXME(isuttonl): review whether it is possible to move Secret.Commit() and Secret.Delete() to
	// ServiceBinder.
	secret := newSecret(
		options.dynClient,
		options.sbr.GetNamespace(),
		options.sbr.GetName(),
	)

	// FIXME(isuttonl): review whether binder can be lazily created in Bind() and Unbind(); also
	// consider renaming to ResourceBinder
	binder := newBinder(
		ctx,
		options.dynClient,
		options.sbr,
		options.binding.volumeKeys,
		options.restMapper,
	)

	options.sbr.Spec.ApplicationSelector.SetDefaults()

	return &serviceBinder{
		logger:    options.logger,
		binder:    binder,
		dynClient: options.dynClient,
		sbr:       options.sbr,
		objects:   options.objects,
		envVars:   options.binding.envVars,
		secret:    secret,
	}, nil
}

type binding struct {
	envVars    map[string][]byte
	volumeKeys []string
}

func buildBinding(
	client dynamic.Interface,
	customEnvVar []corev1.EnvVar,
	svcCtxs serviceContextList,
	globalEnvVarPrefix string,
) (*binding, error) {
	envVars, volumeKeys, err := NewRetriever(client).
		ProcessServiceContexts(globalEnvVarPrefix, svcCtxs, customEnvVar)
	if err != nil {
		return nil, err
	}

	return &binding{
		envVars:    envVars,
		volumeKeys: volumeKeys,
	}, nil
}
