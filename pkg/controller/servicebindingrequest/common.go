package servicebindingrequest

import (
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ServiceBindingRequestResource the name of ServiceBindingRequest resource.
const ServiceBindingRequestResource = "servicebindingrequests"

// ServiceBindingRequestKind defines the name of the CRD kind.
const ServiceBindingRequestKind = "ServiceBindingRequest"

// DeploymentConfigKind defines the name of DeploymentConfig kind.
const DeploymentConfigKind = "DeploymentConfig"

// ClusterServiceVersionKind the name of ClusterServiceVersion kind.
const ClusterServiceVersionKind = "ClusterServiceVersion"

// RequeueOnNotFound inspect error, if not-found then returns Requeue, otherwise expose the error.
func RequeueOnNotFound(err error, requeueAfter int64) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		return Requeue(nil, requeueAfter)
	}
	return Requeue(err, requeueAfter)
}

// RequeueOnConflict in case of conflict error, returning the error with requeue, otherwise Done.
func RequeueOnConflict(err error) (reconcile.Result, error) {
	if errors.IsConflict(err) {
		return RequeueError(err)
	}
	return Done()
}

// RequeueError simply requeue exposing the error.
func RequeueError(err error) (reconcile.Result, error) {
	return reconcile.Result{Requeue: true}, err
}

// Requeue based on empty result and no error informed upstream, request will be requeued.
func Requeue(err error, requeueAfter int64) (reconcile.Result, error) {
	return reconcile.Result{
		RequeueAfter: time.Duration(requeueAfter) * time.Second,
		Requeue:      true,
	}, err
}

// Done when no error is informed and request is not set for requeue.
func Done() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func DoneOnNotFound(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		return Done()
	}
	return NoRequeue(err)
}

func NoRequeue(err error) (reconcile.Result, error) {
	return reconcile.Result{}, err
}

func removeStringSlice(slice []string, str string) []string {
	var cleanSlice []string
	for _, s := range slice {
		if str != s {
			cleanSlice = append(cleanSlice, s)
		}
	}
	return cleanSlice
}
