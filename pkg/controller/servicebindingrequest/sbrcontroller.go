package servicebindingrequest

import (
	"os"

	"github.com/go-logr/logr"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// SBRController hold the controller instance and methods for a ServiceBindingRequest.
type SBRController struct {
	Controller   controller.Controller            // controller-runtime instance
	Client       dynamic.Interface                // kubernetes dynamic api client
	watchingGVKs map[schema.GroupVersionKind]bool // cache to identify GVKs on watch
	logger       logr.Logger                      // logger instance
}

var (
	// controllerName common name of this controller
	controllerName = "servicebindingrequest-controller"
	// defaultPredicate default predicate functions
	defaultPredicate = predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// evaluates to false if the object has been confirmed deleted
			return !e.DeleteStateUnknown
		},
	}
)

// newEnqueueRequestsForSBR returns a handler.EventHandler configured to map any incoming object to a
// ServiceBindingRequest if it contains the required configuration.
func (s *SBRController) newEnqueueRequestsForSBR() handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &SBRRequestMapper{}}
}

// createSourceForGVK creates a *source.Kind for the given gvk.
func (s *SBRController) createSourceForGVK(gvk schema.GroupVersionKind) *source.Kind {
	return &source.Kind{Type: s.createUnstructuredWithGVK(gvk)}
}

// createUnstructuredWithGVK creates a *unstructured.Unstructured with the given gvk.
func (s *SBRController) createUnstructuredWithGVK(gvk schema.GroupVersionKind) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	return u
}

// getWatchingGVKs return a list of GVKs that this controller is interested in watching.
func (s *SBRController) getWatchingGVKs() ([]schema.GroupVersionKind, error) {
	// standard resources types
	gvks := []schema.GroupVersionKind{
		{Group: "", Version: "v1", Kind: "Secret"},
		{Group: "", Version: "v1", Kind: "ConfigMap"},
	}

	olm := NewOLM(s.Client, os.Getenv("WATCH_NAMESPACE"))
	olmGVKs, err := olm.ListCSVOwnedCRDsAsGVKs()
	if err != nil {
		log.Error(err, "On listing owned CSV as GVKs")
		return nil, err
	}
	log.WithValues("CSVOwnedGVK.Amount", len(olmGVKs)).
		Info("Amount of GVK founds in CSV objects.")
	return append(gvks, olmGVKs...), nil
}

// AddWatchForGVK creates a watch on a given GVK, as long as it's not duplicated.
func (s *SBRController) AddWatchForGVK(gvk schema.GroupVersionKind) error {
	logger := s.logger.WithValues("GVK", gvk)
	logger.Info("Adding watch for GVK...")
	if _, exists := s.watchingGVKs[gvk]; exists {
		logger.Info("Skipping watch on GVK twice, it's already under watch!")
		return nil
	}

	// saving GVK in cache
	s.watchingGVKs[gvk] = true

	logger.Info("Creating watch on GVK")
	source := s.createSourceForGVK(gvk)
	return s.Controller.Watch(source, s.newEnqueueRequestsForSBR(), defaultPredicate)
}

// addCSVWatch creates a watch on ClusterServiceVersion.
func (s *SBRController) addCSVWatch() error {
	csvGVK := olmv1alpha1.SchemeGroupVersion.WithKind(ClusterServiceVersionKind)
	source := s.createSourceForGVK(csvGVK)
	err := s.Controller.Watch(source, NewCreateWatchEventHandler(s), defaultPredicate)
	if err != nil {
		return err
	}
	log.WithValues("GVK", csvGVK).Info("Watch added for ClusterServiceVersion")

	return nil
}

// addSBRWatch creates a watchon ServiceBindingRequest GVK.
func (s *SBRController) addSBRWatch() error {
	gvk := v1alpha1.SchemeGroupVersion.WithKind(ServiceBindingRequestKind)
	logger := s.logger.WithValues("GKV", gvk)
	source := s.createSourceForGVK(gvk)
	err := s.Controller.Watch(source, s.newEnqueueRequestsForSBR(), defaultPredicate)
	if err != nil {
		logger.Error(err, "on creating watch for ServiceBindingRequest")
		return err
	}
	logger.Info("Watch added for ServiceBindingRequest")

	return nil
}

// addWhitelistedGVKWatches create watch on GVKs employed on CSVs.
func (s *SBRController) addWhitelistedGVKWatches() error {
	// list of interesting GVKs to watch
	gvks, err := s.getWatchingGVKs()
	if err != nil {
		s.logger.Error(err, "on retrieving list of GVKs to watch")
		return err
	}

	for _, gvk := range gvks {
		logger := s.logger.WithValues("GVK", gvk)
		logger.Info("Adding watch for whitelisted GVK...")
		err = s.AddWatchForGVK(gvk)
		if err != nil {
			logger.Error(err, "on creating watch for GVK")
			return err
		}
	}

	return nil
}

// Watch setup "watch" for all GVKs relevant for SBRController.
func (s *SBRController) Watch() error {
	err := s.addSBRWatch()
	if err != nil {
		s.logger.Error(err, "on adding watch for ServiceBindingRequest")
		return err
	}

	err = s.addWhitelistedGVKWatches()
	if err != nil {
		s.logger.Error(err, "on adding watch for whitelisted GVKs")
		return err
	}

	err = s.addCSVWatch()
	if err != nil {
		s.logger.Error(err, "on adding watch for ClusterServiceVersion")
		return err
	}

	return nil
}

// NewSBRController creates a new SBRController instance. It can return error on bootstrapping a new
// dynamic client.
func NewSBRController(
	mgr manager.Manager,
	options controller.Options,
	client dynamic.Interface,
) (*SBRController, error) {
	c, err := controller.New(controllerName, mgr, options)
	if err != nil {
		return nil, err
	}

	return &SBRController{
		Controller:   c,
		Client:       client,
		watchingGVKs: make(map[schema.GroupVersionKind]bool),
		logger:       logf.Log.WithName("sbrcontroller"),
	}, nil
}
