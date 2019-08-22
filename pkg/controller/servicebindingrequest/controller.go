package servicebindingrequest

import (
	"context"
	"fmt"
	"os"
	"strings"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	scheme "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructuredscheme"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

var watchedOperators []string

// Add creates a new ServiceBindingRequest Controller and adds it to the Manager. The Manager will
// set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r, r.handleCRDChange, r.handleCSVChange)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (*Reconciler, error) {
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}
	return &Reconciler{client: mgr.GetClient(), dynClient: dynClient, scheme: mgr.GetScheme()}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler, crdHandler handler.ToRequestsFunc, csvHandler handler.ToRequestsFunc) error {
	// Create a new controller
	c, err := controller.New("servicebindingrequest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}

	// Watch for changes to primary resource ServiceBindingRequest
	err = c.Watch(&source.Kind{Type: &v1alpha1.ServiceBindingRequest{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	err = addWatchesToKnownCRDs(mgr, r, crdHandler, c)
	if err != nil {
		return err
	}

	err = addWatchToCSVList(mgr, r, csvHandler, c)
	if err != nil {
		return err
	}

	return err
}

func addWatchToCSVList(mgr manager.Manager, r reconcile.Reconciler, crdHandler handler.ToRequestsFunc, c controller.Controller) error {

	predicateForFilteringCSV := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}

	olmGVR := schema.GroupVersionKind{
		Group:   olmv1alpha1.GroupName,
		Version: olmv1alpha1.GroupVersion,
		Kind:    "ClusterServiceVersion",
	}

	objRuntime, _ := scheme.NewUnstructuredCreator().New(olmGVR)
	handlerCRD := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: crdHandler,
	}
	fmt.Println("Adding a watch on " + olmGVR.String())
	return c.Watch(&source.Kind{Type: objRuntime}, handlerCRD, predicateForFilteringCSV)

}

func addWatchesToKnownCRDs(mgr manager.Manager, r reconcile.Reconciler, crdHandler handler.ToRequestsFunc, c controller.Controller) error {

	olmGVR := schema.GroupVersionResource{
		Group:    olmv1alpha1.GroupName,
		Version:  olmv1alpha1.GroupVersion,
		Resource: "clusterserviceversions",
	}

	// Create a dynamic client to fetch all CSVs
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}

	// Get all CSVs into an Unstructured List
	allCSV, err := dynClient.Resource(olmGVR).Namespace("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, csv := range allCSV.Items {

		// Convert each Unstructured item into a CSV
		typedCSV := olmv1alpha1.ClusterServiceVersion{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(csv.Object, &typedCSV)
		if err != nil {
			return err
		}

		if !isOperatorCRDsBeingWatched(typedCSV.GetName()) {
			watchedOperators = append(watchedOperators, typedCSV.Name)

			// For each CSV, iterate through all OWNED CRDs
			for _, crd := range typedCSV.Spec.CustomResourceDefinitions.Owned {

				err = watchCRD(crd, crdHandler, c)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func watchCRD(crd olmv1alpha1.CRDDescription, crdHandler handler.ToRequestsFunc, c controller.Controller) error {

	// a bit messy, needs cleanup
	rsrc := strings.Split(crd.Name, ".")[0]

	grp := strings.Split(crd.Name, rsrc+".")
	fmt.Println(grp[1])

	gvk := schema.GroupVersionKind{
		Group:   grp[1],
		Version: crd.Version,
		Kind:    crd.Kind,
	}

	objRuntime, _ := scheme.NewUnstructuredCreator().New(gvk)
	handlerCRD := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: crdHandler,
	}
	fmt.Println("Adding a watch on " + gvk.String())
	return c.Watch(&source.Kind{Type: objRuntime}, handlerCRD)
}

func isOperatorCRDsBeingWatched(i string) bool {

	present := false
	for _, ele := range watchedOperators {
		if ele == i {
			present = true
		}
	}
	return present

}

func (r *Reconciler) handleCRDChange(o handler.MapObject) []reconcile.Request {
	var result []reconcile.Request

	sbr := &v1alpha1.ServiceBindingRequestList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingRequest",
			APIVersion: "apps.openshift.io/v1alpha1",
		},
	}

	// Get all ServiceBindingRequests
	currentNamespace := &client.ListOptions{Namespace: o.Meta.GetNamespace()}
	if err := r.client.List(context.TODO(), currentNamespace, sbr); err != nil {
		return result
	}

	// If any of the SBRs consume this CR, then we need
	// to trigger reconcile for that SBR.
	for _, sbr := range sbr.Items {

		if sbr.Spec.BackingServiceSelector.ResourceRef == o.Meta.GetName() &&
			sbr.Spec.BackingServiceSelector.Kind == o.Object.GetObjectKind().GroupVersionKind().Kind {
			// could check group and version as well.

			result = append(result, reconcile.Request{
				NamespacedName: client.ObjectKey{Namespace: sbr.Namespace, Name: sbr.Name}})
		}
	}
	return result

}

func (r *Reconciler) handleCSVChange(o handler.MapObject) []reconcile.Request {
	var result []reconcile.Request

	// A new CSV doesn't necessarily imply that a new operator has been installed.
	// When a new namespace is created, CVSs are copied over to the new
	// namespace.
	if isOperatorCRDsBeingWatched(o.Meta.GetName()) {
		// This operator's CRDs are already being watched, Do nothing.
		return result
	}
	// Control comes here if there's a new operator - so kill the container
	// The Pod controller will create a new one anyway.
	os.Exit(0)
	return result
}

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}
