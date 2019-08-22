package catchall

import (
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/common"
)

var log = logf.KBLog.WithName("eventhandler").WithName("EnqueueRequestForObject")

type EnqueueRequestForUnstructured struct{}

func (e EnqueueRequestForUnstructured) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := common.GetSBRSelectorFromObject(evt.Object)
	if err != nil {
		log.Error(err, "error on extracting SBR namespaced-name from annotations")
		return
	}
	if common.IsSBRSelectorEmpty(sbrSelector) {
		return
	}

	q.Add(reconcile.Request{NamespacedName: sbrSelector})
}

func (e EnqueueRequestForUnstructured) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := common.GetSBRSelectorFromObject(evt.ObjectNew)
	if err != nil {
		log.Error(err, "error on extracting SBR namespaced-name from annotations")
		return
	}
	if common.IsSBRSelectorEmpty(sbrSelector) {
		return
	}

	q.Add(reconcile.Request{NamespacedName: sbrSelector})
}

func (e EnqueueRequestForUnstructured) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := common.GetSBRSelectorFromObject(evt.Object)
	if err != nil {
		log.Error(err, "error on extracting SBR namespaced-name from annotations")
		return
	}
	if common.IsSBRSelectorEmpty(sbrSelector) {
		return
	}

	q.Add(reconcile.Request{sbrSelector})
}

func (e EnqueueRequestForUnstructured) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := common.GetSBRSelectorFromObject(evt.Object)
	if err != nil {
		log.Error(err, "error on extracting SBR namespaced-name from annotations")
		return
	}
	if common.IsSBRSelectorEmpty(sbrSelector) {
		return
	}

	q.Add(reconcile.Request{sbrSelector})
}
