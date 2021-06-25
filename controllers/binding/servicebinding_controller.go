/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package binding

import (
	ctx "context"
	v1alpha12 "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/controllers"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/builder"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	dynClient dynamic.Interface // kubernetes dynamic api client

	pipeline pipeline.Pipeline
}

// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods;services;endpoints;persistentvolumeclaims;events;configmaps;secrets,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=*
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ServiceBindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("serviceBinding", req.NamespacedName)
	var ctx = ctx.Background()
	serviceBinding := &v1alpha12.ServiceBinding{}

	err := r.Get(ctx, req.NamespacedName, serviceBinding)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ServiceBinding resource not found. Ignoring since object must be deleted", "name", req.NamespacedName, "err", err)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get ServiceBinding", "name", req.NamespacedName, "err", err)
		return ctrl.Result{}, err
	}
	if serviceBinding.DeletionTimestamp.IsZero() && serviceBinding.MaybeAddFinalizer() {
		if err = r.Update(ctx, serviceBinding); err != nil {
			return ctrl.Result{}, err
		}
	}
	if serviceBinding.DeletionTimestamp.IsZero() {
		log.Info("Reconciling")
	} else {
		log.Info("Deleted, unbind the application")
	}
	retry, err := r.pipeline.Process(serviceBinding)
	if !retry && err == nil {
		if !serviceBinding.DeletionTimestamp.IsZero() {
			if serviceBinding.MaybeRemoveFinalizer() {
				if err = r.Update(ctx, serviceBinding); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}
	result := ctrl.Result{Requeue: retry}
	log.Info("Done", "retry", retry, "error", err)
	if retry {
		return result, err
	}
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	client, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}

	r.dynClient = client

	r.pipeline = builder.DefaultBuilder.WithContextProvider(context.Provider(r.dynClient, context.ResourceLookup(mgr.GetRESTMapper()))).Build()

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha12.ServiceBinding{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: controllers.MaxConcurrentReconciles}).
		Complete(r)
}
