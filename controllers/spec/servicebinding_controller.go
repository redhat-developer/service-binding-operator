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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	specv1alpha2 "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha2"
)

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=service.binding,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=service.binding,resources=servicebindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=service.binding,resources=servicebindings/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods;services;endpoints;persistentvolumeclaims;events;configmaps;secrets,verbs=*
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=*
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ServiceBindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("servicebinding", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&specv1alpha2.ServiceBinding{}).
		Complete(r)
}
