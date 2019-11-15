package e2e

import (
	"context"
	"testing"
	"time"

	ocav1 "github.com/openshift/api/apps/v1"
	pgsqlapis "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/redhat-developer/service-binding-operator/pkg/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestBackingServiceWithDeploymentConfig(t *testing.T) {
	sbrlist := v1alpha1.ServiceBindingRequestList{}
	require.NoError(t, framework.AddToFrameworkScheme(apis.AddToScheme, &sbrlist), "failed to add custom resource scheme to framework")

	dbList := pgv1alpha1.DatabaseList{}
	require.NoError(t, framework.AddToFrameworkScheme(pgsqlapis.AddToScheme, &dbList))

	dcList := ocav1.DeploymentConfigList{}
	require.NoError(t, framework.AddToFrameworkScheme(ocav1.AddToScheme, &dcList))

	secList := corev1.SecretList{}
	require.NoError(t, framework.AddToFrameworkScheme(corev1.AddToScheme, &secList))

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	// get namespace
	namespace, err := ctx.GetNamespace()
	require.NoError(t, err)

	err = ctx.InitializeClusterResources(
		&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: retryInterval},
	)
	require.NoError(t, err, "failed to initialize cluster resources")

	resourceRef := "deploymentconfig-test"
	db := mocks.DatabaseCRMock(namespace, resourceRef)

	// get global framework variables
	f := framework.Global
	err = f.Client.Create(context.TODO(), db, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	require.NoError(t, err)

	// create service binding request custom resource
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "e2e",
	}
	appResourceRef := "application1"
	dp := mocks.DeploymentConfigMock(namespace, appResourceRef, matchLabels)
	err = f.Client.Create(context.TODO(), &dp, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	require.NoError(t, err)

	name := "e2e-service-binding-request"
	sbr := mocks.ServiceBindingRequestMock(namespace, name, resourceRef, appResourceRef, matchLabels, false)
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
}
