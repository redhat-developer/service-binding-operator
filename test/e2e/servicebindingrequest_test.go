package e2e

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	pgsqlapis "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/redhat-developer/service-binding-operator/pkg/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

type Step string
type OnSBRCreate func(sbr *v1alpha1.ServiceBindingRequest)

const (
	DBStep          Step = "create-db"
	AppStep         Step = "create-app"
	SBRStep         Step = "create-sbr"
	SBREtcdStep     Step = "create-etcd-sbr"
	EtcdClusterStep Step = "create-etcd-cluster"
	CSVStep         Step = "create-csv"
)

var (
	retryInterval  = time.Second * 5
	timeout        = time.Second * 180
	cleanupTimeout = time.Second * 5

	// Intermediate secret should have following data
	// for postgres operator
	postgresSecretAssertion = map[string]string{
		"DATABASE_SECRET_USERNAME": "user",
		"DATABASE_SECRET_PASSWORD": "password",
	}

	// Intermediate secret should have following data
	// for etcd operator
	etcdSecretAssertion = map[string]string{
		"ETCDCLUSTER_CLUSTERIP": "172.30.0.129",
	}
	deploymentsGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
)

// TestAddSchemesToFramework starting point of the test, it declare the CRDs that will be using
// during end-to-end tests.
func TestAddSchemesToFramework(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	t.Log("Adding ServiceBindingRequestList scheme to cluster...")
	sbrlist := v1alpha1.ServiceBindingRequestList{}
	require.NoError(t, framework.AddToFrameworkScheme(apis.AddToScheme, &sbrlist))

	t.Log("Adding ClusterServiceVersionList scheme to cluster...")
	csvList := olmv1alpha1.ClusterServiceVersionList{}
	require.NoError(t, framework.AddToFrameworkScheme(olmv1alpha1.AddToScheme, &csvList))

	t.Log("Adding DatabaseList scheme to cluster...")
	dbList := pgv1alpha1.DatabaseList{}
	require.NoError(t, framework.AddToFrameworkScheme(pgsqlapis.AddToScheme, &dbList))

	t.Log("Adding EtcdCluster scheme to cluster...")
	etcdCluster := v1beta2.EtcdCluster{}
	require.Nil(t, framework.AddToFrameworkScheme(v1beta2.AddToScheme, &etcdCluster))

	t.Run("end-to-end", func(t *testing.T) {
		t.Run("scenario-etcd-unannotated-app-db-sbr", func(t *testing.T) {
			ServiceBindingRequest(t, []Step{AppStep, EtcdClusterStep, SBREtcdStep})
		})
		t.Run("scenario-db-app-sbr", func(t *testing.T) {
			ServiceBindingRequest(t, []Step{DBStep, AppStep, SBRStep})
		})
		t.Run("scenario-app-db-sbr", func(t *testing.T) {
			ServiceBindingRequest(t, []Step{AppStep, DBStep, SBRStep})
		})
		t.Run("scenario-db-sbr-app", func(t *testing.T) {
			t.Skip("Currently disabled as not supported by SBO")
			ServiceBindingRequest(t, []Step{DBStep, SBRStep, AppStep})
		})
		t.Run("scenario-app-sbr-db", func(t *testing.T) {
			t.Skip("Currently disabled as not supported by SBO")
			ServiceBindingRequest(t, []Step{AppStep, SBRStep, DBStep})
		})
		t.Run("scenario-sbr-db-app", func(t *testing.T) {
			t.Skip("Currently disabled as not supported by SBO")
			ServiceBindingRequest(t, []Step{SBRStep, DBStep, AppStep})
		})
		t.Run("scenario-sbr-app-db", func(t *testing.T) {
			t.Skip("Currently disabled as not supported by SBO")
			ServiceBindingRequest(t, []Step{SBRStep, AppStep, DBStep})
		})
		t.Run("scenario-csv-db-app-sbr", func(t *testing.T) {
			ServiceBindingRequest(t, []Step{CSVStep, DBStep, AppStep, SBRStep})
		})
		t.Run("scenario-csv-app-db-sbr", func(t *testing.T) {
			ServiceBindingRequest(t, []Step{CSVStep, AppStep, DBStep, SBRStep})
		})
	})
}

// cleanupOptions using global variables to create the object.
func cleanupOptions(ctx *framework.Context) *framework.CleanupOptions {
	return &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: time.Duration(time.Second * retryInterval),
	}
}

