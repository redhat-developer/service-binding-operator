package e2e

import (
	"context"
	"testing"
	"time"

	pgsqlapis "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	knativev1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestKnativeService(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	f := framework.Global
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	retryInterval := time.Second * 5
	cleanupTimeout := time.Second * 5

	//Registering schemes
	t.Log("Adding ServiceBindingRequestList scheme to cluster...")
	sbrlist := v1alpha1.ServiceBindingRequestList{}
	require.Nil(t, framework.AddToFrameworkScheme(apis.AddToScheme, &sbrlist))

	t.Log("Adding ClusterServiceVersionList scheme to cluster...")
	csvList := olmv1alpha1.ClusterServiceVersionList{}
	require.Nil(t, framework.AddToFrameworkScheme(olmv1alpha1.AddToScheme, &csvList))

	t.Log("Adding DatabaseList scheme to cluster...")
	dbList := pgv1alpha1.DatabaseList{}
	require.Nil(t, framework.AddToFrameworkScheme(pgsqlapis.AddToScheme, &dbList))

	t.Log("Adding SecretList scheme to cluster...")
	secList := corev1.SecretList{}
	require.NoError(t, framework.AddToFrameworkScheme(corev1.AddToScheme, &secList))

	t.Log("Adding KnativeService scheme to cluster...")
	serviceList := knativev1.ServiceList{}
	require.Nil(t, framework.AddToFrameworkScheme(knativev1.AddToScheme, &serviceList))

	//Initilizing cluster
	t.Log("Inilizing cluster...")
	err := ctx.InitializeClusterResources(
		&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: retryInterval},
	)
	require.NoError(t, err, "failed to initialize cluster resources")

	// get namespace
	t.Log("Getting namespace...")
	namespace, err := ctx.GetNamespace()
	require.NoError(t, err)
	t.Log("Namespace: ", namespace)

	//Create database mock CR
	t.Log("Creating database CR...")
	resourceRef := "e2e-db-testing"
	db := mocks.DatabaseCRMock(namespace, resourceRef)
	err = f.Client.Create(context.TODO(), db, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	require.NoError(t, err)

	//Create knative service mock CR
	t.Log("Creating knative service CR...")
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "e2e",
	}
	serviceRef := "knative-app"
	ks := mocks.KnativeServiceMock(namespace, serviceRef, matchLabels)
	err = f.Client.Create(context.TODO(), &ks, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	require.NoError(t, err)

	// create service binding request custom resource
	t.Log("Creating service binding request...")
	name := "e2e-service-binding-request"
	gvr := knativev1.SchemeGroupVersion.WithResource("services") // Group/Version/Resource for sbr
	sbr := mocks.ServiceBindingRequestMock(namespace, name, resourceRef, serviceRef, matchLabels, false, gvr)
	err = f.Client.Create(context.TODO(), sbr, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	require.NoError(t, err)

	time.Sleep(1 * time.Minute)

	namespacedName := types.NamespacedName{Namespace: namespace, Name: name}
	sbr2 := &v1alpha1.ServiceBindingRequest{}

	err = wait.Poll(2*time.Second, 30*time.Second, func() (done bool, err error) {
		if err = f.Client.Get(context.TODO(), namespacedName, sbr2); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return true, err
		}
		return true, nil
	})
	require.NoError(t, err)
	//Assert sbr secret
	sbrSecret := &corev1.Secret{}
	require.NoError(t, f.Client.Get(context.TODO(), namespacedName, sbrSecret))
	require.Equal(t, []byte("test-db"), sbrSecret.Data["DATABASE_DBNAME"], "Name not equal")

	//Assert knative service SecretRef
	kserv := &knativev1.Service{}
	namespacedName2 := types.NamespacedName{Namespace: namespace, Name: serviceRef}
	require.NoError(t, f.Client.Get(context.TODO(), namespacedName2, kserv))
	require.Equal(t, name, kserv.Spec.Template.Spec.Containers[0].EnvFrom[0].SecretRef.Name, "secret reference doesn't match")
}
