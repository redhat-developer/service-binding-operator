package e2e

import (
	"context"
	"testing"
	"time"

	pgsqlapis "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
)

func TestAnnoationBasedMetadata(t *testing.T) {
	sbrlist := v1alpha1.ServiceBindingRequestList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, &sbrlist)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	dbList := pgv1alpha1.DatabaseList{}
	require.Nil(t, framework.AddToFrameworkScheme(pgsqlapis.AddToScheme, &dbList))

	dpList := appsv1.DeploymentList{}
	require.Nil(t, framework.AddToFrameworkScheme(appsv1.AddToScheme, &dpList))

	secList := corev1.SecretList{}
	require.Nil(t, framework.AddToFrameworkScheme(corev1.AddToScheme, &secList))

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: retryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}

	// get namespace
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	// get global framework variables
	f := framework.Global

	db := mocks.DatabaseCRMock(namespace, "psql")

	err = f.Client.Create(context.TODO(), db, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	// create service binding request custom resource
	name := "e2e-service-binding-request"
	resourceRef := "e2e-db-testing"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "e2e",
	}

	dp := mocks.DeploymentMock(namespace, "example1", matchLabels)
	err = f.Client.Create(context.TODO(), &dp, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	sbr := mocks.ServiceBindingRequestMock(namespace, name, resourceRef, "", matchLabels, false)
	err = f.Client.Create(context.TODO(), sbr, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Minute)

	sbrSecret := &corev1.Secret{}
	namespacedName := types.NamespacedName{Namespace: namespace, Name: name}
	if err = f.Client.Get(context.TODO(), namespacedName, sbrSecret); err != nil {
		t.Error(err)
	}
}
