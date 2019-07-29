package servicebindingrequest

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/planner"
)

// Reconciler reconciles a ServiceBindingRequest object
type Reconciler struct {
	client client.Client   // kubernetes api client
	scheme *runtime.Scheme // api scheme
}

// Reconcile a ServiceBindingRequest by the following steps:
// 1. Inspecting SBR in order to identify backend service. The service is composed by a CRD name and
//    kind, and by inspecting "connects-to" label identify the name of service instance;
// 2. Using OperatorLifecycleManager standards, identifying which items are intersting for binding
//    by parsing CustomResourceDefinitionDescripton object;
// 3. Search and read contents identified in previous step, creating an intermediary secret to hold
//    data formatted as environment variables key/value.
// 4. Search applications that are interested to bind with given service, by inspecting labels. The
//    Deployment (and other kinds) will be updated in PodTeamplate level updating `envFrom` entry
// 	  to load intermediary secret;
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	logger := logf.Log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling ServiceBindingRequest")

	//
	// Reading ServiceBindingRequest
	//

	instance := &v1alpha1.ServiceBindingRequest{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		logger.Error(err, "Not able to read Service-Binding-Request")
		return RequeueOnNotFound(err)
	}

	logger.WithValues("ServiceBindingRequest.Name", instance.Name).
		Info("Found service binding request to inspect")

	//
	// Planing changes
	//

	bindingPlanner := planner.NewPlanner(ctx, r.client, request.Namespace, instance)
	plan, err := bindingPlanner.Plan()
	if err != nil {
		return RequeueOnNotFound(err)
	}

	//
	// Retrieving data necessary for binding
	//

	retriever := NewRetriever(ctx, r.client, plan)
	if err = retriever.Retrieve(); err != nil {
		return RequeueOnNotFound(err)
	}

	//
	// Updating applications to use intermediary secret
	//

	binder := NewBinder(ctx, r.client, instance)
	if err = binder.Bind(); err != nil {
		return RequeueOnNotFound(err)
	}

	logger.Info("All done!")
	return Done()
}
