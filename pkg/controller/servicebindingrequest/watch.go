package servicebindingrequest

import (
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WatcherMapper struct {
	c *SBRController
}

func (w *WatcherMapper) Map(obj handler.MapObject) []reconcile.Request {
	err := w.c.AddWatchForGVK(obj.Object.GetObjectKind().GroupVersionKind())
	if err != nil {
		// ???
	}

	return []reconcile.Request{}
}

func NewCreateWatchEventHandler(c *SBRController) handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &WatcherMapper{}}
}
