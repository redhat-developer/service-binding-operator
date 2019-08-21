package catchall

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// TODO: Use the same log infrastructure as others
var enqueueLog = logf.KBLog.WithName("eventhandler").WithName("EnqueueRequestForObject")

type EnqueueRequestForUnstructured struct{}

func getServiceBindingRequestSelector(_ runtime.Object) (int, error) {
	// TODO: Should figure out data structure to keep for the worker.
	panic("implement me")
}

func composeNameWithSelector(name string, sbrSelector int) string {
	return fmt.Sprintf("%s!%v", name, sbrSelector)
}

func getSelectorFromName(name string) int {
	return 1
}

func (e EnqueueRequestForUnstructured) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	if evt.Meta == nil {
		enqueueLog.Error(nil, "CreateEvent received with no metadata", "event", evt)
		return
	}

	sbrSelector, err := getServiceBindingRequestSelector(evt.Object)
	if err != nil {
		enqueueLog.Error(err, "change me")
		return
	}

	q.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      composeNameWithSelector(evt.Meta.GetName(), sbrSelector),
			Namespace: evt.Meta.GetNamespace(),
		},
	})
}

func (e EnqueueRequestForUnstructured) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := getServiceBindingRequestSelector(evt.ObjectNew)
	if err != nil {
		enqueueLog.Error(err, "change me")
		return
	}

	if evt.MetaOld != nil {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      composeNameWithSelector(evt.MetaOld.GetName(), sbrSelector),
				Namespace: evt.MetaOld.GetNamespace(),
			},
		})
	} else {
		enqueueLog.Error(nil, "UpdateEvent received with no old metadata", "event", evt)
	}

	if evt.MetaNew != nil {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      composeNameWithSelector(evt.MetaNew.GetName(), sbrSelector),
				Namespace: evt.MetaNew.GetNamespace(),
			},
		})
	} else {
		enqueueLog.Error(nil, "UpdateEvent received with no new metadata", "event", evt)
	}
}

func (e EnqueueRequestForUnstructured) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := getServiceBindingRequestSelector(evt.Object)
	if err != nil {
		enqueueLog.Error(err, "change me")
		return
	}

	if evt.Meta == nil {
		enqueueLog.Error(nil, "DeleteEvent received with no metadata", "event", evt)
		return
	}

	q.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      composeNameWithSelector(evt.Meta.GetName(), sbrSelector),
			Namespace: evt.Meta.GetNamespace(),
		},
	})
}

func (e EnqueueRequestForUnstructured) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	sbrSelector, err := getServiceBindingRequestSelector(evt.Object)
	if err != nil {
		enqueueLog.Error(err, "change me")
		return
	}

	if evt.Meta == nil {
		enqueueLog.Error(nil, "GenericEvent received with no metadata", "event", evt)
		return
	}

	q.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      composeNameWithSelector(evt.Meta.GetName(), sbrSelector),
			Namespace: evt.Meta.GetNamespace(),
		},
	})
}