// bootstrapNamespace execute scaffolding to have a new cluster initialized, and acquire a test
// namespace, the namespace name is returned and framework global variables are returned.
func bootstrapNamespace(t *testing.T, ctx *framework.Context) (string, *framework.Framework) {
	t.Log("Initializing cluster resources...")
	err := ctx.InitializeClusterResources(cleanupOptions(ctx))
	if err != nil {
		t.Logf("Cluster resources initialization error: '%s'", err)
		require.True(t, errors.IsAlreadyExists(err), "failed to setup cluster resources")
	}

	// namespace name is informed on command-line or defined dinamically
	ns, err := ctx.GetNamespace()
	require.NoError(t, err)
	t.Logf("Using namespace '%s' for testing...", ns)

	f := framework.Global
	return ns, f
}

// ServiceBindingRequest bootstrap method to initialize cluster resources and setup a testing
// namespace, after bootstrap operator related tests method is called out.
func ServiceBindingRequest(t *testing.T, steps []Step) {
	t.Log("Creating a new test context...")
	ctx := framework.NewContext(t)
	defer ctx.Cleanup()

	ns, f := bootstrapNamespace(t, ctx)

	// executing testing steps on operator
	serviceBindingRequestTest(t, ctx, f, ns, steps)
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

// assertSBRStatus will determine if SBR is on "success" state
func assertSBRStatus(
	ctx context.Context,
	t *testing.T,
	f *framework.Framework,
	namespacedName types.NamespacedName,
) error {
	sbr := &v1alpha1.ServiceBindingRequest{}
	if err := f.Client.Get(ctx, namespacedName, sbr); err != nil {
		return err
	}

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbr.Status.Conditions,
			servicebindingrequest.CollectionReady,
			corev1.ConditionTrue,
		),
		"CollectionReady condition should exist and true; existing conditions: %+v",
		sbr.Status.Conditions,
	)

	require.True(t,
		conditionsv1.IsStatusConditionPresentAndEqual(
			sbr.Status.Conditions,
			servicebindingrequest.InjectionReady,
			corev1.ConditionTrue,
		),
		"InjectionReady condition should exist and true; existing conditions: %+v",
		sbr.Status.Conditions,
	)
	return nil
}

// assertSBRSecret execute the inspection in a secret created by the operator.
func assertSBRSecret(
	ctx context.Context,
	f *framework.Framework,
	namespacedName types.NamespacedName,
	assertKeys map[string]string,
) (*corev1.Secret, error) {
	sbrSecret := &corev1.Secret{}
	if err := f.Client.Get(ctx, namespacedName, sbrSecret); err != nil {
		return nil, err
	}

	for k, v := range assertKeys {
		if _, contains := sbrSecret.Data[k]; !contains {
			return nil, fmt.Errorf("can't find %q in data", k)
		}
		actual := sbrSecret.Data[k]
		expected := []byte(v)
		if !bytes.Equal(expected, actual) {
			return nil, fmt.Errorf("key %s (%s) is different than expected (%s)",
				k, actual, expected)
		}
	}

	return sbrSecret, nil
}

