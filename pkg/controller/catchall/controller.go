package catchall

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
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

	// TODO: This should be configured externally.
	resyncPeriod := 5 * time.Minute

	for _, gvr := range getGVRs() {
		resourceClient := r.DynClient.Resource(gvr).Namespace("")

		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc:  asUnstructuredLister(resourceClient.List),
				WatchFunc: asUnstructuredWatcher(resourceClient.Watch),
			},
			&unstructured.Unstructured{},
			resyncPeriod,
			nil,
		)

		err = c.Watch(
			&source.Informer{Informer: informer},
			nil,
			nil)
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

func getGVRs() []schema.GroupVersionResource {
	return []schema.GroupVersionResource{}
}
