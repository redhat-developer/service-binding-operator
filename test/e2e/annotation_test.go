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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

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

	resourceRef := "e2e-db-testing"
	db := mocks.DatabaseCRMock(namespace, resourceRef)

	err = f.Client.Create(context.TODO(), db, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	// create service binding request custom resource
	name := "e2e-service-binding-request"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "e2e",
	}
	appResourceRef := "example1"
	dp := mocks.DeploymentMock(namespace, appResourceRef, matchLabels)
	err = f.Client.Create(context.TODO(), &dp, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

	sbr := mocks.ServiceBindingRequestMock(namespace, name, resourceRef, appResourceRef, matchLabels, false)
	err = f.Client.Create(context.TODO(), sbr, &framework.CleanupOptions{TestContext: ctx, Timeout: time.Second * 5, RetryInterval: time.Second * 1})
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Error(err)
	}

	sbrSecret := &corev1.Secret{}
	if err = f.Client.Get(context.TODO(), namespacedName, sbrSecret); err != nil {
		t.Error(err)
	}
	assert.Equal(t, []byte("test-db"), sbrSecret.Data["DATABASE_DBNAME"], "Name not equal")

	dep := &appsv1.Deployment{}
	namespacedName2 := types.NamespacedName{Namespace: namespace, Name: appResourceRef}
	if err = f.Client.Get(context.TODO(), namespacedName2, dep); err != nil {
		t.Error(err)
	}
	assert.Equal(t, name, dep.Spec.Template.Spec.Containers[0].EnvFrom[0].SecretRef.LocalObjectReference.Name, "secret reference doesn't match")
}