// assertSecretNotFound execute assertion to make sure a secret is not found.
func assertSecretNotFound(
	ctx context.Context,
	f *framework.Framework,
	namespacedName types.NamespacedName,
) error {
	secret := &corev1.Secret{}
	err := f.Client.Get(ctx, namespacedName, secret)
	if err == nil {
		return fmt.Errorf("secret '%s' still found", namespacedName)
	}
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

// updateSBRSecret by exchanging all of its keys to "bogus" string.
func updateSBRSecret(
	ctx context.Context,
	t *testing.T,
	f *framework.Framework,
	namespacedName types.NamespacedName,
) error {
	sbrSecret := &corev1.Secret{}
	require.NoError(t, f.Client.Get(ctx, namespacedName, sbrSecret))

	// intentionally bumping the object generation, so the operator will reconcile;
	generation := sbrSecret.GetGeneration()
	generation++
	sbrSecret.SetGeneration(generation)

	for k, v := range sbrSecret.Data {
		t.Logf("Replacing secret '%s=%s' with '%s=bogus'", k, string(v), k)
		sbrSecret.Data[k] = []byte("bogus")
	}

	return f.Client.Update(ctx, sbrSecret)
}

// CreateDB implements end-to-end step for the creation of a Database CR along with the dependend
// Secret serving as a Backing Service to be bound to the application.
func CreateDB(
	ctx context.Context,
	t *testing.T,
	f *framework.Framework,
	cleanupOpts *framework.CleanupOptions,
	namespacedName types.NamespacedName,
	secretName string,
) *pgv1alpha1.Database {
	t.Logf("Creating Database mock object '%#v'...", namespacedName)
	ns := namespacedName.Namespace
	resourceRef := namespacedName.Name

	// order is important: the operator will follow first the database resource,
	// then the secret holding the credential; in the case the secret is not
	// there, default values will be set for each of the contributed
	// configuration values declared in the Database custom resource definition.
	t.Log("Creating Database credentials secret mock object...")
	dbSecret := mocks.SecretMock(ns, secretName, nil)
	require.NoError(t, f.Client.Create(ctx, dbSecret, cleanupOpts))

	db := mocks.DatabaseCRMock(ns, resourceRef)
	require.NoError(t, f.Client.Create(ctx, db, cleanupOpts))

	t.Logf("Updating Database '%#v' status, adding 'DBCredentials'", namespacedName)
	require.Eventually(t, func() bool {
		if err := f.Client.Get(ctx, namespacedName, db); err != nil {
			t.Logf("get error: %s", err)
			return false
		}
		db.Status.DBCredentials = secretName
		if err := f.Client.Status().Update(ctx, db); err != nil {
			t.Logf("update error: %s", err)
			return false
		}
		return true
	}, 10*time.Second, 1*time.Second)

	return db
}

// CreateApp implements end-to-end step for the creation of a Deployment serving as the Application
// to which the Backing Service is bound.
func CreateApp(
	ctx context.Context,
	t *testing.T,
	f *framework.Framework,
	cleanupOpts *framework.CleanupOptions,
	namespacedName types.NamespacedName,
	matchLabels map[string]string,
) appsv1.Deployment {
	t.Logf("Creating Deployment mock object '%#v'...", namespacedName)
	ns := namespacedName.Namespace
	appName := namespacedName.Name

	d := mocks.DeploymentMock(ns, appName, matchLabels)
	require.NoError(t, f.Client.Create(ctx, &d, cleanupOpts))

	// waiting for application deployment to reach one replica
	t.Log("Waiting for application deployment reach one replica...")
	require.NoError(
		t,
		e2eutil.WaitForDeployment(t, f.KubeClient, ns, appName, 1, retryInterval, timeout),
	)

	// retrieveing deployment, to inspect its contents
	t.Logf("Reading application deployment '%s'", appName)
	require.NoError(t, f.Client.Get(ctx, namespacedName, &d))

	return d
}

// CreateSBR implements end-to-end step for creating a Service Binding Request to bind the Backing
// Service and the Application.
func CreateSBR(
	ctx context.Context,
	t *testing.T,
	f *framework.Framework,
	cleanupOpts *framework.CleanupOptions,
	namespacedName types.NamespacedName,
	resourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
	onSBRCreate OnSBRCreate,
) *v1alpha1.ServiceBindingRequest {
	t.Logf("Creating ServiceBindingRequest mock object '%#v'...", namespacedName)
	sbr := mocks.ServiceBindingRequestMock(
		namespacedName.Namespace, namespacedName.Name, nil, resourceRef, "", applicationGVR, matchLabels)

	// This function call explicitly modifies default SBR created by
	// the mock
	if onSBRCreate != nil {
		onSBRCreate(sbr)
	}

	// wait deletion of sbr, this should give time for the operator to finalize the unbind operation
	// and remove the finalizer
	err := e2eutil.WaitForDeletion(t, f.Client.Client, sbr, retryInterval, timeout)
	require.NoError(t, err, "expect waiting for deletion to not return errors")

	require.NoError(t, f.Client.Create(ctx, sbr, cleanupOpts))
	return sbr
}

// setSBRBackendGVK sets backend service selector
func setSBRBackendGVK(
	sbr *v1alpha1.ServiceBindingRequest,
	resourceRef string,
	backendGVK schema.GroupVersionKind,
	envVarPrefix string,
) {
	sbr.Spec.BackingServiceSelector = &v1alpha1.BackingServiceSelector{
		GroupVersionKind: metav1.GroupVersionKind{Group: backendGVK.Group, Version: backendGVK.Version, Kind: backendGVK.Kind},
		ResourceRef:      resourceRef,
		EnvVarPrefix:     &envVarPrefix,
	}
}

// setSBRBindUnannotated makes SBR to detect bindable resource
// without depending on annotation
func setSBRBindUnannotated(sbr *v1alpha1.ServiceBindingRequest, bindUnAnnotated bool) {
	sbr.Spec.DetectBindingResources = bindUnAnnotated
}

// CreateCSV created mocked cluster service version object.
func CreateCSV(
	ctx context.Context,
	t *testing.T,
	f *framework.Framework,
	cleanupOpts *framework.CleanupOptions,
	namespacedName types.NamespacedName,
) {
	t.Logf("Creating ClusterServiceVersion mock object: '%#v'...", namespacedName)
	csv := mocks.ClusterServiceVersionMock(namespacedName.Namespace, namespacedName.Name)
	require.NoError(t, f.Client.Create(ctx, &csv, cleanupOpts))
}

// serviceBindingRequestTest executes the actual end-to-end testing, simulating the components and
// expecting for changes caused by the operator.
func serviceBindingRequestTest(
	t *testing.T,
	ctx *framework.Context,
	f *framework.Framework,
	ns string,
	steps []Step,
) {
	// making sure resource names employed during test are unique
	rand.Seed(time.Now().UnixNano())
	randomSuffix := rand.Int()
	csvName := fmt.Sprintf("cluster-service-version-%d", randomSuffix)
	sbrName := fmt.Sprintf("e2e-service-binding-request-%d", randomSuffix)
	resourceRef := fmt.Sprintf("e2e-db-testing-%d", randomSuffix)
	secretName := fmt.Sprintf("e2e-db-credentials-%d", randomSuffix)
	appName := fmt.Sprintf("e2e-application-%d", randomSuffix)
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": fmt.Sprintf("e2e-%d", randomSuffix),
	}

	t.Logf("Starting end-to-end tests for operator, using suffix '%d'!", randomSuffix)

	resourceRefNamespacedName := types.NamespacedName{Namespace: ns, Name: resourceRef}
	deploymentNamespacedName := types.NamespacedName{Namespace: ns, Name: appName}
	sbrNamespacedName := types.NamespacedName{Namespace: ns, Name: sbrName}
	csvNamespacedName := types.NamespacedName{Namespace: ns, Name: csvName}
	intermediarySecretNamespacedName := types.NamespacedName{Namespace: ns, Name: sbrName}

	cleanupOpts := cleanupOptions(ctx)

	todoCtx := context.TODO()
	assertKeys := postgresSecretAssertion

	var sbr *v1alpha1.ServiceBindingRequest
	for _, step := range steps {
		switch step {
		case CSVStep:
			CreateCSV(todoCtx, t, f, cleanupOpts, csvNamespacedName)
		case DBStep:
			CreateDB(todoCtx, t, f, cleanupOpts, resourceRefNamespacedName, secretName)
		case AppStep:
			CreateApp(todoCtx, t, f, cleanupOpts, deploymentNamespacedName, matchLabels)
		case SBRStep:
			// creating service-binding-request, which will trigger actions in the controller
			sbr = CreateSBR(todoCtx, t, f, cleanupOpts, sbrNamespacedName, resourceRef, deploymentsGVR, matchLabels, nil)
		case SBREtcdStep:
			assertKeys = etcdSecretAssertion
			sbr = CreateSBR(todoCtx, t, f,
				cleanupOpts,
				sbrNamespacedName,
				resourceRef,
				deploymentsGVR,
				matchLabels,
				func(sbr *v1alpha1.ServiceBindingRequest) {
					setSBRBackendGVK(sbr, resourceRef,
						v1beta2.SchemeGroupVersion.WithKind(v1beta2.EtcdClusterResourceKind),
						"ETCDCLUSTER",
					)
					setSBRBindUnannotated(sbr, true)
				})
		case EtcdClusterStep:
			CreateEtcdCluster(todoCtx, t, ctx, f, resourceRefNamespacedName)
		}
	}

	// retrying a few times to identify SBO changes in deployment, this loop is waiting for the
	// operator reconciliation.
	t.Log("Inspecting deployment structure...")
	require.Eventually(t, func() bool {
		t.Logf("Inspecting deployment '%s'", deploymentNamespacedName)
		_, err := assertDeploymentEnvFrom(todoCtx, f, deploymentNamespacedName, sbrName)
		if err != nil {
			t.Logf("Error on inspecting deployment: '%s'", err)
			return false
		}
		t.Logf("Deployment: Result after attempts, error: '%#v'", err)
		return true
	}, 120*time.Second, 2*time.Second)

	// retrying a few times to identify SBR status change to "success"
	t.Log("Inspecting SBR status...")
	require.Eventually(t, func() bool {
		t.Logf("Inspecting SBR: '%s'", sbrNamespacedName)
		err := assertSBRStatus(todoCtx, t, f, sbrNamespacedName)
		if err != nil {
			t.Logf("Error on inspecting SBR: '%#v'", err)
			return false
		}
		t.Logf("SBR-Status: Result after attempts, error: '%#v'", err)
		return true
	}, 120*time.Second, 2*time.Second)

	require.Eventually(t, func() bool {
		// checking intermediary secret contents, right after deployment the secrets must be in place
		sbrSecret, err := assertSBRSecret(todoCtx, f, intermediarySecretNamespacedName, assertKeys)
		if err != nil {
			// it can be that assertSBRSecret returns nil until service configuration values can be
			// collected.
			t.Logf("binding secret contents are invalid: %v", sbrSecret)
			return false
		}
		return true
	}, 120*time.Second, 2*time.Second)

	// editing intermediary secret in order to trigger update event
	require.Eventually(t, func() bool {
		t.Logf("Updating intermediary secret to have bogus data: '%s'", intermediarySecretNamespacedName)
		err := updateSBRSecret(todoCtx, t, f, intermediarySecretNamespacedName)
		if err != nil {
			t.Log(err)
			return false
		}
		return true
	}, 120*time.Second, 2*time.Second)

	// retrying a few times to see if secret is back on original state, waiting for operator to
	// reconcile again when detecting the change
	t.Log("Inspecting intermediary secret...")
	require.Eventually(t, func() bool {
		t.Logf("Inspecting secret '%s'...", intermediarySecretNamespacedName)
		_, err := assertSBRSecret(todoCtx, f, intermediarySecretNamespacedName, assertKeys)
		if err != nil {
			t.Logf("Secret inspection error: '%#v'", err)
			return false
		}
		t.Logf("Secret: Result after attempts, error: '%#v'", err)
		return true
	}, 120*time.Second, 2*time.Second)

	// executing deletion of the request, triggering unbinding actions
	err := f.Client.Delete(todoCtx, sbr)
	require.NoError(t, err, "expect deletion to not return errors")

	// after deletion, secret should not be found anymore
	require.Eventually(t, func() bool {
		t.Logf("Searching for secret '%s'...", intermediarySecretNamespacedName)
		err := assertSecretNotFound(todoCtx, f, intermediarySecretNamespacedName)
		if err != nil {
			t.Logf("Secret search error: '%#v'", err)
			return false
		}
		t.Logf("Secret: Result after attempts, error: '%#v'", err)
		return true
	}, 120*time.Second, 2*time.Second)

	// after deletion, deployment should not contain envFrom directive anymore
	require.Eventually(t, func() bool {
		t.Logf("Inspecting deployment '%s'", deploymentNamespacedName)
		_, err := assertDeploymentEnvFrom(todoCtx, f, deploymentNamespacedName, sbrName)
		if err != nil {
			t.Log("expect deployment not to be carrying envFrom directive")
			return true
		}
		return false
	}, 120*time.Second, 2*time.Second)
}

func CreateEtcdCluster(
	todoCtx context.Context,
	t *testing.T,
	ctx *framework.Context,
	f *framework.Framework,
	namespacedName types.NamespacedName,
) (*v1beta2.EtcdCluster, *corev1.Service) {
	ns := namespacedName.Namespace
	name := namespacedName.Name
	t.Log("Create etcd cluster")
	etcd, etcdSvc := mocks.CreateEtcdClusterMock(ns, name)
	require.Nil(t, f.Client.Create(todoCtx, etcd, cleanupOptions(ctx)))
	trueBool := true
	falseBool := false
	etcdSvc.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         v1beta2.SchemeGroupVersion.Version,
			Kind:               v1beta2.EtcdClusterResourceKind,
			Name:               etcd.Name,
			UID:                etcd.UID,
			Controller:         &trueBool,
			BlockOwnerDeletion: &falseBool,
		},
	})
	require.Nil(t, f.Client.Create(todoCtx, etcdSvc, cleanupOptions(ctx)))
	return etcd, etcdSvc
}
