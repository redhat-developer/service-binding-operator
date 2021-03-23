package controllers

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func TestSBRControllerBuildSBRPredicate(t *testing.T) {
	// keep the predicate around
	pred := buildSBRPredicate(log.NewLog("test-log"))

	// the expected behavior is that every create event triggers a reconciliation
	t.Run("create", func(t *testing.T) {
		if got := pred.Create(event.CreateEvent{}); !got {
			t.Errorf("newSBRPredicate() = %v, want %v", got, true)
		}
	})

	// update exercises changes that should or not trigger the reconciliation
	t.Run("update", func(t *testing.T) {
		sbrA := &v1alpha1.ServiceBinding{
			Spec: v1alpha1.ServiceBindingSpec{
				Services: []v1alpha1.Service{
					{
						NamespacedRef: v1alpha1.NamespacedRef{
							Ref: v1alpha1.Ref{
								Group: "test", Version: "v1alpha1", Kind: "TestHost", Name: "",
							},
						},
					},
				},
			},
		}
		sbrB := &v1alpha1.ServiceBinding{
			Spec: v1alpha1.ServiceBindingSpec{
				Services: []v1alpha1.Service{
					{
						NamespacedRef: v1alpha1.NamespacedRef{
							Ref: v1alpha1.Ref{
								Group: "test", Version: "v1", Kind: "TestHost",
							},
						},
					},
				},
			},
		}

		tests := []struct {
			name string
			want bool
			a    runtime.Object
			b    runtime.Object
		}{
			{name: "same-spec", want: false, a: sbrA, b: sbrA},
			{name: "changed-spec", want: true, a: sbrA, b: sbrB},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := pred.Update(event.UpdateEvent{ObjectOld: tt.a, ObjectNew: tt.b}); got != tt.want {
					t.Errorf("newSBRPredicate() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	// delete verifies that SBRs will be reconciled prior to its deletion
	t.Run("delete", func(t *testing.T) {
		tests := []struct {
			name           string
			want           bool
			confirmDeleted bool
		}{
			// FIXME: validate whether this is the behavior we want
			{name: "delete-not-confirmed", confirmDeleted: false, want: true},
			{name: "delete-confirmed", confirmDeleted: true, want: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := pred.Delete(event.DeleteEvent{DeleteStateUnknown: tt.confirmDeleted}); got != tt.want {
					t.Errorf("newSBRPredicate() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}

type fakeController struct {
	watchCallback func(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error
}

var _ controller.Controller = (*fakeController)(nil)

func (f *fakeController) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (f *fakeController) Start(stop <-chan struct{}) error {
	return nil
}

func (f *fakeController) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	if f.watchCallback != nil {
		return f.watchCallback(src, eventhandler, predicates...)
	}
	return nil
}
