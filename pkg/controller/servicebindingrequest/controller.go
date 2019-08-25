package servicebindingrequest

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

var (
	controllerName = "servicebindingrequest-controller"
)

// Add creates a new ServiceBindingRequest Controller and adds it to the Manager. The Manager will
// set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}
	return &Reconciler{client: mgr.GetClient(), dynClient: dynClient, scheme: mgr.GetScheme()}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	opts := controller.Options{Reconciler: r}
	c, err := controller.New(controllerName, mgr, opts)
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// FIXME: support unstructured.Unstructured types. This block is currently causing a
			// panic with the assertion error.
			/*
				oldKind := e.ObjectOld.DeepCopyObject().GetObjectKind().GroupVersionKind().Kind
				if oldKind == ServiceBindingRequestKind {
					oldSBR := e.ObjectOld.(*v1alpha1.ServiceBindingRequest)
					newSBR := e.ObjectNew.(*v1alpha1.ServiceBindingRequest)

					// if event was triggered as part of resetting TriggerRebinding True to False,
					// we shall ignore it to avoid an infinite loop.
					if newSBR.Spec.TriggerRebinding != nil &&
						oldSBR.Spec.TriggerRebinding != nil &&
						!*newSBR.Spec.TriggerRebinding &&
						*oldSBR.Spec.TriggerRebinding {
						return false
					}
				}
			*/

			// ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// evaluates to false if the object has been confirmed deleted
			return !e.DeleteStateUnknown
		},
	}

	for _, gvk := range getWatchingGVKs() {
		logger := log.WithValues("GroupVersionKind", gvk)
		err = c.Watch(createSourceForGVK(gvk), newEnqueueRequestsForSBR(), pred)
		if err != nil {
			return err
		}
		logger.Info("Watch added")
	}

	return nil
}

// newEnqueueRequestsForSBR returns a handler.EventHandler configured to map any incoming object to a
// ServiceBindingRequest if it contains the required configuration.
func newEnqueueRequestsForSBR() handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &SBRRequestMapper{}}
}

// createSourceForGVK creates a *source.Kind for the given gvk.
func createSourceForGVK(gvk schema.GroupVersionKind) *source.Kind {
	return &source.Kind{Type: createUnstructuredWithGVK(gvk)}
}

// createUnstructuredWithGVK creates a *unstructured.Unstructured with the given gvk.
func createUnstructuredWithGVK(gvk schema.GroupVersionKind) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	return u
}

func getWatchingGVKs() []schema.GroupVersionKind {
	return []schema.GroupVersionKind{
		v1alpha1.SchemeGroupVersion.WithKind("ServiceBindingRequest"),
		{Group: "", Version: "v1", Kind: "Secret"},
		{Group: "", Version: "v1", Kind: "ConfigMap"},
	}
}

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}
