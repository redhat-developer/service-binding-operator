package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
)

const finalizerName = "finalizer.servicebinding.openshift.io"

func MaybeAddFinalizer(obj Object) bool {
	finalizers := obj.GetFinalizers()
	for _, f := range finalizers {
		if f == finalizerName {
			return false
		}
	}
	obj.SetFinalizers(append(finalizers, finalizerName))
	return true
}

func MaybeRemoveFinalizer(obj Object) bool {
	finalizers := obj.GetFinalizers()
	for i, f := range finalizers {
		if f == finalizerName {
			obj.SetFinalizers(append(finalizers[:i], finalizers[i+1:]...))
			return true
		}
	}
	return false
}

type Object interface {
	runtime.Object
	GetFinalizers() []string
	SetFinalizers([]string)
	HasDeletionTimestamp() bool
}
