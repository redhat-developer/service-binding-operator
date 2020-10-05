package servicebinding

import (
	"context"
	"errors"
	"fmt"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"gotest.tools/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	// bindingFail binding has failed
	bindingFail = "BindingFail"
	//finalizer annotation used in finalizer steps
	finalizer = "finalizer.servicebinding.openshift.io"
	// time in seconds to wait before requeuing requests
	requeueAfter int64 = 45
)

// defaultPathToContainers has the logical path logical path
// to find containers on supported objects
// Used as []string{"spec", "template", "spec", "containers"}
const defaultPathToContainers = "spec.template.spec.containers"

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
	sbr                    *v1alpha1.ServiceBinding
	objects                []*unstructured.Unstructured
	binding                *internalBinding
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

// serviceBinder manages binding for a Service Binding and associated objects.
type serviceBinder struct {
	// binder is responsible for interacting with the cluster and apply binding related changes.
	binder *binder
	// envVars contains the environment variables to bind.
	envVars map[string][]byte
	// dynClient is the Kubernetes dynamic client used to interact with the cluster.
	dynClient dynamic.Interface
	// logger provides logging facilities for internal components.
	logger *log.Log
	// objects is a list of additional unstructured objects related to the Service Binding.
	objects []*unstructured.Unstructured
	// sbr is the ServiceBinding associated with binding.
	sbr *v1alpha1.ServiceBinding
	// secret is the secret associated with the Service Binding.
	secret *secret
}

// updateServiceBinding execute update API call on a SBR request. It can return errors from
// this action.
func updateServiceBinding(
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBinding,
) (*v1alpha1.ServiceBinding, error) {
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

// updateServiceBinding execute update API call on a SBR request. It can return errors from
// this action.
func (b *serviceBinder) updateServiceBinding(
	sbr *v1alpha1.ServiceBinding,
) (*v1alpha1.ServiceBinding, error) {
	return updateServiceBinding(b.dynClient, sbr)
}

// unbind removes the relationship between a Service Binding and its related objects.
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
	if _, err := b.updateServiceBinding(b.sbr); err != nil {
		b.logger.Error(err, "Updating ServiceBinding")
		return noRequeue(err)
	}

	return done()
}

// UpdateServiceBindingStatus execute update API call on a SBR Status. It can return errors from
// this action.
func updateServiceBindingStatus(
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBinding,
	conditions ...conditionsv1.Condition,
) (*v1alpha1.ServiceBinding, error) {
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

// updateStatusServiceBinding updates the Service Binding's status field.
func (b *serviceBinder) updateStatusServiceBinding(
	sbr *v1alpha1.ServiceBinding,
	sbrStatus *v1alpha1.ServiceBindingStatus,
) (
	*v1alpha1.ServiceBinding,
	error,
) {
	// do not update if both statuses are equal
	if result := cmp.DeepEqual(sbr.Status, sbrStatus)(); result.Success() {
		return sbr, nil
	}

	// coping status over informed object
	sbr.Status = *sbrStatus

	return updateServiceBindingStatus(b.dynClient, sbr)
}

// onError comprise the update of ServiceBinding status to set error flag, and inspect
// informed error to apply a different behavior for not-founds.
func (b *serviceBinder) onError(
	err error,
	sbr *v1alpha1.ServiceBinding,
	sbrStatus *v1alpha1.ServiceBindingStatus,
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
	newSbr, errStatus := b.updateStatusServiceBinding(sbr, sbrStatus)
	if errStatus != nil {
		return requeueError(errStatus)
	}
	b.sbr = newSbr

	return requeueOnNotFound(err, requeueAfter)
}

// isApplicationEmpty returns true if application is not declared in
// the Service Binding.
func isApplicationEmpty(
	application *v1alpha1.Application,
) bool {
	fmt.Println("Current APPLICATION IS : ", application)
	if application == nil {
		return true
	}
	emptyApplication := &v1alpha1.Application{
		LabelSelector: &v1.LabelSelector{},
	}
	return application == emptyApplication ||
		application.Name == "" && application.LabelSelector != nil && application.LabelSelector.MatchLabels == nil
}

func addFinalizer(sbr *v1alpha1.ServiceBinding) {
	sbr.SetFinalizers(append(removeStringSlice(sbr.GetFinalizers(), finalizer), finalizer))
}

func removeFinalizer(sbr *v1alpha1.ServiceBinding) {
	sbr.SetFinalizers(removeStringSlice(sbr.GetFinalizers(), finalizer))
}

// handleApplicationError handles scenarios when:
// 1. application not declared in the Service Binding
// 2. application not found
func (b *serviceBinder) handleApplicationError(reason string, applicationError error, sbrStatus *v1alpha1.ServiceBindingStatus) (reconcile.Result, error) {
	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:    InjectionReady,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: applicationError.Error(),
	})

	// updating status of request instance
	sbr, err := b.updateStatusServiceBinding(b.sbr, sbrStatus)
	if err != nil {
		return requeueError(err)
	}

	b.logger.Info(applicationError.Error())

	if errors.Is(applicationError, errApplicationNotFound) {
		removeFinalizer(b.sbr)
		if _, err = b.updateServiceBinding(sbr); err != nil {
			b.logger.Error(err, "Updating ServiceBinding")
			return requeueError(err)
		}
	}

	return done()
}

