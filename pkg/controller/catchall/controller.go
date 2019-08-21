package catchall

import (
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

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("catchall-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	var listFunc cache.ListFunc
	var watchFunc cache.WatchFunc

	switch v := r.(type) {
	// TODO: Check why CatchAllReconciler doesn't implement the right interface.
	case CatchAllReconciler:
		gvr := schema.GroupVersionResource{}
		listFunc = asUnstructuredLister(v.DynClient.Resource(gvr).Namespace("").List)
		watchFunc = asUnstructuredWatcher(v.DynClient.Resource(gvr).Namespace("").Watch)
	default:
		panic("WIP, r is not a CatchAllReconciler")
	}

	lw := &cache.ListWatch{
		ListFunc:  listFunc,
		WatchFunc: watchFunc,
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
