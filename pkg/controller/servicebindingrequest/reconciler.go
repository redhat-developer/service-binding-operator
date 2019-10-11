package servicebindingrequest

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	log "github.com/redhat-developer/service-binding-operator/pkg/log"
)

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	client    client.Client     // kubernetes api client
	dynClient dynamic.Interface // kubernetes dynamic api client
	scheme    *runtime.Scheme   // api scheme
}

const (
	// binding is in progress
	bindingInProgress = "InProgress"
	// binding has succeeded
	bindingSuccess = "Success"
	// binding has failed
	bindingFail = "Fail"
	// time in seconds to wait before requeuing requests
	requeueAfter int64 = 45
)

var (
	reconcilerLog = log.NewLog("reconciler")
)

// setSecretName update the CR status field to "in progress", and setting secret name.
func (r *Reconciler) setSecretName(sbrStatus *v1alpha1.ServiceBindingRequestStatus, name string) {
	sbrStatus.BindingStatus = bindingInProgress
	sbrStatus.Secret = name
}

// setStatus update the CR status field.
func (r *Reconciler) setStatus(sbrStatus *v1alpha1.ServiceBindingRequestStatus, status string) {
	sbrStatus.BindingStatus = status
}

// setApplicationObjects set the ApplicationObject status field, and also set the overall status as
// success, since it was able to bind applications.
func (r *Reconciler) setApplicationObjects(
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) {
	names := []string{}
	for _, obj := range objs {
		names = append(names, fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))
	}
	sbrStatus.BindingStatus = bindingSuccess
	sbrStatus.ApplicationObjects = names
}

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

// updateStatusServiceBindingRequest update the status field of a ServiceBindingRequest.
func (r *Reconciler) updateStatusServiceBindingRequest(
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
) error {
	// coping status over informed object
	sbr.Status = *sbrStatus

	// converting object into unstructured
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sbr)
	if err != nil {
		return err
	}
	u := &unstructured.Unstructured{Object: data}

	gr := v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)
	resourceClient := r.dynClient.Resource(gr).Namespace(sbr.GetNamespace())
	_, err = resourceClient.UpdateStatus(u, metav1.UpdateOptions{})
	return err
}

// onError comprise the update of ServiceBindingRequest status to set error flag, and inspect
// informed error to apply a different behavior for not-founds.
func (r *Reconciler) onError(
	err error,
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objs []*unstructured.Unstructured,
) (reconcile.Result, error) {
	// settting overall status to failed
	r.setStatus(sbrStatus, bindingFail)
	//
	if objs != nil {
		r.setApplicationObjects(sbrStatus, objs)
	}
	errStatus := r.updateStatusServiceBindingRequest(sbr, sbrStatus)
	if errStatus != nil {
		return RequeueError(errStatus)
	}
	return RequeueOnNotFound(err, requeueAfter)
}

// checkSBR checks the Service Binding Request
func checkSBR(sbr *v1alpha1.ServiceBindingRequest, log *log.Log) error {
	// Check if application ResourceRef is present
	if sbr.Spec.ApplicationSelector.ResourceRef == "" {
		log.Debug("Spec.ApplicationSelector.ResourceRef not found")

		// Check if MatchLabels is present
		if sbr.Spec.ApplicationSelector.MatchLabels == nil {

			err := errors.New("NotFoundError")
			log.Error(err, "Spec.ApplicationSelector.MatchLabels not found")
			return err
		}
	}
	return nil
}

// Reconcile a ServiceBindingRequest by the following steps:
// 1. Inspecting SBR in order to identify backend service. The service is composed by a CRD name and
//    kind, and by inspecting "connects-to" label identify the name of service instance;
// 2. Using OperatorLifecycleManager standards, identifying which items are intersting for binding
//    by parsing CustomResourceDefinitionDescripton object;
// 3. Search and read contents identified in previous step, creating an intermediary secret to hold
//    data formatted as environment variables key/value;
// 4. Search applications that are interested to bind with given service, by inspecting labels. The
//    Deployment (and other kinds) will be updated in "spec" level.
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	objectsToAnnotate := []*unstructured.Unstructured{}

	log := reconcilerLog.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name,
	)
	log.Info("Reconciling ServiceBindingRequest...")

	// fetch the ServiceBindingRequest instance
	sbr, err := r.getServiceBindingRequest(request.NamespacedName)
	if err != nil {
		log.Error(err, "On retrieving service-binding-request instance.")
		return RequeueError(err)
	}

	log = log.WithValues("ServiceBindingRequest.Name", sbr.Name)
	log.Debug("Found service binding request to inspect")

	// splitting instance from it's status
	sbrStatus := sbr.Status

	// Check Service Binding Request
	err = checkSBR(sbr, log)
	if err != nil {
		log.Error(err, "")
		return RequeueError(err)
	}

	//
	// Planing changes
	//

	log.Debug("Creating a plan based on OLM and CRD.")
	planner := NewPlanner(ctx, r.dynClient, sbr)
	plan, err := planner.Plan()
	if err != nil {
		log.Error(err, "On creating a plan to bind applications.")
		return r.onError(err, sbr, &sbrStatus, nil)
	}

	// storing CR in objects to annotate
	objectsToAnnotate = append(objectsToAnnotate, plan.CR)

	//
	// Retrieving data
	//

	log.Debug("Retrieving data to create intermediate secret.")
	retriever := NewRetriever(r.dynClient, plan, sbr.Spec.EnvVarPrefix)
	retrievedObjects, err := retriever.Retrieve()
	if err != nil {
		log.Error(err, "On retrieving binding data.")
		return r.onError(err, sbr, &sbrStatus, nil)
	}

	r.setSecretName(&sbrStatus, plan.Name)

	// storing objects used in Retriever
	objectsToAnnotate = append(objectsToAnnotate, retrievedObjects...)

	//
	// Updating applications to use intermediary secret
	//

	log.Info("Binding applications with intermediary secret.")
	binder := NewBinder(ctx, r.client, r.dynClient, sbr, retriever.volumeKeys)
	updatedObjects, err := binder.Bind()
	if err != nil {
		log.Error(err, "On binding application.")
		return r.onError(err, sbr, &sbrStatus, updatedObjects)
	}

	// saving on status the list of objects that have been touched
	r.setApplicationObjects(&sbrStatus, updatedObjects)
	// storing objects used in Binder
	objectsToAnnotate = append(objectsToAnnotate, updatedObjects...)

	//
	// Annotating objects related to binding
	//

	if err = SetSBRAnnotations(r.dynClient, request.NamespacedName, objectsToAnnotate); err != nil {
		log.Error(err, "On setting annotations in related objects.")
		return r.onError(err, sbr, &sbrStatus, updatedObjects)
	}

	// updating status of request instance
	if err = r.updateStatusServiceBindingRequest(sbr, &sbrStatus); err != nil {
		log.Error(err, "On updating status of ServiceBindingRequest.")
		return RequeueError(err)
	}

	log.Info("All done!")
	return Done()
}
