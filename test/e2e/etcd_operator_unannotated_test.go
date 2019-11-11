package e2e

import (
	"context"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

func ServiceBindingRequestSetup(t *testing.T, steps []Step) {
	t.Log("Creating a new test context...")
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	ns, f := bootstrapNamespace(t, ctx, true)

	// executing testing steps on operator
	serviceBindingRequestTestSetup(t, ctx, f, ns, steps)
}

func serviceBindingRequestTestSetup(t *testing.T, ctx *framework.TestCtx, f *framework.Framework, ns string, steps []Step) {
	todoCtx := context.TODO()

	name := "e2e-service-binding-request"
	resourceRef := "e2e-db-testing"
	appName := "e2e-application"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "e2e",
	}

	t.Log("Starting end-to-end tests for operator!")

	resourceRefNamespacedName := types.NamespacedName{Namespace: ns, Name: resourceRef}
	deploymentNamespacedName := types.NamespacedName{Namespace: ns, Name: appName}
	serviceBindingRequestNamespacedName := types.NamespacedName{Namespace: ns, Name: name}

	for _, step := range steps {
		switch step {
		case AppStep:
			CreateApp(todoCtx, t, ctx, f, deploymentNamespacedName, matchLabels)
		case SBREtcdStep:
			CreateServiceBindingRequest(
				todoCtx,
				t,
				ctx,
				f,
				serviceBindingRequestNamespacedName,
				resourceRef,
				matchLabels,
				&v1.GroupVersionKind{
					Group:   "etcd.database.coreos.com",
					Version: "v1beta2",
					Kind:    "EtcdCluster",
				},
				true,
			)
		case EtcdClusterStep:
			CreateEtcdCluster(todoCtx, t, ctx, f, resourceRefNamespacedName)
		}
	}

	err := retry(10, 5*time.Second, func() error {
		t.Logf("Inspecting deployment '%s'", deploymentNamespacedName)
		_, err := assertAppDeployed(todoCtx, f, deploymentNamespacedName)
		if err != nil {
			t.Logf("Error on inspecting deployment: '%#v'", err)
		}
		return err
	})
	t.Logf("Deployment: Result after attempts, error: '%#v'", err)
	assert.NoError(t, err)

}
