package e2e

import (
	"context"
	"testing"
	"time"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/service-binding-operator/pkg/apis"
	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 120
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
	replicas             = 1
	operatorKind         = "ServiceBindingRequest"
	operatorAPIVersion   = "apps.openshift.io/v1alpha1"
)

// TestAddToScheme main method for end-to-end testing, will call out the other components in order
// to test service-binding-operator, after including custom-resource-definitions are part of the
// local scheme.
func TestAddToScheme(t *testing.T) {
	var err error

	t.Log("Adding ServiceBindingRequest scheme to cluster...")
	serviceBindingRequestList := &v1alpha1.ServiceBindingRequestList{
		TypeMeta: metav1.TypeMeta{
			Kind:       operatorKind,
			APIVersion: operatorAPIVersion,
		},
	}

	if err = framework.AddToFrameworkScheme(apis.AddToScheme, serviceBindingRequestList); err != nil {
		t.Fatalf("Error on adding ServiceBindingRequest CRD to cluster!")
	}

	t.Log("Adding ClusterServiceVersion scheme to cluster...")
	clusterServiceVersionListObj := &olmv1alpha1.ClusterServiceVersionList{}

	if err = framework.AddToFrameworkScheme(apis.AddToScheme, clusterServiceVersionListObj); err != nil {
		t.Fatalf("Error on adding ServiceBindingRequest CRD to cluster!")
	}

	t.Run("end-to-end", func(t *testing.T) {
		t.Run("scenario-1", ServiceBindingRequest)
	})
}

func cleanUpOptions(ctx *framework.TestCtx) *framework.CleanupOptions {
	return &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: time.Duration(time.Second * retryInterval),
	}
}

func mockedObjects(t *testing.T, ns string, f *framework.Framework, ctx *framework.TestCtx) {
	var err error

	t.Log("Starting end-to-end tests for operator...")

	crdName := "e2e-resource-name"
	crdVersion := "0.0.1"
	secretName := "e2e-secret"

	/*
		clusterServiceVersionObj := olmv1alpha1.ClusterServiceVersion{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterServiceVersion",
				APIVersion: "operators.coreos.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "e2e-cluster-service-version",
				Namespace: ns,
			},
			Spec: olmv1alpha1.ClusterServiceVersionSpec{
				CustomResourceDefinitions: olmv1alpha1.CustomResourceDefinitions{
					Owned: []olmv1alpha1.CRDDescription{{
						Name:    crdName,
						Version: crdVersion,
						SpecDescriptors: []olmv1alpha1.SpecDescriptor{{
							DisplayName:  secretName,
							Path:         secretName,
							XDescriptors: []string{"urn:alm:descriptor:io.kubernetes:Secret"},
						}},
					}},
				},
			},
		}

		clusterServiceVersionListObj := olmv1alpha1.ClusterServiceVersionList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterServiceVersionList",
				APIVersion: "operators.coreos.com/v1alpha1",
			},
			Items: []olmv1alpha1.ClusterServiceVersion{clusterServiceVersionObj},
		}

		if err = f.Client.Create(context.TODO(), &clusterServiceVersionListObj, cleanUpOptions(ctx)); err != nil {
			t.Fatalf("Error on creating CSV list object: '%s'", err)
		}
	*/

	secretObj := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: ns,
		},
		Data: map[string][]byte{"secret-entry": []byte("secret-value")},
	}

	if err = f.Client.Create(context.TODO(), &secretObj, cleanUpOptions(ctx)); err != nil {
		t.Fatalf("Error on creating secret object: '%s'", err)
	}

	serviceBindingRequestObj := v1alpha1.ServiceBindingRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       operatorKind,
			APIVersion: operatorAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-bind-request",
			Namespace: ns,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			BackingSelector: v1alpha1.BackingSelector{
				ResourceName:    crdName,
				ResourceVersion: crdVersion,
			},
		},
	}

	if err = f.Client.Create(context.TODO(), &serviceBindingRequestObj, cleanUpOptions(ctx)); err != nil {
		t.Fatalf("Error on creating service-binding-request object: '%s'", err)
	}
}

// ServiceBindingRequest bootstrap method to initialize cluster resources and setup a testing
// namespace, after bootstrap operator related tests method is called out.
func ServiceBindingRequest(t *testing.T) {
	t.Log("Creating a new test context...")
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	t.Log("Initializing cluster resources...")

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: time.Duration(time.Second * retryInterval),
	})
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			t.Fatalf("Failed to setup cluster resources: '%s'", err)
		}
	}

	// namespace name is informed on command-line or defined dinamically
	ns, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Error on acquiring a test namespace: '%s'", err)
	}

	t.Logf("Using namespace '%s' for testing...", ns)

	f := framework.Global
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, ns, "service-binding-operator", replicas, retryInterval, timeout)
	if err != nil {
		t.Fatalf("Error on waiting for operator deployment: '%s'", err)
	}

	serviceBindingRequestTest(t, ns, f, ctx)
}

func serviceBindingRequestTest(t *testing.T, ns string, f *framework.Framework, ctx *framework.TestCtx) {
	mockedObjects(t, ns, f, ctx)
}
