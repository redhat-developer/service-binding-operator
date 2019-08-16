package servicebindingrequest

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RequeueOnNotFound inspect error, if not-found then returns Requeue, otherwise expose the error.
func RequeueOnNotFound(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		return Requeue()
	}
	return reconcile.Result{Requeue: true}, err
}

// RequeueError simply requeue exposing the error.
func RequeueError(err error) (reconcile.Result, error) {
	return reconcile.Result{Requeue: true}, err
}

// Requeue based on empty result and no error informed upstream, request will be requeued.
func Requeue() (reconcile.Result, error) {
	return reconcile.Result{Requeue: true}, nil
}

// Done when no error is informed and request is not set for requeue.
func Done() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
