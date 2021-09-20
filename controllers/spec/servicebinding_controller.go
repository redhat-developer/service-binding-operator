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

package spec

import (
	"github.com/go-logr/logr"
	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
	"github.com/redhat-developer/service-binding-operator/controllers"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/builder"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	controllers.BindingReconciler
}

// +kubebuilder:rbac:groups=servicebinding.io,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=servicebinding.io,resources=servicebindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=servicebinding.io,resources=servicebindings/finalizers,verbs=update

func New(client client.Client, log logr.Logger, scheme *runtime.Scheme) *ServiceBindingReconciler {
	r := &ServiceBindingReconciler{
		BindingReconciler: controllers.BindingReconciler{
			Client: client,
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
				return builder.SpecBuilder.WithContextProvider(context.SpecProvider(client, authClient.SubjectAccessReviews(), lookup)).Build(), nil
			},
			ReconcilingObject: func() apis.Object { return &v1alpha3.ServiceBinding{} },
		},
	}

	return r
}
