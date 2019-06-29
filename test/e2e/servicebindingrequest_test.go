package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	pgsqlapis "github.com/baijum/postgresql-operator/pkg/apis"
	pgsql "github.com/baijum/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
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
		Items: []v1alpha1.ServiceBindingRequest{{}},
	}

	t.Log("Adding ServiceBindingRequest scheme to cluster...")
	err := framework.AddToFrameworkScheme(apis.AddToScheme, serviceBindingRequestList)
	assert.Nil(t, err)

	clusterServiceVersionListObj := &olm.ClusterServiceVersionList{
		Items: []olm.ClusterServiceVersion{{}},
	}

	t.Log("Adding ClusterServiceVersion scheme to cluster...")
	err = framework.AddToFrameworkScheme(olm.AddToScheme, clusterServiceVersionListObj)
	assert.Nil(t, err)

	databaseListObj := &pgsql.DatabaseList{
		Items: []pgsql.Database{{}},
	}

	t.Log("Adding Database scheme to cluster...")
	err = framework.AddToFrameworkScheme(pgsqlapis.AddToScheme, databaseListObj)
	assert.Nil(t, err)

	t.Run("end-to-end", func(t *testing.T) {
		t.Run("scenario-1", ServiceBindingRequest)
	})
}

// cleanUpOptions using global variables to create the object.
func cleanUpOptions(ctx *framework.TestCtx) *framework.CleanupOptions {
	return nil
	/*
	   return &framework.CleanupOptions{
	       TestContext:   ctx,
	       Timeout:       cleanupTimeout,
	       RetryInterval: time.Duration(time.Second * retryInterval),
	   }
	*/
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

// mockedObjects creates all required CRDs in the cluster, using common values to link them as
// service-binding-operator expects.
func mockedObjects(t *testing.T, ns string, f *framework.Framework, ctx *framework.TestCtx) {
	todoCtx := context.TODO()

	crdName := "postgresql.baiju.dev"
	crdVersion := "v1alpha1"
	crdKind := "Database"
	secretName := "e2e-secret"

	labelConnectTo := "postgresql"
	labelEnvironment := "e2e"

	strategy := olminstall.StrategyDetailsDeployment{
		DeploymentSpecs: []olminstall.StrategyDeploymentSpec{{
			Name: "deployment",
			Spec: appsv1.DeploymentSpec{},
		}},
	}

	strategyJSON, err := json.Marshal(strategy)
	assert.Nil(t, err)

	clusterServiceVersionObj := olm.ClusterServiceVersion{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterServiceVersion",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-cluster-service-version",
			Namespace: ns,
		},
		Spec: olm.ClusterServiceVersionSpec{
			DisplayName: "e2e database csv",
			InstallStrategy: olm.NamedInstallStrategy{
				StrategyName:    "deployment",
				StrategySpecRaw: strategyJSON,
			},
			CustomResourceDefinitions: olm.CustomResourceDefinitions{
				Owned: []olm.CRDDescription{{
					Name:        crdName,
					DisplayName: crdKind,
					Description: "e2e csv based on postgresql-operator",
					Kind:        crdKind,
					Version:     crdVersion,
					StatusDescriptors: []olm.StatusDescriptor{{
						DisplayName: "DB Password Credentials",
						Description: "Database credentials secret",
						Path:        "dbCredentials",
						XDescriptors: []string{
							"urn:alm:descriptor:io.kubernetes:Secret",
							"urn:alm:descriptor:io.servicebindingrequest:secret:user",
							"urn:alm:descriptor:io.servicebindingrequest:secret:password",
						},
					}},
				}},
			},
		},
	}

	t.Log("Creating ClusterServiceVersion object...")
	// err = f.Client.Create(todoCtx, &clusterServiceVersionObj, cleanUpOptions(ctx))
	_ = f.Client.Create(todoCtx, &clusterServiceVersionObj, cleanUpOptions(ctx))
	// assert.Nil(t, err)

	pgDatabaseObj := pgsql.Database{
		TypeMeta: metav1.TypeMeta{
			Kind:       crdKind,
			APIVersion: fmt.Sprintf("%s/%s", crdName, crdVersion),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      labelConnectTo,
			Namespace: ns,
		},
		Spec: pgsql.DatabaseSpec{
			Image:     "database/image",
			ImageName: "database",
		},
	}

	t.Log("Creating a database CRD object...")
	// err = f.Client.Create(todoCtx, &pgDatabaseObj, cleanUpOptions(ctx))
	_ = f.Client.Create(todoCtx, &pgDatabaseObj, nil)
	// assert.Nil(t, err)

	/*
	   t.Log("Adding db-credentials to status...")
	   pgDatabaseObj.Status.DBCredentials = secretName
	   err = f.Client.Status().Update(todoCtx, &pgDatabaseObj)
	   assert.Nil(t, err)
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
		Data: map[string][]byte{
			"user":     []byte("user"),
			"password": []byte("password"),
		},
	}

	t.Log("Creating secret object...")
	// err = f.Client.Create(todoCtx, &secretObj, cleanUpOptions(ctx))
	_ = f.Client.Create(todoCtx, &secretObj, cleanUpOptions(ctx))
	// assert.Nil(t, err)

	serviceBindingRequestObj := v1alpha1.ServiceBindingRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       operatorKind,
			APIVersion: operatorAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e2e-service-binding-request",
			Namespace: ns,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			BackingSelector: v1alpha1.BackingSelector{
				ResourceName:    crdName,
				ResourceVersion: crdVersion,
			},
			ApplicationSelector: v1alpha1.ApplicationSelector{
				MatchLabels: map[string]string{
					"connects-to": labelConnectTo,
					"environment": labelEnvironment,
				},
			},
		},
	}

	t.Log("Creating ServiceBindingRequest object...")
	// err = f.Client.Create(todoCtx, &serviceBindingRequestObj, cleanUpOptions(ctx))
	_ = f.Client.Create(todoCtx, &serviceBindingRequestObj, cleanUpOptions(ctx))
	// assert.Nil(t, err)
}

func serviceBindingRequestTest(t *testing.T, ns string, f *framework.Framework, ctx *framework.TestCtx) {
	t.Log("Starting end-to-end tests for operator...")
}
