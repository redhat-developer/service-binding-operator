package servicebinding

import (
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// serviceBindingRequestResource the name of ServiceBinding resource.
	serviceBindingRequestResource = "servicebindings"
	// serviceBindingRequestKind defines the name of the CRD kind.
	serviceBindingRequestKind = "ServiceBinding"
	// clusterServiceVersionKind the name of ClusterServiceVersion kind.
	clusterServiceVersionKind = "ClusterServiceVersion"
	// secretResource defines the resource name for Secrets.
	secretResource = "secrets"
	// secretKind defines the name of Secret kind.
	secretKind = "Secret"
)

// requeueOnNotFound inspect error, if not-found then returns Requeue, otherwise expose the error.
func requeueOnNotFound(err error, requeueAfter int64) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		return requeue(nil, requeueAfter)
	}
	return noRequeue(err)
}

// requeueOnConflict in case of conflict error, returning the error with requeue, otherwise Done.
func requeueOnConflict(err error) (reconcile.Result, error) {
	if errors.IsConflict(err) {
		return requeueError(err)
	}
	return done()
}

// requeueError simply requeue exposing the error.
func requeueError(err error) (reconcile.Result, error) {
	return reconcile.Result{Requeue: true}, err
}

// requeue based on empty result and no error informed upstream, request will be requeued.
func requeue(err error, requeueAfter int64) (reconcile.Result, error) {
	return reconcile.Result{
		RequeueAfter: time.Duration(requeueAfter) * time.Second,
		Requeue:      true,
	}, err
}

// done when no error is informed and request is not set for requeue.
func done() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// doneOnNotFound will return done when error is not-found, otherwise it calls out NoRequeue.
func doneOnNotFound(err error) (reconcile.Result, error) {
	if errors.IsNotFound(err) {
		return done()
	}
	return noRequeue(err)
}

// noRequeue returns error without requeue flag.
func noRequeue(err error) (reconcile.Result, error) {
	return reconcile.Result{}, err
}

// containsStringSlice given a string slice and a string, returns boolean when is contained.
func containsStringSlice(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// removeStringSlice given a string slice and a string, returns a new slice without given string.
func removeStringSlice(slice []string, str string) []string {
	var cleanSlice []string
	for _, s := range slice {
		if str != s {
			cleanSlice = append(cleanSlice, s)
		}
	}
	return cleanSlice
}