// bind configures binding between the Service Binding and its related objects.
func (b *serviceBinder) bind() (reconcile.Result, error) {
	sbrStatus := b.sbr.Status.DeepCopy()

	b.logger.Debug("Saving data on intermediary secret...")

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

	if isApplicationEmpty(b.sbr.Spec.Application) {
		return b.handleApplicationError(EmptyApplicationReason, errEmptyApplication, sbrStatus)
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

	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:   InjectionReady,
		Status: corev1.ConditionTrue,
	})

	// updating status of request instance
	sbr, err := b.updateStatusServiceBinding(b.sbr, sbrStatus)
	if err != nil {
		return requeueOnConflict(err)
	}

	// appending finalizer, should be later removed upon resource deletion
	addFinalizer(sbr)

	if _, err = b.updateServiceBinding(sbr); err != nil {
		b.logger.Error(err, "Updating ServiceBinding")
		return requeueError(err)
	}

	b.logger.Info("All done!")
	return done()
}

// setApplicationObjects replaces the Status's equivalent field.
func (b *serviceBinder) setApplicationObjects(
	sbrStatus *v1alpha1.ServiceBindingStatus,
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

// set default value for application selector
func ensureDefaults(applicationSelector *v1alpha1.Application) {
	if applicationSelector != nil {
		if applicationSelector.LabelSelector == nil {
			applicationSelector.LabelSelector = &metav1.LabelSelector{}
		}
		if applicationSelector.BindingPath == nil {
			applicationSelector.BindingPath = &v1alpha1.BindingPath{
				ContainersPath: defaultPathToContainers,
			}
		}
	} else {
		applicationSelector = &v1alpha1.Application{}
		applicationSelector.LabelSelector = &metav1.LabelSelector{}
		applicationSelector.BindingPath = &v1alpha1.BindingPath{
			ContainersPath: defaultPathToContainers,
		}
	}
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

	ensureDefaults(options.sbr.Spec.Application)

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

type internalBinding struct {
	envVars    map[string][]byte
	volumeKeys []string
}

func buildBinding(
	client dynamic.Interface,
	customEnvVar []corev1.EnvVar,
	svcCtxs serviceContextList,
	globalEnvVarPrefix string,
) (*internalBinding, error) {
	envVars, volumeKeys, err := NewRetriever(client).
		ProcessServiceContexts(globalEnvVarPrefix, svcCtxs, customEnvVar)
	if err != nil {
		return nil, err
	}

	return &internalBinding{
		envVars:    envVars,
		volumeKeys: volumeKeys,
	}, nil
}
