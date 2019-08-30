package servicebindingrequest

import (
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type CreateWatchEventHandler struct {
	c *SBRController
}

func (h CreateWatchEventHandler) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	err := h.c.AddWatchForGVK(e.Object.GetObjectKind().GroupVersionKind())
	if err != nil {
		// ???
	}
}

func (h CreateWatchEventHandler) Update(event.UpdateEvent, workqueue.RateLimitingInterface) {
}

func (h CreateWatchEventHandler) Delete(event.DeleteEvent, workqueue.RateLimitingInterface) {
}

func (h CreateWatchEventHandler) Generic(event.GenericEvent, workqueue.RateLimitingInterface) {
}

func NewCreateWatchEventHandler(c *SBRController) *CreateWatchEventHandler {
	return &CreateWatchEventHandler{
		c: c,
	}
}
