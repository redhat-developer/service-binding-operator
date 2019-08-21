package catchall

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CatchAllReconciler struct {
	Client    client.Client
	DynClient dynamic.Interface
	Scheme    *runtime.Scheme
}

func (r *CatchAllReconciler) Reconcile(request UnstructuredRequest) (reconcile.Result, error) {
	panic("implement me")
}
