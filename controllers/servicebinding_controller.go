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

package controllers

import (
	"github.com/redhat-developer/service-binding-operator/pkg/log"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// reconcilerLog local logger instance
var reconcilerLog = log.NewLog("reconciler")

// ServiceBindingReconciler reconciles a ServiceBinding object
type ServiceBindingReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	dynClient       dynamic.Interface // kubernetes dynamic api client
	resourceWatcher ResourceWatcher   // ResourceWatcher to add watching for specific GVK/GVR
	restMapper      meta.RESTMapper
}

// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=servicebindings/finalizers,verbs=update
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
	return r.doReconcile(req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	client, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	opts := controller.Options{Reconciler: r, MaxConcurrentReconciles: maxConcurrentReconciles}
	c, err := controller.New(controllerName, mgr, opts)
	if err != nil {
		return err
	}

	r.dynClient = client
	r.restMapper = mgr.GetRESTMapper()
	sbr := &sbrController{
		Controller:   c,
		Client:       client,
		typeLookup:   r,
		watchingGVKs: make(map[schema.GroupVersionKind]bool),
		logger:       log.NewLog("sbrcontroller"),
	}
	r.resourceWatcher = sbr
	return sbr.Watch()
}
