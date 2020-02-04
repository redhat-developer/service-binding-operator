package servicebindingrequest

import (
	"context"
	"errors"
	"fmt"

	"gotest.tools/assert/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	client    client.Client     // kubernetes api client
	dynClient dynamic.Interface // kubernetes dynamic api client
	scheme    *runtime.Scheme   // api scheme
}

const (
	// BindingSuccess binding has succeeded
	BindingSuccess = "Success"
	// sbrFinalizer annotation used in finalizer steps
	sbrFinalizer = "finalizer.servicebindingrequest.openshift.io"
)

// reconcilerLog local logger instance
var reconcilerLog = log.NewLog("reconciler")

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
	sbrStatus.BindingStatus = BindingSuccess
	sbrStatus.ApplicationObjects = names
}

// getServiceBindingRequest retrieve the SBR object based on namespaced-name.
func (r *Reconciler) getServiceBindingRequest(
	namespacedName types.NamespacedName,
) (*v1alpha1.ServiceBindingRequest, error) {
	gvr := v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)
	resourceClient := r.dynClient.Resource(gvr).Namespace(namespacedName.Namespace)
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
) (*v1alpha1.ServiceBindingRequest, error) {
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

	gvr := v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)
	resourceClient := r.dynClient.Resource(gvr).Namespace(sbr.GetNamespace())
	u, err = resourceClient.UpdateStatus(u, metav1.UpdateOptions{})
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
func (r *Reconciler) updateServiceBindingRequest(
	sbr *v1alpha1.ServiceBindingRequest,
) (*v1alpha1.ServiceBindingRequest, error) {
	u, err := converter.ToUnstructured(sbr)
	if err != nil {
		return nil, err
	}
	gvr := v1alpha1.SchemeGroupVersion.WithResource(ServiceBindingRequestResource)
	resourceClient := r.dynClient.Resource(gvr).Namespace(sbr.GetNamespace())
	u, err = resourceClient.Update(u, metav1.UpdateOptions{})
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
	_, errStatus := r.updateStatusServiceBindingRequest(sbr, sbrStatus)
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
		if sbr.Spec.ApplicationSelector.LabelSelector == nil {
			err := errors.New("NotFoundError")
			log.Error(err, "Spec.ApplicationSelector.LabelSelector not found")
			return err
		}
	}
	return nil
}

// unbind removes the relationship between the given sbr and the manifests the operator has
// previously modified. This process also deletes any manifests created to support the binding
// functionality, such as ConfigMaps and Secrets.
func (r *Reconciler) unbind(
	logger *log.Log,
	binder *Binder,
	secret *Secret,
	sbr *v1alpha1.ServiceBindingRequest,
	objectsToAnnotate []*unstructured.Unstructured,
) (reconcile.Result, error) {
	logger = logger.WithName("unbind")

	// when finalizer is not found anymore, it can be safely removed
	if !containsStringSlice(sbr.GetFinalizers(), sbrFinalizer) {
		logger.Info("Resource can be safely deleted!")
		return Done()
	}

	logger.Debug("Reading intermediary secret before deletion.")
	secretObj, _, err := secret.Get()
	if err != nil {
		logger.Error(err, "On reading intermediary secret.")
		return RequeueError(err)
	}

	// adding secret on list of objects, to remove annotations from secret before deletion
	objectsToAnnotate = append(objectsToAnnotate, secretObj)

	logger.Info("Cleaning related objects from operator's annotations...")
	if err = RemoveSBRAnnotations(r.dynClient, objectsToAnnotate); err != nil {
		logger.Error(err, "On removing annotations from related objects.")
		return RequeueError(err)
	}

	logger.Info("Executing unbinding steps...")
	err = binder.Unbind()
	if err != nil {
		logger.Error(err, "On unbinding application.")
		return RequeueError(err)
	}

	logger.Info("Deleting intermediary secret...")
	if err = secret.Delete(); err != nil {
		logger.Error(err, "On deleting intermediary secret.")
		return RequeueError(err)
	}

	logger.Debug("Removing resource finalizers...")
	sbr.SetFinalizers(removeStringSlice(sbr.GetFinalizers(), sbrFinalizer))
	if _, err = r.updateServiceBindingRequest(sbr); err != nil {
		return NoRequeue(err)
	}

	logger.Debug("Deletion done!")
	return Done()
}

