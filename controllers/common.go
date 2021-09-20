package controllers

import (
	ctx "context"
	"flag"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"

	"github.com/go-logr/logr"
	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list
// +kubebuilder:rbac:groups="operators.coreos.com",resources=clusterserviceversions,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=get;list;update;patch
// +kubebuilder:rbac:groups=apps.openshift.io,resources=deploymentconfigs,verbs=get;list;update;patch
// +kubebuilder:rbac:groups="",resources=pods;secrets;services;endpoints;configmaps,verbs=get;list
// +kubebuilder:rbac:groups="",resources=pods;secrets,verbs=update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=create

var (
	MaxConcurrentReconciles int
)

func RegisterFlags(flags *flag.FlagSet) {
	flags.IntVar(&MaxConcurrentReconciles, "max-concurrent-reconciles", 1, "max-concurrent-reconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 1.")
}

type BindingReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	pipeline pipeline.Pipeline

	PipelineProvider func(*rest.Config, kubernetes.K8STypeLookup) (pipeline.Pipeline, error)

	ReconcilingObject func() apis.Object
}

// SetupWithManager sets up the controller with the Manager.
func (r *BindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pipeline, err := r.PipelineProvider(mgr.GetConfig(), kubernetes.ResourceLookup(mgr.GetRESTMapper()))
	if err != nil {
		return err
	}
	r.pipeline = pipeline
	return ctrl.NewControllerManagedBy(mgr).
		For(r.ReconcilingObject()).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: MaxConcurrentReconciles}).
		Complete(r)
}

// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups=authorization.k8s.io,resources=selfsubjectaccessreviews,verbs=create
// +kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *BindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("serviceBinding", req.NamespacedName)
	var ctx = ctx.Background()
	serviceBinding := r.ReconcilingObject()

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
	if !serviceBinding.HasDeletionTimestamp() && apis.MaybeAddFinalizer(serviceBinding) {
		if err = r.Update(ctx, serviceBinding); err != nil {
			return ctrl.Result{}, err
		}
	}
	if !serviceBinding.HasDeletionTimestamp() {
		log.Info("Reconciling")
	} else {
		log.Info("Deleted, unbind the application")
	}
	retry, err := r.pipeline.Process(serviceBinding)
	if !retry && err == nil {
		if serviceBinding.HasDeletionTimestamp() {
			if apis.MaybeRemoveFinalizer(serviceBinding) {
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
