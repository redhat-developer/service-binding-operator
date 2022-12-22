/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package binding

import (
	ctx "context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/controllers"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/builder"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	controllers.BindingReconciler
}

// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings/finalizers,verbs=update

func validateLabels(fromSB, fromResource map[string]string) bool {
	fl := len(fromSB)
	l := 0
	for k, v := range fromResource {
		for m, n := range fromSB {
			fmt.Println(k, v, m, n)
			if k == m && v == n {
				l = l + 1
			}
		}
	}
	if fl == l {
		return true
	}
	return false
}

func New(clnt client.Client, log logr.Logger, scheme *runtime.Scheme) *ServiceBindingReconciler {
	r := &ServiceBindingReconciler{
		BindingReconciler: controllers.BindingReconciler{
			Client: clnt,
			Log:    log,
			Scheme: scheme,
			PipelineProvider: func(conf *rest.Config, lookup kubernetes.K8STypeLookup) (pipeline.Pipeline, error) {
				client, err := dynamic.NewForConfig(conf)
				if err != nil {
					return nil, err
				}
				authClient, err := authv1.NewForConfig(conf)
				if err != nil {
					return nil, err
				}
				return builder.DefaultBuilder.WithContextProvider(context.Provider(client, authClient.SubjectAccessReviews(), lookup)).Build(), nil
			},
			ReconcilingObject: func() apis.Object { return &v1alpha1.ServiceBinding{} },
		},
	}
	r.MapWorkloadToSB = func(a client.Object) []reconcile.Request {
		sbList := &v1alpha1.ServiceBindingList{}
		opts := &client.ListOptions{}
		if err := r.List(ctx.Background(), sbList, opts); err != nil {
			return []reconcile.Request{}
		}
		reply := make([]reconcile.Request, 0, len(sbList.Items))
		for _, sb := range sbList.Items {
			if sb.Spec.Application.Kind == a.GetObjectKind().GroupVersionKind().Kind &&
				validateLabels(sb.Spec.Application.LabelSelector.MatchLabels, a.GetLabels()) {
				reply = append(reply, reconcile.Request{NamespacedName: types.NamespacedName{
					Namespace: sb.Namespace,
					Name:      sb.Name,
				}})
			}
		}
		return reply
	}
	r.ResourceToWatch = func(ctx ctx.Context, key client.ObjectKey) (string, string, string) {
		sb := &v1alpha1.ServiceBinding{}
		err := r.Get(ctx, key, sb)
		if err != nil {
			return sb.Spec.Application.Group, sb.Spec.Application.Version, sb.Spec.Application.Kind
		}
		return "", "", ""
	}
	return r
}
