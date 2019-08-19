package servicebindingrequest

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	client    client.Client     // kubernetes api client
	dynClient dynamic.Interface // kubernetes dynamic api client
	scheme    *runtime.Scheme   // api scheme
}

const (
	bindingInProgress = "inProgress"
	bindingSuccess    = "success"
	bindingFail       = "fail"
)

// setSecretName update the CR status field to "in progress", and setting secret name.
func (r *Reconciler) setSecretName(
	ctx context.Context,
	instance *v1alpha1.ServiceBindingRequest,
	name string,
) error {
	instance.Status.BindingStatus = bindingInProgress
	instance.Status.Secret = name
	return r.client.Status().Update(ctx, instance)

}

// setStatus update the CR status field.
func (r *Reconciler) setStatus(
	ctx context.Context,
	instance *v1alpha1.ServiceBindingRequest,
	status string,
) error {
	instance.Status.BindingStatus = status
	return r.client.Status().Update(ctx, instance)
}

// setApplicationObjects set the ApplicationObject status field, and also set the overall status as
// success, since it was able to bind applications.
func (r *Reconciler) setApplicationObjects(
	ctx context.Context,
	instance *v1alpha1.ServiceBindingRequest,
	objs []string,
) error {
	instance.Status.BindingStatus = bindingSuccess
	instance.Status.ApplicationObjects = objs
	return r.client.Status().Update(ctx, instance)
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
	logger := logf.Log.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name,
	)
	logger.Info("Reconciling ServiceBindingRequest...")

	// fetch the ServiceBindingRequest instance
	instance := &v1alpha1.ServiceBindingRequest{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		logger.Error(err, "On retrieving service-binding-request instance.")
		return RequeueOnNotFound(err)
	}

	logger = logger.WithValues("ServiceBindingRequest.Name", instance.Name)
	logger.Info("Found service binding request to inspect")

	if err = r.setStatus(ctx, instance, bindingInProgress); err != nil {
		logger.Error(err, "On updating service-binding-request status.")
		return RequeueError(err)
	}

	//
	// Planing changes
	//

	logger.Info("Creating a plan based on OLM and CRD.")
	planner := NewPlanner(ctx, r.dynClient, instance)
	plan, err := planner.Plan()
	if err != nil {
		_ = r.setStatus(ctx, instance, bindingFail)
		logger.Error(err, "On creating a plan to bind applications.")
		return RequeueOnNotFound(err)
	}

	//
	// Retrieving data
	//

	logger.Info("Retrieving data to create intermediate secret.")
	retriever := NewRetriever(ctx, r.client, plan, instance.Spec.EnvVarPrefix)
	if err = retriever.Retrieve(); err != nil {
		_ = r.setStatus(ctx, instance, bindingFail)
		logger.Error(err, "On retrieving binding data.")
		return RequeueOnNotFound(err)
	}

	if err = r.setSecretName(ctx, instance, plan.Name); err != nil {
		logger.Error(err, "On updating service-binding-request status.")
		return RequeueError(err)
	}

	//
	// Updating applications to use intermediary secret
	//

	logger.Info("Binding applications with intermediary secret.")
	binder := NewBinder(ctx, r.client, r.dynClient, instance, retriever.volumeKeys)
	if err = binder.Bind(); err != nil {
		_ = r.setStatus(ctx, instance, bindingFail)
		logger.Error(err, "On binding application.")
		return RequeueOnNotFound(err)
	}

	if err = r.setApplicationObjects(ctx, instance, binder.UpdatedObjectNames); err != nil {
		logger.Error(err, "On updating application objects status field.")
		return RequeueError(err)
	}

	logger.Info("All done!")
	return Done()
}
