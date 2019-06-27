package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	olminstall "github.com/operator-framework/operator-lifecycle-manager/pkg/controller/install"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
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

// TestAddSchemesToFramework starting point of the test, it declare the CRDs that will be using
// during end-to-end tests.
func TestAddSchemesToFramework(t *testing.T) {
	serviceBindingRequestList := &v1alpha1.ServiceBindingRequestList{
		Items: []v1alpha1.ServiceBindingRequest{v1alpha1.ServiceBindingRequest{}},
	}

	t.Log("Adding ServiceBindingRequest scheme to cluster...")
	err := framework.AddToFrameworkScheme(apis.AddToScheme, serviceBindingRequestList)
	assert.Nil(t, err)

	clusterServiceVersionListObj := &olmv1alpha1.ClusterServiceVersionList{
		Items: []olmv1alpha1.ClusterServiceVersion{olmv1alpha1.ClusterServiceVersion{}},
	}

	t.Log("Adding ClusterServiceVersion scheme to cluster...")
	err = framework.AddToFrameworkScheme(olmv1alpha1.AddToScheme, clusterServiceVersionListObj)
	assert.Nil(t, err)

	t.Run("end-to-end", func(t *testing.T) {
		t.Run("scenario-1", ServiceBindingRequest)
	})
}

// cleanUpOptions using global variables to create the object.
func cleanUpOptions(ctx *framework.TestCtx) *framework.CleanupOptions {
	return &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: time.Duration(time.Second * retryInterval),
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
	assert.Nil(t, err)

	t.Logf("Using namespace '%s' for testing...", ns)
	f := framework.Global
	err = e2eutil.WaitForOperatorDeployment(
		t, f.KubeClient, ns, "service-binding-operator", replicas, retryInterval, timeout)
	assert.Nil(t, err)

	// populating cluster with mocked CRDs
	mockedObjects(t, ns, f, ctx)
	// executing testing steps on operator
	serviceBindingRequestTest(t, ns, f, ctx)
}

// mockedObjects creates all required CRDs in the cluster.
func mockedObjects(t *testing.T, ns string, f *framework.Framework, ctx *framework.TestCtx) {
	crdName := "e2e-resource-name"
	crdVersion := "0.0.1"
	secretName := "e2e-secret"

	strategy := olminstall.StrategyDetailsDeployment{
		DeploymentSpecs: []olminstall.StrategyDeploymentSpec{{
			Name: "deployment",
			Spec: appsv1.DeploymentSpec{},
		}},
	}

	strategyJSON, err := json.Marshal(strategy)
	assert.Nil(t, err)

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
			DisplayName: "e2e csv",
			InstallStrategy: olmv1alpha1.NamedInstallStrategy{
				StrategyName:    "deployment",
				StrategySpecRaw: strategyJSON,
			},
			CustomResourceDefinitions: olmv1alpha1.CustomResourceDefinitions{
				Owned: []olmv1alpha1.CRDDescription{{
					DisplayName: crdName,
					Name:        crdName,
					Version:     crdVersion,
					Description: "e2e csv example",
					SpecDescriptors: []olmv1alpha1.SpecDescriptor{{
						DisplayName:  secretName,
						Description:  "e2e csv example secret",
						Path:         secretName,
						XDescriptors: []string{"urn:alm:descriptor:io.kubernetes:Secret"},
					}},
				}},
			},
		},
	}

	t.Log("Creating ClusterServiceVersion object...")
	err = f.Client.Create(context.TODO(), &clusterServiceVersionObj, cleanUpOptions(ctx))
	assert.Nil(t, err)

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

	t.Log("Creating secret object...")
	err = f.Client.Create(context.TODO(), &secretObj, cleanUpOptions(ctx))
	assert.Nil(t, err)

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
			ApplicationSelector: v1alpha1.ApplicationSelector{
				MatchLabels: map[string]string{
					"connects-to": "postgres",
					"environment": "production",
				},
			},
		},
	}

	t.Log("Creating ServiceBindingRequest object...")
	err = f.Client.Create(context.TODO(), &serviceBindingRequestObj, cleanUpOptions(ctx))
	assert.Nil(t, err)
}
func serviceBindingRequestTest(t *testing.T, ns string, f *framework.Framework, ctx *framework.TestCtx) {
	t.Log("Starting end-to-end tests for operator...")

}
