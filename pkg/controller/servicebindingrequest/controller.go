package servicebindingrequest

import (
	"context"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new ServiceBindingRequest Controller and adds it to the Manager. The Manager will
// set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r, r.reconcileIfAssociatedWithAServiceBinding)
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
func add(mgr manager.Manager, r reconcile.Reconciler, nonServiceBindingOwnedTrigger handler.ToRequestsFunc) error {
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

	handlerSecret := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: nonServiceBindingOwnedTrigger,
	}
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, handlerSecret)
	if err != nil {
		return err
	}

	handlerConfigMap := &handler.EnqueueRequestsFromMapFunc{
		ToRequests: nonServiceBindingOwnedTrigger,
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, handlerConfigMap)
	if err != nil {
		return err
	}

	return nil
}

// reconcileIfAssociatedWithAServiceBinding triggers a reconcile event
// if a secret/configmap owned by any of the BackingServices changes.
func (r *Reconciler) reconcileIfAssociatedWithAServiceBinding(o handler.MapObject) []reconcile.Request {
	var result []reconcile.Request

	var objOwner *metav1.OwnerReference
	for _, owner := range o.Meta.GetOwnerReferences() {
		objOwner = &owner
		if owner.Kind == "ServiceBindingRequest" {
			// Typical reqeue for Owner. Nothing fancy here.
			return append(result, reconcile.Request{
				NamespacedName: client.ObjectKey{Namespace: o.Meta.GetNamespace(), Name: owner.Name}})
		}
	}

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

	for _, sbr := range sbr.Items {
		// if the secret/configmap is owned  a CR which was bound in
		// a ServiceBindingRequest previously, reconcile is needed.

		if objOwner != nil {
			if sbr.Spec.BackingServiceSelector.ResourceRef == objOwner.Name &&
				sbr.Spec.BackingServiceSelector.Kind == objOwner.Kind {

				result = append(result, reconcile.Request{
					NamespacedName: client.ObjectKey{Namespace: sbr.Namespace, Name: sbr.Name}})
			}

		} else {
			plannerRef := NewPlanner(context.TODO(), r.dynClient, &sbr)
			plan, err := plannerRef.Plan()
			if err != nil {
				continue
			}
			retrieverObj := NewRetriever(context.TODO(), r.client, plan, "")
			if o.Object.GetObjectKind().GroupVersionKind().Kind == "Secret" {
				for _, s := range retrieverObj.secrets {
					// if it happens to be one of the secrets consumed
					// by the CR in the spec but not necessarily owned
					// by the CR.
					if s.Name == o.Meta.GetName() {
						result = append(result, reconcile.Request{
							NamespacedName: client.ObjectKey{Namespace: sbr.Namespace, Name: sbr.Name}})
						break
					}
				}
			} else if o.Object.GetObjectKind().GroupVersionKind().Kind == "ConfigMap" {
				for _, s := range retrieverObj.configmaps {
					// if it happens to be one of the configmaps
					// consumed by the CR in the spec
					// but not necessarily owned by the CR.
					if s.Name == o.Meta.GetName() {
						result = append(result, reconcile.Request{
							NamespacedName: client.ObjectKey{Namespace: sbr.Namespace, Name: sbr.Name}})
						break
					}
				}
			}
		}
	}
	return result
}

// blank assignment to verify that ReconcileServiceBindingRequest implements reconcile.Reconciler
var _ reconcile.Reconciler = &Reconciler{}
