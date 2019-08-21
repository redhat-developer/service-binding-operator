package catchall

import (
	"errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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

func getDynClient(r reconcile.Reconciler) (dynamic.Interface, error) {
	if v, ok := r.(*CatchAllReconciler); ok {
		return v.DynClient, nil
	}
	return nil, errors.New("given argument is not a CatchAllReconciler instance")
}

func getGVRs() []schema.GroupVersionResource {
	return []schema.GroupVersionResource{}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("catchall-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	dynClient, err := getDynClient(r)
	if err != nil {
		return err
	}

	for _, gvr := range getGVRs() {

		namespacedResource := dynClient.Resource(gvr).Namespace("")

		lw := &cache.ListWatch{
			ListFunc:  asUnstructuredLister(namespacedResource.List),
			WatchFunc: asUnstructuredWatcher(namespacedResource.Watch),
		}

		resyncPeriod := 5 * time.Minute

		informer := cache.NewSharedIndexInformer(
			lw,
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

func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
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
