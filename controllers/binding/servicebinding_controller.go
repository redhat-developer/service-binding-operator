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
	"github.com/go-logr/logr"
	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/controllers"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/builder"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	controllers.BindingReconciler
}

// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings/finalizers,verbs=update

func New(client client.Client, log logr.Logger, scheme *runtime.Scheme) *ServiceBindingReconciler {
	r := &ServiceBindingReconciler{
		BindingReconciler: controllers.BindingReconciler{
			Client: client,
			Log:    log,
			Scheme: scheme,
			PipelineProvider: func(client dynamic.Interface, lookup context.K8STypeLookup) pipeline.Pipeline {
				return builder.DefaultBuilder.WithContextProvider(context.Provider(client, lookup)).Build()
			},
			ReconcilingObject: func() apis.Object { return &v1alpha1.ServiceBinding{} },
		},
	}

	return r
}
