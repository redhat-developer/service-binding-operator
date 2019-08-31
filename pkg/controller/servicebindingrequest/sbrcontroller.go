package servicebindingrequest

import (
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	controllerName = "servicebindingrequest-controller"
)

type SBRController struct {
	Controller controller.Controller
	Client     dynamic.Interface
}

func NewSBRController(
	mgr manager.Manager,
	options controller.Options,
	client dynamic.Interface,
) (*SBRController, error) {
	c, err := controller.New(controllerName, mgr, options)
	if err != nil {
		return nil, err
	}

	sbrc := &SBRController{
		Controller: c,
		Client:     client,
	}

	// NOTE: Perhaps those two methods can compose another method, responsible only for the actual
	// 		 initialization.

	err = sbrc.addServiceBindingRequestWatch()
	if err != nil {
		return nil, err
	}

	err = sbrc.addWhitelistedGVKWatches()
	if err != nil {
		return nil, err
	}

	err = sbrc.addCSVWatch()
	if err != nil {
		// FIXME: Expose errors in logs.
		return nil, err
	}

	return sbrc, nil
}

func (sbrc *SBRController) AddWatchForGVK(gvk schema.GroupVersionKind) error {
	return createWatch(sbrc.Controller, gvk)
}

func (sbrc *SBRController) addCSVWatch() error {
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

	csvGVK := olmv1alpha1.SchemeGroupVersion.WithKind("ClusterServiceVersion")
	err := sbrc.Controller.Watch(createSourceForGVK(csvGVK), NewCreateWatchEventHandler(), pred)
	if err != nil {
		return err
	}
	log.WithValues("GroupVersionKind", csvGVK).Info("Watch added for ClusterServiceVersion")

	return nil
}

func (sbrc *SBRController) addServiceBindingRequestWatch() error {
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
	err := sbrc.Controller.Watch(createSourceForGVK(sbrGVK), newEnqueueRequestsForSBR(), pred)
	if err != nil {
		return err
	}
	log.WithValues("GroupVersionKind", sbrGVK).Info("Watch added for ServiceBindingRequest")

	return nil
}

func (sbrc *SBRController) addWhitelistedGVKWatches() error {
	c := sbrc.Controller
	client := sbrc.Client

	// list of interesting GVKs to watch
	gvks, err := getWatchingGVKs(client)
	if err != nil {
		return err
	}

	for _, gvk := range gvks {
		err = createWatch(c, gvk)
		if err != nil {
			return err
		}
	}

	return nil
}

func createWatch(c controller.Controller, gvk schema.GroupVersionKind) error {
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

	// FIXME: Do not watch the same GVK twice.
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
