package servicebindingrequest

import (
	"context"
	"testing"

	osappsv1 "github.com/openshift/api/apps/v1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

func TestServiceBindingRequestController(t *testing.T) {
	var (
		backingOperatorName = "postgresql-operator.v0.1.0"
		name                = "postgres"
		deploymentName      = "my-app"
		namespace           = "default"
	)
	sbr := &v1alpha1.ServiceBindingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			BackingSelector: v1alpha1.BackingSelector{
				ResourceName: "specialdb.example.org",
			},
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
						Name: "specialdb.example.org",
						SpecDescriptors: []olmv1alpha1.SpecDescriptor{
							{
								XDescriptors: []string{"urn:alm:descriptor:servicebindingrequest:secret:password", "aaa:ccc:aa"},
							},
						},
					},
					{
						Name: "specialdb.example.org",
						SpecDescriptors: []olmv1alpha1.SpecDescriptor{
							{
								XDescriptors: []string{"urn:alm:descriptor:servicebindingrequest:configmap:username", "aaa:ccc:aa"},
							},
						},
					},
				},
			},
		},
	}

	s.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, csv)

	t.Run("Deployment", func(t *testing.T) {

		// Add Deployment scheme
		if err := appsv1.AddToScheme(s); err != nil {
			t.Fatalf("Unable to add Deployment scheme: (%v)", err)
		}

		dp := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: namespace,
				Labels: map[string]string{
					"connects-to": "postgres",
					"environment": "production",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app"},
						},
					},
				},
			},
		}

		s.AddKnownTypes(appsv1.SchemeGroupVersion, dp)

		// Objects to track in the fake client.
		objs := []runtime.Object{sbr, csv, dp}

		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileServiceBindingRequest object with the scheme and fake client.
		r := &Reconciler{client: cl, scheme: s}

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

		dpOut := &appsv1.Deployment{}
		nn := types.NamespacedName{
			Name:      deploymentName,
			Namespace: namespace,
		}
		err = r.client.Get(context.TODO(), nn, dpOut)
		if err != nil {
			t.Fatalf("get deployment: (%v)", err)
		}
		n := dpOut.Spec.Template.Spec.Containers[0].Env[0].Name
		if n != "POSTGRES_PASSWORD" {
			t.Errorf("Environment name not matching: %s", n)
		}
		n2 := dpOut.Spec.Template.Spec.Containers[0].Env[1].Name
		if n2 != "POSTGRES_USERNAME" {
			t.Errorf("Environment name not matching: %s", n2)
		}
	})

	t.Run("DeploymentConfig", func(t *testing.T) {

		// Add DeploymentConfig scheme
		if err := osappsv1.AddToScheme(s); err != nil {
			t.Fatalf("Unable to add DeploymentConfig scheme: (%v)", err)
		}

		dp := &osappsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: namespace,
				Labels: map[string]string{
					"connects-to": "postgres",
					"environment": "production",
				},
			},
			Spec: osappsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app"},
						},
					},
				},
			},
		}

		s.AddKnownTypes(osappsv1.SchemeGroupVersion, dp)

		sbr.Spec.ApplicationSelector.ResourceKind = "DeploymentConfig"
		// Objects to track in the fake client.
		objs := []runtime.Object{sbr, csv, dp}

		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileServiceBindingRequest object with the scheme and fake client.
		r := &Reconciler{client: cl, scheme: s}

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

		dpOut := &osappsv1.DeploymentConfig{}
		nn := types.NamespacedName{
			Name:      deploymentName,
			Namespace: namespace,
		}
		err = r.client.Get(context.TODO(), nn, dpOut)
		if err != nil {
			t.Fatalf("get deployment: (%v)", err)
		}
		n := dpOut.Spec.Template.Spec.Containers[0].Env[0].Name
		if n != "POSTGRES_PASSWORD" {
			t.Errorf("Environment name not matching: %s", n)
		}
		n2 := dpOut.Spec.Template.Spec.Containers[0].Env[1].Name
		if n2 != "POSTGRES_USERNAME" {
			t.Errorf("Environment name not matching: %s", n2)
		}
	})

	t.Run("StatefulSet", func(t *testing.T) {

		// Add StatefulSet scheme
		if err := appsv1.AddToScheme(s); err != nil {
			t.Fatalf("Unable to add StatefulSet scheme: (%v)", err)
		}

		dp := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: namespace,
				Labels: map[string]string{
					"connects-to": "postgres",
					"environment": "production",
				},
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app"},
						},
					},
				},
			},
		}

		s.AddKnownTypes(appsv1.SchemeGroupVersion, dp)

		sbr.Spec.ApplicationSelector.ResourceKind = "StatefulSet"
		// Objects to track in the fake client.
		objs := []runtime.Object{sbr, csv, dp}

		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileServiceBindingRequest object with the scheme and fake client.
		r := &Reconciler{client: cl, scheme: s}

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

		dpOut := &appsv1.StatefulSet{}
		nn := types.NamespacedName{
			Name:      deploymentName,
			Namespace: namespace,
		}
		err = r.client.Get(context.TODO(), nn, dpOut)
		if err != nil {
			t.Fatalf("get deployment: (%v)", err)
		}
		n := dpOut.Spec.Template.Spec.Containers[0].Env[0].Name
		if n != "POSTGRES_PASSWORD" {
			t.Errorf("Environment name not matching: %s", n)
		}
		n2 := dpOut.Spec.Template.Spec.Containers[0].Env[1].Name
		if n2 != "POSTGRES_USERNAME" {
			t.Errorf("Environment name not matching: %s", n2)
		}
	})

	t.Run("DaemonSet", func(t *testing.T) {

		// Add DaemonSet scheme
		if err := appsv1.AddToScheme(s); err != nil {
			t.Fatalf("Unable to add DaemonSet scheme: (%v)", err)
		}

		dp := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: namespace,
				Labels: map[string]string{
					"connects-to": "postgres",
					"environment": "production",
				},
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "app"},
						},
					},
				},
			},
		}

		s.AddKnownTypes(appsv1.SchemeGroupVersion, dp)

		sbr.Spec.ApplicationSelector.ResourceKind = "DaemonSet"
		// Objects to track in the fake client.
		objs := []runtime.Object{sbr, csv, dp}

		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileServiceBindingRequest object with the scheme and fake client.
		r := &Reconciler{client: cl, scheme: s}

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

		dpOut := &appsv1.DaemonSet{}
		nn := types.NamespacedName{
			Name:      deploymentName,
			Namespace: namespace,
		}
		err = r.client.Get(context.TODO(), nn, dpOut)
		if err != nil {
			t.Fatalf("get deployment: (%v)", err)
		}
		n := dpOut.Spec.Template.Spec.Containers[0].Env[0].Name
		if n != "POSTGRES_PASSWORD" {
			t.Errorf("Environment name not matching: %s", n)
		}
		n2 := dpOut.Spec.Template.Spec.Containers[0].Env[1].Name
		if n2 != "POSTGRES_USERNAME" {
			t.Errorf("Environment name not matching: %s", n2)
		}
	})

}

