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
	"context"
	"fmt"
	"sync"

	"github.com/redhat-developer/service-binding-operator/pkg/binding/registry"

	"github.com/go-logr/logr"
	bindingapi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context/service"
	"github.com/redhat-developer/service-binding-operator/pkg/util"
	v1apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CrdReconciler reconciles a CustomResourceDefinition resources
type CrdReconciler struct {
	client.Client
	serviceBuilder     service.Builder
	Log                logr.Logger
	Scheme             *runtime.Scheme
	bindableKinds      *sync.Map
	annotationRegistry registry.Registry
}

// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=bindablekinds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=bindablekinds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=binding.operators.coreos.com,resources=bindablekinds/finalizers,verbs=update
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CrdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcileResult ctrl.Result, reconcileError error) {
	defer func() {
		if err := recover(); err != nil {
			reconcileResult = ctrl.Result{}
			reconcileError = fmt.Errorf("panic occurred: %v", err)
		}
	}()
	log := r.Log.WithValues("CRD", req.NamespacedName)
	crd := &v1apiextensions.CustomResourceDefinition{}
	err := r.Get(ctx, req.NamespacedName, crd)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("CRD resource not found. Ignoring since object must be deleted", "err", err)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get CRD")
		return ctrl.Result{}, err
	}

	toPersist := false

	for i := range crd.Spec.Versions {
		if !crd.Spec.Versions[i].Served {
			continue
		}
		gvk := schema.GroupVersionKind{Group: crd.Spec.Group, Kind: crd.Spec.Names.Kind, Version: crd.Spec.Versions[i].Name}
		if !crd.GetDeletionTimestamp().IsZero() {
			r.bindableKinds.Delete(gvk)
			toPersist = true
			continue
		}
		fakeServiceContent := &unstructured.Unstructured{}
		fakeServiceContent.SetName("s1")
		fakeServiceContent.SetGroupVersionKind(gvk)
		service, err := r.serviceBuilder.Build(fakeServiceContent, service.CrdReaderOption(func(gvk *schema.GroupVersionResource) (*unstructured.Unstructured, error) {
			return converter.ToUnstructured(crd)
		}))
		if err != nil {
			return ctrl.Result{}, err
		}
		bindable, err := service.IsBindable()
		if err != nil {
			return ctrl.Result{}, err
		}
		if bindable {
			r.bindableKinds.Store(gvk, true)
			toPersist = true
			log.Info("bindable", "gvk", gvk)
		} else {
			annotations, found := r.annotationRegistry.GetAnnotations(gvk)
			if found {
				log.Info("Found bindable annotations", "gvk", gvk, "annotations", annotations)
				crd.SetAnnotations(util.MergeMaps(crd.GetAnnotations(), annotations))
				err := r.Update(ctx, crd)
				if err != nil {
					log.Error(err, "Error updating CRD")
					return ctrl.Result{}, err
				}
				log.Info("Annotations applied")
			}
		}
	}

	if !toPersist {
		log.Info("Done")
		return ctrl.Result{}, nil
	}
	bk := &bindingapi.BindableKinds{}
	err = r.Get(ctx, client.ObjectKey{Name: "bindable-kinds"}, bk)
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "Failed to get bindable kinds", "err", err)
			return ctrl.Result{}, err
		}
	}
	bk.Status = make([]bindingapi.BindableKindsStatus, 0)
	r.bindableKinds.Range(func(key, value interface{}) bool {
		gvk, ok := key.(schema.GroupVersionKind)
		if ok {
			bk.Status = append(bk.Status, bindingapi.BindableKindsStatus{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})
		}
		return true
	})

	if bk.UID == "" {
		bk.Name = "bindable-kinds"
		if err := r.Create(ctx, bk); err != nil {
			log.Error(err, "on create")
			return ctrl.Result{}, err
		}
		log.Info("created bindable kinds")
	} else {
		if err := r.Update(ctx, bk); err != nil {
			log.Error(err, "on update")
			return ctrl.Result{}, err
		}
		log.Info("updated bindable kinds")
	}

	log.Info("Done")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CrdReconciler) SetupWithManager(mgr ctrl.Manager, bindableKinds *sync.Map, olm bool) error {
	r.bindableKinds = bindableKinds
	r.annotationRegistry = registry.ServiceAnnotations
	dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	r.serviceBuilder = service.NewBuilder(kubernetes.ResourceLookup(mgr.GetRESTMapper())).WithClient(dynamicClient)
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1apiextensions.CustomResourceDefinition{}).
		Complete(r)
}
