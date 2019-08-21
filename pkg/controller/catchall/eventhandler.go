package catchall

import (
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

// TODO: Better name for this request
type UnstructuredRequest struct {
	reconcile.Request
	ServiceBindingRequestSelector int
}

func getServiceBindingRequestSelector(_ runtime.Object) (int, error) {
	// TODO: Should figure out data structure to keep for the worker.
	panic("implement me")
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

	q.Add(UnstructuredRequest{
		ServiceBindingRequestSelector: sbrSelector,
		Request: reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      evt.Meta.GetName(),
				Namespace: evt.Meta.GetNamespace(),
			},
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
		q.Add(UnstructuredRequest{
			ServiceBindingRequestSelector: sbrSelector,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      evt.MetaOld.GetName(),
					Namespace: evt.MetaOld.GetNamespace(),
				},
			},
		})
	} else {
		enqueueLog.Error(nil, "UpdateEvent received with no old metadata", "event", evt)
	}

	if evt.MetaNew != nil {
		q.Add(UnstructuredRequest{
			ServiceBindingRequestSelector: sbrSelector,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      evt.MetaNew.GetName(),
					Namespace: evt.MetaNew.GetNamespace(),
				},
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

	q.Add(UnstructuredRequest{
		ServiceBindingRequestSelector: sbrSelector,
		Request: reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      evt.Meta.GetName(),
				Namespace: evt.Meta.GetNamespace(),
			},
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

	q.Add(UnstructuredRequest{
		ServiceBindingRequestSelector: sbrSelector,
		Request: reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      evt.Meta.GetName(),
				Namespace: evt.Meta.GetNamespace(),
			},
		},
	})
}