// bind steps to bind backing service and applications together. It receive the elements collected
// in the common parts of the reconciler, and execute the final binding steps.
func (r *Reconciler) bind(
	logger *log.Log,
	binder *Binder,
	secret *Secret,
	retrievedData map[string][]byte,
	retrievedCache map[string]interface{},
	sbr *v1alpha1.ServiceBindingRequest,
	sbrStatus *v1alpha1.ServiceBindingRequestStatus,
	objectsToAnnotate []*unstructured.Unstructured,
) (reconcile.Result, error) {
	logger = logger.WithName("bind")

	logger.Info("Saving data on intermediary secret...")
	secretObj, err := secret.Commit(retrievedData, retrievedCache)
	if err != nil {
		logger.Error(err, "On saving secret data..")
		return r.onError(err, sbr, sbrStatus, nil)
	}

	// appending intermediary secret in the list of objects to annotate
	objectsToAnnotate = append(objectsToAnnotate, secretObj)
	// making sure secret name is part of status
	r.setSecretName(sbrStatus, secretObj.GetName())

	logger.Info("Binding applications with intermediary secret...")
	updatedObjects, err := binder.Bind()
	if err != nil {
		logger.Error(err, "On binding application.")
		return r.onError(err, sbr, sbrStatus, updatedObjects)
	}

	// saving on status the list of objects that have been touched
	r.setApplicationObjects(sbrStatus, updatedObjects)
	namespacedName := types.NamespacedName{Namespace: sbr.GetNamespace(), Name: sbr.GetName()}

	// annotating objects related to binding
	if err = SetSBRAnnotations(r.dynClient, namespacedName, objectsToAnnotate); err != nil {
		logger.Error(err, "On setting annotations in related objects.")
		return r.onError(err, sbr, sbrStatus, updatedObjects)
	}

	// updating status of request instance
	if sbr, err = r.updateStatusServiceBindingRequest(sbr, sbrStatus); err != nil {
		return RequeueOnConflict(err)
	}

	// appending finalizer, should be later removed upon resource deletion
	if !containsStringSlice(sbr.GetFinalizers(), sbrFinalizer) {
		sbr.SetFinalizers(append(sbr.GetFinalizers(), sbrFinalizer))
		if _, err = r.updateServiceBindingRequest(sbr); err != nil {
			return NoRequeue(err)
		}
	}

	logger.Info("All done!")
	return Done()
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
	ctx := context.TODO()
	objectsToAnnotate := []*unstructured.Unstructured{}

	logger := reconcilerLog.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name,
	)

	logger.Info("Reconciling ServiceBindingRequest...")

	// fetch the ServiceBindingRequest instance
	sbr, err := r.getServiceBindingRequest(request.NamespacedName)
	if err != nil {
		logger.Error(err, "On retrieving service-binding-request instance.")
		return DoneOnNotFound(err)
	}

	logger = logger.WithValues("ServiceBindingRequest.Name", sbr.Name)
	logger.Debug("Found service binding request to inspect")

	// splitting instance from it's status
	sbrStatus := &sbr.Status

	// check Service Binding Request
	if err = checkSBR(sbr, logger); err != nil {
		return RequeueError(err)
	}

	//
	// Planing changes
	//

	logger.Debug("Creating a plan based on OLM and CRD.")
	planner := NewPlanner(ctx, r.dynClient, sbr)
	plan, err := planner.Plan()
	if err != nil {
		logger.Error(err, "On creating a plan to bind applications.")
		return r.onError(err, sbr, sbrStatus, nil)
	}

	// storing CR in objects to annotate
	objectsToAnnotate = append(objectsToAnnotate, plan.CR)

	//
	// Retrieving data
	//

	logger.Debug("Retrieving data to create intermediate secret.")
	retriever := NewRetriever(r.dynClient, plan, sbr.Spec.EnvVarPrefix)
	retrievedData, retrievedCache, err := retriever.Retrieve()
	if err != nil {
		logger.Error(err, "On retrieving binding data.")
		return r.onError(err, sbr, sbrStatus, nil)
	}

	// storing objects used in Retriever
	objectsToAnnotate = append(objectsToAnnotate, retriever.Objects...)

	//
	// Binding and unbind intermediary secret
	//

	secret := NewSecret(r.dynClient, plan)
	binder := NewBinder(ctx, r.client, r.dynClient, sbr, retriever.volumeKeys)

	if sbr.GetDeletionTimestamp() != nil {
		logger.Info("Resource is marked for deletion...")
		return r.unbind(logger, binder, secret, sbr, objectsToAnnotate)
	}

	logger.Info("Starting the bind of application(s) with backing service...")
	return r.bind(logger, binder, secret, retrievedData, retrievedCache, sbr, sbrStatus, objectsToAnnotate)
}
