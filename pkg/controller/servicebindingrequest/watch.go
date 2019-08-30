package servicebindingrequest

import (
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WatcherMapper struct {
	c *SBRController
}

func (w *WatcherMapper) Map(obj handler.MapObject) []reconcile.Request {
	olm := NewOLM(w.c.Client, obj.Meta.GetNamespace())

	csvGVKs, err := olm.ListCSVOwnedCRDsAsGVKs()
	if err != nil {
		log.Error(err, "Failed to list CRDs as GVKs")
	}

	for _, gvk := range csvGVKs {
		err := w.c.AddWatchForGVK(gvk)
		if err != nil {
			log.WithValues("GroupVersionKind", gvk).Error(err, "Failed to create a watch")
		}
	}

	return []reconcile.Request{}
}

func NewCreateWatchEventHandler(c *SBRController) handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &WatcherMapper{}}
}
