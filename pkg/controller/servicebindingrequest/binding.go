package servicebindingrequest

import (
	"context"
	"errors"
	"fmt"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"gotest.tools/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/conditions"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	// BindingSuccess binding has succeeded
	BindingSuccess = "BindingSuccess"
	// BindingFail binding has failed
	BindingFail = "BindingFail"
	//Finalizer annotation used in finalizer steps
	Finalizer = "finalizer.servicebindingrequest.openshift.io"
	// time in seconds to wait before requeuing requests
	requeueAfter int64 = 45
)

// GroupVersion represents the service binding request resource's group version.
var GroupVersion = v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)

// message converts the error to string for the Message field in the Status condition
func (b *ServiceBinder) message(err error) string {
	return err.Error()
}

// ServiceBinderOptions is BuildServiceBinder arguments.
type ServiceBinderOptions struct {
	Logger                 *log.Log
	DynClient              dynamic.Interface
	DetectBindingResources bool
	EnvVarPrefix           string
	SBR                    *v1alpha1.ServiceBindingRequest
	Client                 client.Client
}

// Valid returns whether the options are valid.
func (o *ServiceBinderOptions) Valid() bool {
	return o.SBR != nil && o.DynClient != nil && o.Client != nil
}

// ServiceBinder manages binding for a Service Binding Request and associated objects.
type ServiceBinder struct {
	// Binder is responsible for interacting with the cluster and apply binding related changes.
	Binder *Binder
	// Data is the collection of all data read by the manager.
	Data map[string][]byte
	// DynClient is the Kubernetes dynamic client used to interact with the cluster.
	DynClient dynamic.Interface
	// Logger provides logging facilities for internal components.
	Logger *log.Log
	// Objects is a list of additional unstructured objects related to the Service Binding Request.
	Objects []*unstructured.Unstructured
	// SBR is the ServiceBindingRequest associated with binding.
	SBR *v1alpha1.ServiceBindingRequest
	// Secret is the Secret associated with the Service Binding Request.
	bindingDataHandler *BindingDataHandler
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
		Resource(GroupVersion).
		Namespace(sbr.GetNamespace())

	u, err = nsClient.Update(u, metav1.UpdateOptions{})

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
func (b *ServiceBinder) updateServiceBindingRequest(
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	return updateServiceBindingRequest(b.DynClient, sbr)
}

// Unbind removes the relationship between a Service Binding Request and its related objects.
func (b *ServiceBinder) Unbind() (reconcile.Result, error) {
	logger := b.Logger.WithName("Unbind")

	logger.Info("Cleaning related objects from operator's annotations...")
	if err := RemoveSBRAnnotations(b.DynClient, b.Objects); err != nil {
		logger.Error(err, "On removing annotations from related objects.")
		return RequeueError(err)
	}

	if err := b.Binder.Unbind(); err != nil {
		logger.Error(err, "On unbinding related objects")
		return RequeueError(err)
	}

	logger.Info("Deleting intermediary secret")
	if err := b.bindingDataHandler.Delete(); err != nil {
		logger.Error(err, "On deleting intermediary secret.")
		return RequeueError(err)
	}

	logger.Debug("Removing resource finalizers...")
	b.SBR.SetFinalizers(removeStringSlice(b.SBR.GetFinalizers(), Finalizer))
	if _, err := b.updateServiceBindingRequest(b.SBR); err != nil {
		return NoRequeue(err)
	}

	return Done()
}

// UpdateServiceBindingRequestStatus execute update API call on a SBR Status. It can return errors from
// this action.
func updateServiceBindingRequestStatus(
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return nil, err
	}

	nsClient := dynClient.
		Resource(GroupVersion).
		Namespace(sbr.GetNamespace())

	u, err = nsClient.UpdateStatus(u, metav1.UpdateOptions{})

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
func (b *ServiceBinder) updateStatusServiceBindingRequest(
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

	return updateServiceBindingRequestStatus(b.DynClient, sbr)
}

// onError comprise the update of ServiceBindingRequest status to set error flag, and inspect
// informed error to apply a different behavior for not-founds.
func (b *ServiceBinder) onError(
	err error,
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) (reconcile.Result, error) {

	if objs != nil {
		b.setApplicationObjects(sbrStatus, objs)
	}
	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:    conditions.BindingReady,
		Status:  corev1.ConditionFalse,
		Reason:  BindingFail,
		Message: b.message(err),
	})
	sbrStatus.BindingStatus = BindingFail
	newSbr, errStatus := b.updateStatusServiceBindingRequest(sbr, sbrStatus)
	if errStatus != nil {
		return RequeueError(errStatus)
	}
	b.SBR = newSbr

	return RequeueOnNotFound(err, requeueAfter)
}

