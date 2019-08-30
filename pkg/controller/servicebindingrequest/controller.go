package servicebindingrequest

import (
	"os"

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
	client, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	r, err := newReconciler(mgr, client)
	if err != nil {
		return err
	}
	return add(mgr, r, client)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, client dynamic.Interface) (reconcile.Reconciler, error) {
	return &Reconciler{client: mgr.GetClient(), dynClient: client, scheme: mgr.GetScheme()}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler, client dynamic.Interface) error {
	opts := controller.Options{Reconciler: r}
	_, err := NewSBRController(mgr, opts, client)
	if err != nil {
		return err
	}

	return nil
}

type SBRController struct {
	Controller controller.Controller
	Client     dynamic.Interface
}

func NewSBRController(
	mgr manager.Manager,
	options controller.Options,
	client dynamic.Interface,
) (*SBRController, error) {
	c, err := controller.New("service-binding-controller", mgr, options)
	if err != nil {
		return nil, err
	}

	err = addServiceBindingRequestWatch(c)
	if err != nil {
		return nil, err
	}

	err = addDynamicGVKsWatches(c, client)
	if err != nil {
		return nil, err
	}

	return &SBRController{
		Controller: c,
		Client:     client,
	}, nil
}

func (sbrc *SBRController) AddWatchForGVK(kind schema.GroupVersionKind) error {
	err := createWatch(sbrc.Controller, kind)
	if err != nil {
		return err
	}
	return nil
}

func addServiceBindingRequestWatch(c controller.Controller) error {
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

	// watching operator's main CRD -- ServiceBindingRequest
	sbrGVK := v1alpha1.SchemeGroupVersion.WithKind(ServiceBindingRequestKind)
	err := c.Watch(createSourceForGVK(sbrGVK), newEnqueueRequestsForSBR(), pred)
	if err != nil {
		return err
	}
	log.WithValues("GroupVersionKind", sbrGVK).Info("Watch added for ServiceBindingRequest")

	return nil
}

func addDynamicGVKsWatches(
	controller controller.Controller,
	client dynamic.Interface,
) error {
	// list of interesting GVKs to watch
	gvks, err := getWatchingGVKs(client)
	if err != nil {
		return err
	}

	for _, gvk := range gvks {
		err = createWatch(controller, gvk)
		if err != nil {
			return err
		}
	}

	return nil
}

func createWatch(
	c controller.Controller,
	gvk schema.GroupVersionKind,
) error {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// evaluates to false if the object has been confirmed deleted
			return !e.DeleteStateUnknown
		},
	}

	log.WithValues("GroupVersionKind", gvk).Info("Watch added")
	return c.Watch(createSourceForGVK(gvk), newEnqueueRequestsForSBR(), pred)
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

// getWatchingGVKs return a list of GVKs that this controller is interested in watching.
func getWatchingGVKs(client dynamic.Interface) ([]schema.GroupVersionKind, error) {
	// standard resources types
	gvks := []schema.GroupVersionKind{
		{Group: "", Version: "v1", Kind: "Secret"},
		{Group: "", Version: "v1", Kind: "ConfigMap"},
	}

	olm := NewOLM(client, os.Getenv("WATCH_NAMESPACE"))
	olmGVKs, err := olm.ListCSVOwnedCRDsAsGVKs()
	if err != nil {
		log.Error(err, "On listing owned CSV as GVKs")
		return nil, err
	}
	log.WithValues("CSVOwnedGVK.Amount", len(olmGVKs)).
		Info("Amount of GVK founds in CSV objects.")
	return append(gvks, olmGVKs...), nil
}

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}
