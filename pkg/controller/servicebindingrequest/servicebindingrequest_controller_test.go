package servicebindingrequest

import (
	"testing"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestServiceBindingRequestController(t *testing.T) {
	var (
		backingOperatorName = "postgresql-operator.v0.1.0"
		name                = "postgres"
		namespace           = "default"
	)
	sbr := &v1alpha1.ServiceBindingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			BackingOperatorName: backingOperatorName,
			CSVNamespace:        namespace,
			ApplicationSelector: v1alpha1.ApplicationSelector{
				MatchLabels: map[string]string{
					"connects-to": "postgres",
					"environment": "production",
				},
			},
		},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, sbr)
	// Add CSV scheme
	if err := olmv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add CSV scheme: (%v)", err)
	}

	csv := &olmv1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backingOperatorName,
			Namespace: namespace,
		},
		Spec: olmv1alpha1.ClusterServiceVersionSpec{
			CustomResourceDefinitions: olmv1alpha1.CustomResourceDefinitions{
				Owned: []olmv1alpha1.CRDDescription{
					{
						Name: "some name",
						SpecDescriptors: []olmv1alpha1.SpecDescriptor{
							{
								XDescriptors: []string{"urn:alm:descriptor:servicebindingrequest:secret:password", "aaa:ccc:aa"},
							},
						},
					},
				},
			},
		},
	}

	s.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, csv)

	// Add Deployment scheme
	if err := appsv1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add Deployment scheme: (%v)", err)
	}

	dp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"connects-to": "postgres",
				"environment": "production",
			},
		},
	}

	s.AddKnownTypes(appsv1.SchemeGroupVersion, dp)

	// Objects to track in the fake client.
	objs := []runtime.Object{sbr, csv, dp}

	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileServiceBindingRequest object with the scheme and fake client.
	r := &ReconcileServiceBindingRequest{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}
