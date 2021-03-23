package controllers

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// sbrController hold the controller instance and methods for a ServiceBinding.
type sbrController struct {
	Controller   controller.Controller // controller-runtime instance
	Client       dynamic.Interface     // kubernetes dynamic api client
	typeLookup   K8STypeLookup
	watchingGVKs map[schema.GroupVersionKind]bool // cache to identify GVKs on watch
	logger       *log.Log                         // logger instance
}

// controllerName common name of this controller
const controllerName = "servicebinding-controller"

// compareObjectFields compares a nested field of two given objects.
func compareObjectFields(objOld, objNew runtime.Object, fields ...string) (*comparisonResult, error) {
	mapNew, err := runtime.DefaultUnstructuredConverter.ToUnstructured(objNew)
	if err != nil {
		return nil, err
	}
	mapOld, err := runtime.DefaultUnstructuredConverter.ToUnstructured(objOld)
	if err != nil {
		return nil, err
	}

	return nestedUnstructuredComparison(
		&unstructured.Unstructured{Object: mapNew},
		&unstructured.Unstructured{Object: mapOld},
		fields...,
	)
}

// newEnqueueRequestsForSBR returns a handler.EventHandler configured to map any incoming object to a
// ServiceBinding if it contains the required configuration.
func (s *sbrController) newEnqueueRequestsForSBR() handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &sbrRequestMapper{
		client:     s.Client,
		typeLookup: s.typeLookup,
	}}
}

// createSourceForGVK creates a *source.Kind for the given gvk.
func (s *sbrController) createSourceForGVK(gvk schema.GroupVersionKind) *source.Kind {
	return &source.Kind{Type: s.createUnstructuredWithGVK(gvk)}
}

// createUnstructuredWithGVK creates a *unstructured.Unstructured with the given gvk.
func (s *sbrController) createUnstructuredWithGVK(gvk schema.GroupVersionKind) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	return u
}

// buildSBRPredicate construct the predicates for service-bindings.
func buildSBRPredicate(logger *log.Log) predicate.Funcs {
	logger = logger.WithName("buildSBRPredicate")
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			logger.Debug("Create Predicate", "reconcile", true)
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// should reconcile when resource is marked for deletion
			if e.MetaNew != nil && e.MetaNew.GetDeletionTimestamp() != nil {
				logger.Debug("Executing reconcile, object is marked for deletion.")
				return true
			}

			// verifying if the actual spec field of the object has changed, should reconcile when
			// not equals
			specsAreEqual, err := compareObjectFields(e.ObjectOld, e.ObjectNew, "spec")
			if err != nil {
				logger.Error(err, "")
			}
			logger.Debug("Predicate evaluated", "specsAreEqual", specsAreEqual.Success)
			if !specsAreEqual.Success {
				logger.Trace("Specs are not equal", "ObjectOld", e.ObjectOld, "ObjectNew", e.ObjectNew)
			}
			return !specsAreEqual.Success
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// evaluates to false, if the object is confirmed deleted
			reconcile := !e.DeleteStateUnknown
			logger.Debug("Delete Predicate", "reconcile", reconcile)
			return reconcile
		},
	}
}

// addSBRWatch creates a watchon ServiceBinding GVK.
func (s *sbrController) addSBRWatch() error {
	l := s.logger.WithValues("GVK", v1alpha1.GroupVersionKind)
	src := s.createSourceForGVK(v1alpha1.GroupVersionKind)
	err := s.Controller.Watch(src, s.newEnqueueRequestsForSBR(), buildSBRPredicate(l))
	if err != nil {
		l.Error(err, "on creating watch for ServiceBinding")
		return err
	}
	l.Debug("Watch added for ServiceBinding")

	return nil
}

// Watch setup "watch" for all GVKs relevant for SBRController.
func (s *sbrController) Watch() error {
	log := s.logger
	err := s.addSBRWatch()
	if err != nil {
		log.Error(err, "on adding watch for ServiceBinding")
		return err
	}

	return nil
}

// NewSBRController creates a new SBRController instance. It can return error on bootstrapping a new
// dynamic client.
//func NewSBRController(
//	mgr manager.Manager,
//	controller controller.Controller,,
//	client dynamic.Interface,
//) (*sbrController, error) {
//	c, err := controller.New(controllerName, mgr, options)
//	if err != nil {
//		return nil, err
//	}
//
//	return &sbrController{
//		Controller:   c,
//		Client:       client,
//		RestMapper:   mgr.GetRESTMapper(),
//		watchingGVKs: make(map[schema.GroupVersionKind]bool),
//		logger:       log.NewLog("sbrcontroller"),
//	}, nil
//}
