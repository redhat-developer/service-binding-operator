package catchall

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type CatchAllReconciler struct {
	Client    client.Client
	DynClient dynamic.Interface
	Scheme    *runtime.Scheme
}

func (r *CatchAllReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	sbrSelector := getSelectorFromName(request.Name)
	_ = sbrSelector

	logf.Log.WithValues("Reconcile.Request", fmt.Sprintf("%#v", request)).Info("Debug -> Request!")

	// Do something with sbrSelector
	logf.Log.Info("Inside CatchAllReconciler.Reconcile. Will panic!")

	panic(fmt.Sprintf("%#v\n", request))
}
