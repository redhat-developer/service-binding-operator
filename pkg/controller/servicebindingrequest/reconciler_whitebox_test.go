package servicebindingrequest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestSetRetriggerBinding(t *testing.T) {
	ctx := context.TODO()
	resourceRef := "test-retrigger"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "reconciler",
	}

	f := mocks.NewFake(t, reconcilerNs)

	f.AddMockedServiceBindingRequest(reconcilerName, resourceRef, matchLabels)
	fakeClient := f.FakeClient()

	t.Run("reconcile-using-trigger-binding", func(t *testing.T) {
		namespacedName := types.NamespacedName{Namespace: reconcilerNs, Name: reconcilerName}
		sbr := v1alpha1.ServiceBindingRequest{}
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &sbr))

		reconciler := &Reconciler{client: fakeClient, dynClient: f.FakeDynClient(), scheme: f.S}
		reconciler.setTriggerRebindingFlag(ctx, &sbr)

		require.NoError(t, fakeClient.Get(ctx, namespacedName, &sbr))

		// If nothing was set initially, nothing will be set.
		require.Nil(t, sbr.Spec.TriggerRebinding)

		// lets as True initially and see what happens
		triggerTrue := true
		sbr.Spec.TriggerRebinding = &triggerTrue
		require.NoError(t, fakeClient.Update(ctx, &sbr))

		reconciler.setTriggerRebindingFlag(ctx, &sbr)
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &sbr))
		require.False(t, *sbr.Spec.TriggerRebinding)

		// lets as False initially and see what happens
		triggerFalse := false
		sbr.Spec.TriggerRebinding = &triggerFalse
		require.NoError(t, fakeClient.Update(ctx, &sbr))

		reconciler.setTriggerRebindingFlag(ctx, &sbr)
		require.NoError(t, fakeClient.Get(ctx, namespacedName, &sbr))

		// remains false
		require.False(t, *sbr.Spec.TriggerRebinding)

	})
}