// Bind configures binding between the Service Binding Request and its related objects.
func (b *ServiceBinder) Bind() (reconcile.Result, error) {
	sbrStatus := b.SBR.Status.DeepCopy()

	b.Logger.Info("Saving data on intermediary secret...")
	bindingData, err := b.bindingDataHandler.Commit(b.Data)
	if err != nil {
		b.Logger.Error(err, "On saving secret data..")
		return b.onError(err, b.SBR, sbrStatus, nil)
	}
	bindingDataAPIGroup := fmt.Sprintf("%s/%s", bindingData.GroupVersionKind().Group, bindingData.GroupVersionKind().Version)
	sbrStatus.BindingData = v1alpha1.BindingData{
		TypedLocalObjectReference: v1.TypedLocalObjectReference{
			APIGroup: &bindingDataAPIGroup,
			Kind:     bindingData.GetKind(),
			Name:     bindingData.GetName(),
		},
	}

	updatedObjects, err := b.Binder.Bind()
	if err != nil {
		b.Logger.Error(err, "On binding application.")
		return b.onError(err, b.SBR, sbrStatus, updatedObjects)
	}
	b.setApplicationObjects(sbrStatus, updatedObjects)

	// annotating objects related to binding
	namespacedName := types.NamespacedName{Namespace: b.SBR.GetNamespace(), Name: b.SBR.GetName()}
	if err = SetSBRAnnotations(b.DynClient, namespacedName, append(b.Objects, bindingData)); err != nil {
		b.Logger.Error(err, "On setting annotations in related objects.")
		return b.onError(err, b.SBR, sbrStatus, updatedObjects)
	}

	sbrStatus.BindingStatus = BindingSuccess
	conditionsv1.SetStatusCondition(&sbrStatus.Conditions, conditionsv1.Condition{
		Type:   conditions.BindingReady,
		Status: corev1.ConditionTrue,
	})

	// updating status of request instance
	sbr, err := b.updateStatusServiceBindingRequest(b.SBR, sbrStatus)
	if err != nil {
		return RequeueOnConflict(err)
	}

	// appending finalizer, should be later removed upon resource deletion
	sbr.SetFinalizers(append(removeStringSlice(b.SBR.GetFinalizers(), Finalizer), Finalizer))
	if _, err = b.updateServiceBindingRequest(sbr); err != nil {
		return NoRequeue(err)
	}

	b.Logger.Info("All done!")
	return Done()
}

// setApplicationObjects replaces the Status's equivalent field.
func (b *ServiceBinder) setApplicationObjects(
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) {
	boundApps := []v1alpha1.BoundApplication{}
	for _, obj := range objs {
		boundApp := v1alpha1.BoundApplication{
			GroupVersionKind: metav1.GroupVersionKind{
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

// buildPlan creates a new plan.
func buildPlan(
	ctx context.Context,
	dynClient dynamic.Interface,
	sbr *v1alpha1.ServiceBindingRequest,
) (*Plan, error) {
	planner := NewPlanner(ctx, dynClient, sbr)
	return planner.Plan()
}

// InvalidOptionsErr is returned when ServiceBinderOptions are not valid.
var InvalidOptionsErr = errors.New("invalid options")

// BuildServiceBinder creates a new binding manager according to options.
func BuildServiceBinder(options *ServiceBinderOptions) (*ServiceBinder, error) {
	if !options.Valid() {
		return nil, InvalidOptionsErr
	}

	// objs groups all extra objects related to the informed SBR
	objs := make([]*unstructured.Unstructured, 0)

	// plan is a source of information regarding the binding process
	ctx := context.Background()
	plan, err := buildPlan(ctx, options.DynClient, options.SBR)
	if err != nil {
		return nil, err
	}

	rs := plan.GetCRs()
	// append all SBR related CRs
	objs = append(objs, rs...)

	// retriever is responsible for gathering data related to the given plan.
	retriever := NewRetriever(options.DynClient, plan, options.EnvVarPrefix)

	// read bindable data from the specified resources
	if options.DetectBindingResources {
		err := retriever.ReadBindableResourcesData(&plan.SBR, rs)
		if err != nil {
			return nil, err
		}
	}

	// read bindable data from the CRDDescription found by the planner
	for _, r := range plan.GetRelatedResources() {
		err = retriever.ReadCRDDescriptionData(r.CR, r.CRDDescription)
		if err != nil {
			return nil, err
		}
	}

	// gather retriever's read data
	// TODO: do not return error
	retrievedData, err := retriever.Get()
	if err != nil {
		return nil, err
	}

	// gather related secret, again only appending it if there's a value.
	bindindDataHandler := NewBindingDataHandler(options.DynClient, plan)

	return &ServiceBinder{
		Logger:             options.Logger,
		Binder:             NewBinder(ctx, options.Client, options.DynClient, options.SBR, retriever.VolumeKeys),
		DynClient:          options.DynClient,
		SBR:                options.SBR,
		Objects:            objs,
		Data:               retrievedData,
		bindingDataHandler: bindindDataHandler,
	}, nil
}
