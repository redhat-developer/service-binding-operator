package e2e

import (
	"context"
	"testing"
	"time"

	pgsqlapis "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// retrieveing deployment, to inspect it's generation
	t.Logf("Reading application deployment, extrating generation from '%s'", appName)
	require.Nil(t, f.Client.Get(todoCtx, types.NamespacedName{Namespace: ns, Name: appName}, &d))

	// creating service-binding-request, which will trigger actions in the controller
	t.Log("Creating ServiceBindingRequest mock object...")
	sbr := mocks.ServiceBindingRequestMock(ns, name, resourceRef, matchLabels)
	// making sure object does not exist before testing
	_ = f.Client.Delete(todoCtx, sbr)
	require.Nil(t, f.Client.Create(todoCtx, sbr, cleanUpOptions(ctx)))

	// waiting again for deployment
	t.Log("Waiting for application deployment reach one replica, again...")
	require.Nil(t, e2eutil.WaitForDeployment(t, f.KubeClient, ns, appName, 1, retryInterval, timeout))

	// retrieveing deployment again until new generation or timeout
	for attempts := 0; attempts < 10; attempts++ {
		t.Logf("Reading application deployment '%s' ('%d')", appName, attempts)
		require.Nil(t, f.Client.Get(todoCtx, types.NamespacedName{Namespace: ns, Name: appName}, &d))

		generation := d.GetGeneration()
		t.Logf("Deployment generation: '%d'", generation)
		if generation > 1 {
			break
		}
		time.Sleep(time.Second * 6)
	}

	// making sure envFrom is added to the container
	t.Logf("Inspecting '%s' searching for 'envFrom'...", appName)
	containers := d.Spec.Template.Spec.Containers
	require.Equal(t, 1, len(containers))
	require.Equal(t, 1, len(containers[0].EnvFrom))
	assert.NotNil(t, containers[0].EnvFrom[0].SecretRef)
	assert.Equal(t, name, containers[0].EnvFrom[0].SecretRef.Name)

	// checking intermediary secret contents
	intermediarySecretNamespacedName := types.NamespacedName{Namespace: ns, Name: name}
	_ = inspectSBRSecret(t, todoCtx, f, intermediarySecretNamespacedName)

	// editing intermediary secret in order to trigger update event
	intermediarySecretGeneration := updateSecret(t, todoCtx, f, intermediarySecretNamespacedName)

	// waiting for reconciliation, when generation changes
	sbrSecret := &corev1.Secret{}
	for attempts := 0; attempts < 10; attempts++ {
		require.Nil(t, f.Client.Get(todoCtx, intermediarySecretNamespacedName, sbrSecret))
		generation := sbrSecret.GetGeneration()

		t.Logf("Waiting for reconciliation, secret generation '%d' '%d/10'...",
			generation, attempts)
		if intermediarySecretGeneration < generation {
			t.Logf("Secret generation: '%d'", generation)
			break
		}
		time.Sleep(time.Second * 8)
	}
	t.Logf("Intermediary secret geranetion: '%d'", sbrSecret.GetGeneration())

	// inspecting secret contents again, expect to be original
	_ = inspectSBRSecret(t, todoCtx, f, intermediarySecretNamespacedName)

	// cleaning up
	t.Log("Cleaning all up!")
	_ = f.Client.Delete(todoCtx, sbr)
	_ = f.Client.Delete(todoCtx, sbrSecret)
	_ = f.Client.Delete(todoCtx, &d)
}

// inspectSBRSecret execute the inspection in a secret created by the operator.
func inspectSBRSecret(
	t *testing.T,
	ctx context.Context,
	f *framework.Framework,
	namespacedName types.NamespacedName,
) *corev1.Secret {
	t.Logf("Checking intermediary secret '%s'...", namespacedName.String())

	sbrSecret := &corev1.Secret{}
	require.Nil(t, f.Client.Get(ctx, namespacedName, sbrSecret))

	assert.Contains(t, sbrSecret.Data, "DATABASE_SECRET_USER")
	assert.Equal(t, []byte("user"), sbrSecret.Data["DATABASE_SECRET_USER"])
	assert.Contains(t, sbrSecret.Data, "DATABASE_SECRET_PASSWORD")
	assert.Equal(t, []byte("password"), sbrSecret.Data["DATABASE_SECRET_PASSWORD"])

	return sbrSecret
}

// updateSecret by exchanging all of its keys to "bogus" string.
func updateSecret(
	t *testing.T,
	ctx context.Context,
	f *framework.Framework,
	namespacedName types.NamespacedName,
) int64 {
	sbrSecret := &corev1.Secret{}
	require.Nil(t, f.Client.Get(ctx, namespacedName, sbrSecret))

	generation := sbrSecret.GetGeneration()
	t.Logf("Secret generation: '%d'", generation)

	// intentionally bumping the object generation, so the operator will reconcile;
	generation++
	sbrSecret.SetGeneration(generation)

	for k, v := range sbrSecret.Data {
		t.Logf("Replacing secret '%s=%s' with '%s=bogus'", k, string(v), k)
		sbrSecret.Data[k] = []byte("bogus")
	}

	require.Nil(t, f.Client.Update(ctx, sbrSecret))

	return generation
}
