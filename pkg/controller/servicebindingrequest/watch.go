package servicebindingrequest

import (
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CSVToWatcherMapper struct {
	c *SBRController
}

func (w *CSVToWatcherMapper) Map(obj handler.MapObject) []reconcile.Request {
	// FIXME: Use informer lister instead of client here somehow.
	u, err := w.c.Client.
		Resource(olmv1alpha1.SchemeGroupVersion.WithResource("clusterserviceversions")).
		Namespace(obj.Meta.GetNamespace()).
		Get(obj.Meta.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Failed to get ClusterServiceVersion")
	}

	csv := &olmv1alpha1.ClusterServiceVersion{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, csv)
	if err != nil {
		log.Error(err, "Failed to convert ClusterServiceVersion object")
	}

	for _, crdDescription := range csv.Spec.CustomResourceDefinitions.Owned {
		_, gv := schema.ParseResourceArg(crdDescription.Name)

		gvk := schema.GroupVersionKind{
			Group:   gv.Group,
			Version: crdDescription.Version,
			Kind:    crdDescription.Kind,
		}

		err := w.c.AddWatchForGVK(gvk)
		if err != nil {
			log.WithValues("GroupVersionKind", gvk).Error(err, "Failed to create a watch")
		}
	}

	return []reconcile.Request{}
}

func NewCreateWatchEventHandler() handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{ToRequests: &CSVToWatcherMapper{}}
}
