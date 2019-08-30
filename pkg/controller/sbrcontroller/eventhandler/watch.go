package eventhandler

import (
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/sbrcontroller"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type CreateWatchEventHandler struct {
	c *sbrcontroller.SBRController
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

func NewCreateWatchEventHandler(c *sbrcontroller.SBRController) *CreateWatchEventHandler {
	return &CreateWatchEventHandler{
		c: c,
	}
}
