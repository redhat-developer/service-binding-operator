package catchall

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// Shamelessly copied from https://github.com/isutton/eventing/blob/a4598653d46752d5ca8ef55f28d8b37cb1fd8e3f/pkg/adapter/apiserver/adapter.go#L144-L160

type unstructuredLister func(metav1.ListOptions) (*unstructured.UnstructuredList, error)

func asUnstructuredLister(ulist unstructuredLister) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		ul, err := ulist(opts)
		if err != nil {
			return nil, err
		}
		return ul, nil
	}
}

func asUnstructuredWatcher(wf cache.WatchFunc) cache.WatchFunc {
	return func(lo metav1.ListOptions) (watch.Interface, error) {
		return wf(lo)
	}
}
