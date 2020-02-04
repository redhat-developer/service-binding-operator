package servicebindingrequest

import (
	"context"
	"errors"
	"fmt"

	"gotest.tools/assert/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

const (
	// bindingInProgress binding is in progress
	bindingInProgress = "InProgress"
	// bindingFail binding has failed
	bindingFail = "Fail"
	// time in seconds to wait before requeuing requests
	requeueAfter int64 = 45
	// Finalizer annotation used in finalizer steps
	Finalizer = "finalizer.servicebindingrequest.openshift.io"
)

// GroupVersion represents the service binding request resource's group version.
var GroupVersion = v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)

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
	// SBR is the ServiceBindingRequest associated with binding..
	SBR *v1alpha1.ServiceBindingRequest
	// Secret is the Secret associated with the Service Binding Request.
	Secret *Secret
}

// updateServiceBindingRequest execute update API call on a SBR request. It can return errors from
// this action.
func (b *ServiceBinder) updateServiceBindingRequest(
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return nil, err
	}

	u, err = b.DynClient.
		Resource(GroupVersion).
		Namespace(sbr.GetNamespace()).
		Update(u, v1.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}
	return sbr, nil
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
	if err := b.Secret.Delete(); err != nil {
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

	// converting object into unstructured
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return nil, err
	}

	gr := v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)
	resourceClient := b.DynClient.Resource(gr).Namespace(sbr.GetNamespace())
	u, err = resourceClient.UpdateStatus(u, v1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sbr)
	if err != nil {
		return nil, err
	}
	return sbr, nil
}

// onError comprise the update of ServiceBindingRequest status to set error flag, and inspect
// informed error to apply a different behavior for not-founds.
func (b *ServiceBinder) onError(
	err error,
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) (reconcile.Result, error) {
	sbrStatus.BindingStatus = bindingFail

	if objs != nil {
		sbrStatus.BindingStatus = BindingSuccess
		b.setApplicationObjects(sbrStatus, objs)
	}

	_, errStatus := b.updateStatusServiceBindingRequest(sbr, sbrStatus)
	if errStatus != nil {
		return RequeueError(errStatus)
	}

	return RequeueOnNotFound(err, requeueAfter)
}

// Bind configures binding between the Service Binding Request and its related objects.
func (b *ServiceBinder) Bind() (reconcile.Result, error) {
	sbrStatus := b.SBR.Status.DeepCopy()

	b.Logger.Info("Saving data on intermediary secret...")
	secretObj, err := b.Secret.Commit(b.Data, nil)
	if err != nil {
		b.Logger.Error(err, "On saving secret data..")
		return b.onError(err, b.SBR, sbrStatus, nil)
	}

	// update status information
	sbrStatus.BindingStatus = bindingInProgress
	sbrStatus.Secret = secretObj.GetName()

	updatedObjects, err := b.Binder.Bind()
	if err != nil {
		b.Logger.Error(err, "On binding application.")
		return b.onError(err, b.SBR, sbrStatus, updatedObjects)
	}

	// saving on status the list of objects that have been touched
	sbrStatus.BindingStatus = BindingSuccess
	b.setApplicationObjects(sbrStatus, updatedObjects)

	// annotating objects related to binding
	namespacedName := types.NamespacedName{Namespace: b.SBR.GetNamespace(), Name: b.SBR.GetName()}
	if err = SetSBRAnnotations(b.DynClient, namespacedName, b.Objects); err != nil {
		b.Logger.Error(err, "On setting annotations in related objects.")
		return b.onError(err, b.SBR, sbrStatus, updatedObjects)
	}

	// updating status of request instance
	sbr, err := b.updateStatusServiceBindingRequest(b.SBR, sbrStatus)
	if err != nil {
		return RequeueOnConflict(err)
	}

	// appending finalizer, should be later removed upon resource deletion
	sbr.SetFinalizers(append(sbr.GetFinalizers(), Finalizer))
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
	names := []string{}
	for _, obj := range objs {
		names = append(names, fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))
	}
	sbrStatus.ApplicationObjects = names
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
	secret := NewSecret(options.DynClient, plan)
	secretObj, found, err := secret.Get()
	if err != nil {
		return nil, err
	}
	if found {
		objs = append(objs, secretObj)
	}

	return &ServiceBinder{
		Logger:    options.Logger,
		Binder:    NewBinder(ctx, options.Client, options.DynClient, options.SBR, retriever.volumeKeys),
		DynClient: options.DynClient,
		SBR:       options.SBR,
		Objects:   objs,
		Data:      retrievedData,
		Secret:    secret,
	}, nil
}
