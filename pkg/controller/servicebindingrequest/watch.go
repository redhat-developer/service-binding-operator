package servicebindingrequest

import (
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WatcherMapper struct {
	c *SBRController
}

func (w *WatcherMapper) Map(obj handler.MapObject) []reconcile.Request {
	gvk := obj.Object.GetObjectKind().GroupVersionKind()
	err := w.c.AddWatchForGVK(gvk)
	if err != nil {
		log.WithValues("GroupVersionKind", gvk).Error(err, "Failed to create a watch")
	}

	return []reconcile.Request{}
}

func NewCreateWatchEventHandler(c *SBRController) handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &WatcherMapper{}}
}
