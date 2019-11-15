package servicebindingrequest

import (
	"os"
	"strings"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// SBRController hold the controller instance and methods for a ServiceBindingRequest.
type SBRController struct {
	Controller   controller.Controller            // controller-runtime instance
	Client       dynamic.Interface                // kubernetes dynamic api client
	watchingGVKs map[schema.GroupVersionKind]bool // cache to identify GVKs on watch
	logger       *log.Log                         // logger instance
}

var (
	// controllerName common name of this controller
	controllerName   = "servicebindingrequest-controller"
	sbrControllerLog = log.NewLog("sbrcontroller")
)

// compareObjectFields compares a nested field of two given objects.
func compareObjectFields(objOld, objNew runtime.Object, fields ...string) (bool, error) {
	var (
		mapNew map[string]interface{}
		mapOld map[string]interface{}
		err    error
	)

	if mapNew, err = runtime.DefaultUnstructuredConverter.ToUnstructured(objNew); err != nil {
		return false, err
	}
	if mapOld, err = runtime.DefaultUnstructuredConverter.ToUnstructured(objOld); err != nil {
		return false, err
	}

	return nestedMapComparison(
		&unstructured.Unstructured{Object: mapNew},
		&unstructured.Unstructured{Object: mapOld},
		fields...)
}

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
	log := s.logger
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
	log.Debug("Amount of GVK founds in CSV objects.", "CSVOwnedGVK.Amount", len(olmGVKs))
	return append(gvks, olmGVKs...), nil
}

func createUpdateFuncWithLog(l *log.Log) func(updateEvent event.UpdateEvent) bool {
	return func(e event.UpdateEvent) bool {
		isSecret := isOfKind(e.ObjectNew, "Secret")
		isConfigMap := isOfKind(e.ObjectNew, "ConfigMap")

		if isSecret || isConfigMap {
			var (
				dataFieldsAreEqual bool
				err                error
			)

			if dataFieldsAreEqual, err = compareObjectFields(e.ObjectNew, e.ObjectOld, "data"); err != nil {
				l.Error(err, "")
				return false
			}

			shouldReconcile := !dataFieldsAreEqual

			l.Debug("Predicate evaluated", "ShouldReconcile", shouldReconcile)
			return shouldReconcile
		}

		// ignore updates to CR status in which case metadata.Generation does not change
		return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
	}
}

func newGVKPredicate(l *log.Log) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: createUpdateFuncWithLog(l),
		DeleteFunc: func(e event.DeleteEvent) bool {
			// evaluates to false if the object has been confirmed deleted
			return !e.DeleteStateUnknown
		},
	}
}

// AddWatchForGVK creates a watch on a given GVK, as long as it's not duplicated.
func (s *SBRController) AddWatchForGVK(gvk schema.GroupVersionKind) error {
	l := s.logger.WithValues("GVK", gvk)
	l.Debug("Adding watch for GVK...")
	if _, exists := s.watchingGVKs[gvk]; exists {
		l.Debug("Skipping watch on GVK twice, it's already under watch!")
		return nil
	}

	// saving GVK in cache
	s.watchingGVKs[gvk] = true

	l.Debug("Creating watch on GVK")
	src := s.createSourceForGVK(gvk)
	return s.Controller.Watch(src, s.newEnqueueRequestsForSBR(), newGVKPredicate(l.WithName("gvk-predicate-log")))
}

// isOfKind evaluates whether the given object has a specific kind.
func isOfKind(obj runtime.Object, kind string) bool {
	return strings.EqualFold(obj.GetObjectKind().GroupVersionKind().Kind, kind)
}

// addCSVWatch creates a watch on ClusterServiceVersion.
func (s *SBRController) addCSVWatch() error {
	log := s.logger
	csvGVK := olmv1alpha1.SchemeGroupVersion.WithKind(ClusterServiceVersionKind)
	source := s.createSourceForGVK(csvGVK)
	err := s.Controller.Watch(source, NewCreateWatchEventHandler(s))
	if err != nil {
		return err
	}
	log.Debug("Watch added for ClusterServiceVersion", "GVK", csvGVK)

	return nil
}

func newSBRPredicate(l *log.Log) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			l.WithName("sbr-create-log").Debug("Predicate evaluated", "ShouldReconcile", true)
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			specsAreEqual, err := compareObjectFields(e.ObjectOld, e.ObjectNew, "spec")
			if err != nil {
				l.Error(err, "")
			}

			shouldReconcile := !specsAreEqual

			l.Debug(
				"Predicate evaluated",
				"Object.New", e.ObjectNew,
				"Object.Old", e.ObjectOld,
				"ShouldReconcile", shouldReconcile)

			return shouldReconcile
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// evaluates to false if the object has been confirmed deleted
			return !e.DeleteStateUnknown
		},
	}
}

// addSBRWatch creates a watchon ServiceBindingRequest GVK.
func (s *SBRController) addSBRWatch() error {
	gvk := v1alpha1.SchemeGroupVersion.WithKind(ServiceBindingRequestKind)
	l := s.logger.WithValues("GKV", gvk)
	src := s.createSourceForGVK(gvk)
	err := s.Controller.Watch(src, s.newEnqueueRequestsForSBR(), newSBRPredicate(l))
	if err != nil {
		l.Error(err, "on creating watch for ServiceBindingRequest")
		return err
	}
	l.Debug("Watch added for ServiceBindingRequest")

	return nil
}

// addWhitelistedGVKWatches create watch on GVKs employed on CSVs.
func (s *SBRController) addWhitelistedGVKWatches() error {
	log := s.logger
	// list of interesting GVKs to watch
	gvks, err := s.getWatchingGVKs()
	if err != nil {
		log.Error(err, "on retrieving list of GVKs to watch")
		return err
	}

	for _, gvk := range gvks {
		log.Debug("Adding watch for whitelisted GVK...", "GVK", gvk)
		err = s.AddWatchForGVK(gvk)
		if err != nil {
			log.Error(err, "on creating watch for GVK")
			return err
		}
	}

	return nil
}

// Watch setup "watch" for all GVKs relevant for SBRController.
func (s *SBRController) Watch() error {
	log := s.logger
	err := s.addSBRWatch()
	if err != nil {
		log.Error(err, "on adding watch for ServiceBindingRequest")
		return err
	}

	err = s.addWhitelistedGVKWatches()
	if err != nil {
		log.Error(err, "on adding watch for whitelisted GVKs")
		return err
	}

	err = s.addCSVWatch()
	if err != nil {
		log.Error(err, "on adding watch for ClusterServiceVersion")
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
		logger:       sbrControllerLog,
	}, nil
}