func TestCornerCasesServiceBindingRequestController(t *testing.T) {
	var (
		backingOperatorName = "postgresql-operator.v0.1.0"
		name                = "postgres"
		deploymentName      = "my-app"
		namespace           = "default"
	)
	sbr := &v1alpha1.ServiceBindingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			BackingSelector: v1alpha1.BackingSelector{
				ResourceName:    "specialdb.example.org",
				ResourceVersion: "v1alpha1",
			},
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

	t.Run("object not added to client", func(t *testing.T) {
		// Objects to track in the fake client.
		objs := []runtime.Object{}

		cl := fake.NewFakeClient(objs...)

		// Create a ReconcileServiceBindingRequest object with the scheme and fake client.
		r := &Reconciler{client: cl, scheme: s}

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
		if res.Requeue {
			t.Error("reconcile did not requeue request as expected")
		}
	})
	t.Run("match version", func(t *testing.T) {
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
							Name: "specialdb.example.org",
							SpecDescriptors: []olmv1alpha1.SpecDescriptor{
								{
									XDescriptors: []string{"urn:alm:descriptor:servicebindingrequest:secret:password", "aaa:ccc:aa"},
								},
							},
						},
						{
							Name: "specialdb.example.org",
							SpecDescriptors: []olmv1alpha1.SpecDescriptor{
								{
									XDescriptors: []string{"urn:alm:descriptor:servicebindingrequest:configmap:username", "aaa:ccc:aa"},
								},
							},
						},
					},
				},
			},
		}

		s.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, csv)
		t.Run("Deployment", func(t *testing.T) {

			// Add Deployment scheme
			if err := appsv1.AddToScheme(s); err != nil {
				t.Fatalf("Unable to add Deployment scheme: (%v)", err)
			}

			dp := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deploymentName,
					Namespace: namespace,
					Labels: map[string]string{
						"connects-to": "postgres",
						"environment": "production",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{Name: "app"},
							},
						},
					},
				},
			}

			s.AddKnownTypes(appsv1.SchemeGroupVersion, dp)

			// Objects to track in the fake client.
			objs := []runtime.Object{sbr, csv, dp}

			cl := fake.NewFakeClient(objs...)

			// Create a ReconcileServiceBindingRequest object with the scheme and fake client.
			r := &Reconciler{client: cl, scheme: s}

			// Mock request to simulate Reconcile() being called on an event for a
			// watched resource .
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      name,
					Namespace: namespace,
				},
			}

			_, err := r.Reconcile(req)
			if err == nil {
				t.Fatalf("reconcile worked without version match")
			}

		})

	})
}
