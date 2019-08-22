package catchall

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

func add(mgr manager.Manager, r *CatchAllReconciler) error {
	c, err := controller.New("catchall-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	for _, gvk := range getGVKs() {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		err = c.Watch(&source.Kind{Type: u}, &EnqueueRequestForUnstructured{})
		if err != nil {
			return err
		}
	}

	return nil
}

func newReconciler(mgr manager.Manager) (*CatchAllReconciler, error) {
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}

	return &CatchAllReconciler{
		Client:    mgr.GetClient(),
		DynClient: dynClient,
		Scheme:    mgr.GetScheme(),
	}, nil
}

func getGVKs() []schema.GroupVersionKind {
	return []schema.GroupVersionKind{
		{Group: "apps.openshift.io", Version: "v1alpha1", Kind: "ServiceBindingRequest"},
	}
}
