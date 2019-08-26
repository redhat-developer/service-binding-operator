package e2e

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	pgsqlapis "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

var (
	retryInterval  = time.Second * 5
	timeout        = time.Second * 120
	cleanupTimeout = time.Second * 5
)

// TestAddSchemesToFramework starting point of the test, it declare the CRDs that will be using
// during end-to-end tests.
func TestAddSchemesToFramework(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	t.Log("Adding ServiceBindingRequestList scheme to cluster...")
	sbrlist := v1alpha1.ServiceBindingRequestList{}
	require.Nil(t, framework.AddToFrameworkScheme(apis.AddToScheme, &sbrlist))

	t.Log("Adding ClusterServiceVersionList scheme to cluster...")
	csvList := olmv1alpha1.ClusterServiceVersionList{}
	require.Nil(t, framework.AddToFrameworkScheme(olmv1alpha1.AddToScheme, &csvList))

	t.Log("Adding DatabaseList scheme to cluster...")
	dbList := pgv1alpha1.DatabaseList{}
	require.Nil(t, framework.AddToFrameworkScheme(pgsqlapis.AddToScheme, &dbList))

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

// bootstrapNamespace execute scaffolding to have a new cluster initialized, and acquire a test
// namespace, the namespace name is returned and framework global variables are returned.
func bootstrapNamespace(t *testing.T, ctx *framework.TestCtx) (string, *framework.Framework) {
	t.Log("Initializing cluster resources...")
	err := ctx.InitializeClusterResources(cleanUpOptions(ctx))
	if err != nil {
		t.Logf("Cluster resources initialization error: '%s'", err)
		require.True(t, errors.IsAlreadyExists(err), "failed to setup cluster resources")
	}

	// namespace name is informed on command-line or defined dinamically
	ns, err := ctx.GetNamespace()
	require.Nil(t, err)
	t.Logf("Using namespace '%s' for testing...", ns)

	f := framework.Global
	return ns, f
}

// ServiceBindingRequest bootstrap method to initialize cluster resources and setup a testing
// namespace, after bootstrap operator related tests method is called out.
func ServiceBindingRequest(t *testing.T) {
	t.Log("Creating a new test context...")
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	ns, f := bootstrapNamespace(t, ctx)

	// executing testing steps on operator
	serviceBindingRequestTest(t, ctx, f, ns)
}

// assertDeploymentEnvFrom execute the inspection of a deployment type, making sure the containers
// are set, and are having "envFrom" directive.
func assertDeploymentEnvFrom(
	ctx context.Context,
	f *framework.Framework,
	namespacedName types.NamespacedName,
	secretRefName string,
) (*appsv1.Deployment, error) {
	d := &appsv1.Deployment{}
	if err := f.Client.Get(ctx, namespacedName, d); err != nil {
		return nil, err
	}

	containers := d.Spec.Template.Spec.Containers

	if len(containers) != 1 {
		return nil, fmt.Errorf("can't find a container in deployment-spec")
	}
	if len(containers[0].EnvFrom) != 1 {
		return nil, fmt.Errorf("can't find envFrom in first container")
	}
	if secretRefName != containers[0].EnvFrom[0].SecretRef.Name {
		return nil, fmt.Errorf("secret-ref attribute named '%s' not found", secretRefName)
	}

	return d, nil
}

// assertSBRSecret execute the inspection in a secret created by the operator.
func assertSBRSecret(
	ctx context.Context,
	f *framework.Framework,
	namespacedName types.NamespacedName,
) (*corev1.Secret, error) {
	sbrSecret := &corev1.Secret{}
	if err := f.Client.Get(ctx, namespacedName, sbrSecret); err != nil {
		return nil, err
	}

	if _, contains := sbrSecret.Data["DATABASE_SECRET_USER"]; !contains {
		return nil, fmt.Errorf("can't find DATABASE_SECRET_USER in data")
	}
	if !bytes.Equal([]byte("user"), sbrSecret.Data["DATABASE_SECRET_USER"]) {
		return nil, fmt.Errorf("key DATABASE_SECRET_USER is different than expected")
	}

	if _, contains := sbrSecret.Data["DATABASE_SECRET_PASSWORD"]; !contains {
		return nil, fmt.Errorf("can't find DATABASE_SECRET_PASSWORD in data")
	}
	if !bytes.Equal([]byte("password"), sbrSecret.Data["DATABASE_SECRET_PASSWORD"]) {
		return nil, fmt.Errorf("key DATABASE_SECRET_PASSWORD is different than expected")
	}

	return sbrSecret, nil
}

// updateSecret by exchanging all of its keys to "bogus" string.
func updateSecret(
	ctx context.Context,
	t *testing.T,
	f *framework.Framework,
	namespacedName types.NamespacedName,
) {
	sbrSecret := &corev1.Secret{}
	require.Nil(t, f.Client.Get(ctx, namespacedName, sbrSecret))

	// intentionally bumping the object generation, so the operator will reconcile;
	generation := sbrSecret.GetGeneration()
	generation++
	sbrSecret.SetGeneration(generation)

	for k, v := range sbrSecret.Data {
		t.Logf("Replacing secret '%s=%s' with '%s=bogus'", k, string(v), k)
		sbrSecret.Data[k] = []byte("bogus")
	}

	require.Nil(t, f.Client.Update(ctx, sbrSecret))
}

// retry the informed method a few times, with sleep between attempts.
func retry(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := attempts; i > 0; i-- {
		err = fn()
		if err == nil {
			break
		}
	}
	return err
}

// serviceBindingRequestTest executes the actual end-to-end testing, simulating the components and
// expecting for changes caused by the operator.
func serviceBindingRequestTest(t *testing.T, ctx *framework.TestCtx, f *framework.Framework, ns string) {
	todoCtx := context.TODO()

	name := "e2e-service-binding-request"
	resourceRef := "e2e-db-testing"
	secretName := "e2e-db-credentials"
	appName := "e2e-application"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "e2e",
	}

	t.Log("Starting end-to-end tests for operator!")

	t.Log("Creating ClusterServiceVersion mock object...")
	csv := mocks.ClusterServiceVersionMock(ns, "cluster-service-version")
	require.Nil(t, f.Client.Create(todoCtx, &csv, cleanUpOptions(ctx)))

	t.Log("Creating Database mock object...")
	db := mocks.DatabaseCRMock(ns, resourceRef)
	require.Nil(t, f.Client.Create(todoCtx, db, cleanUpOptions(ctx)))

	t.Log("Updating Database status, adding 'DBCredentials'")
	require.Nil(t, f.Client.Get(todoCtx, types.NamespacedName{Namespace: ns, Name: resourceRef}, db))
	db.Status.DBCredentials = secretName
	require.Nil(t, f.Client.Status().Update(todoCtx, db))

	t.Log("Creating Database credentials secret mock object...")
	dbSecret := mocks.SecretMock(ns, secretName)
	require.Nil(t, f.Client.Create(todoCtx, dbSecret, cleanUpOptions(ctx)))

	t.Log("Creating Deployment mock object...")
	d := mocks.DeploymentMock(ns, appName, matchLabels)
	require.Nil(t, f.Client.Create(todoCtx, &d, cleanUpOptions(ctx)))

	// waiting for application deployment to reach one replica
	t.Log("Waiting for application deployment reach one replica...")
	require.Nil(t, e2eutil.WaitForDeployment(t, f.KubeClient, ns, appName, 1, retryInterval, timeout))

	// retrieveing deployment, to inspect its contents
	deploymentNamespacedName := types.NamespacedName{Namespace: ns, Name: appName}
	t.Logf("Reading application deployment '%s'", appName)
	require.Nil(t, f.Client.Get(todoCtx, deploymentNamespacedName, &d))

	// creating service-binding-request, which will trigger actions in the controller
	t.Log("Creating ServiceBindingRequest mock object...")
	sbr := mocks.ServiceBindingRequestMock(ns, name, resourceRef, matchLabels)
	// making sure object does not exist before testing
	_ = f.Client.Delete(todoCtx, sbr)
	require.Nil(t, f.Client.Create(todoCtx, sbr, cleanUpOptions(ctx)))

	// waiting again for deployment
	t.Log("Waiting for application deployment reach one replica, again...")
	require.Nil(t, e2eutil.WaitForDeployment(t, f.KubeClient, ns, appName, 1, retryInterval, timeout))

	// retrying a few times to identify SBO changes in deployment, this loop is waiting for the
	// operator reconciliation.
	t.Log("Inspecting deployment structure...")
	err := retry(10, 5*time.Second, func() error {
		t.Logf("Inspecting deployment '%s'", deploymentNamespacedName)
		_, err := assertDeploymentEnvFrom(todoCtx, f, deploymentNamespacedName, name)
		if err != nil {
			t.Logf("Error on inspecting deployment: '%#v'", err)
		}
		return err
	})
	t.Logf("Deployment: Result after attempts, error: '%#v'", err)
	assert.NoError(t, err)

	// checking intermediary secret contents, right after deployment the secrets must be in place
	intermediarySecretNamespacedName := types.NamespacedName{Namespace: ns, Name: name}
	sbrSecret, err := assertSBRSecret(todoCtx, f, intermediarySecretNamespacedName)
	assert.NoError(t, err)

	// editing intermediary secret in order to trigger update event
	t.Logf("Updating intermediary secret to have bogus data: '%s'", intermediarySecretNamespacedName)
	updateSecret(todoCtx, t, f, intermediarySecretNamespacedName)

	// retrying a few times to see if secret is back on original state, waiting for operator to
	// reconcile again when detecting the change
	t.Log("Inspecting intermediary secret...")
	err = retry(10, 5*time.Second, func() error {
		t.Log("Inspecting SBR generated secret...")
		_, err := assertSBRSecret(todoCtx, f, intermediarySecretNamespacedName)
		if err != nil {
			t.Logf("SBR generated secret inspection error: '%#v'", err)
		}
		return err
	})
	t.Logf("Intermediary-Secret: Result after attempts, error: '%#v'", err)
	assert.NoError(t, err)

	// cleaning up
	t.Log("Cleaning all up!")
	_ = f.Client.Delete(todoCtx, sbr)
	_ = f.Client.Delete(todoCtx, sbrSecret)
	_ = f.Client.Delete(todoCtx, &d)
}
