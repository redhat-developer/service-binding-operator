package servicebindingrequest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestBinderNew(t *testing.T) {
	ns := "binder"
	name := "service-binding-request"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}

	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, "ref", matchLabels)
	f.AddMockedUnstructuredDeployment("ref", matchLabels)

	binder := NewBinder(
		context.TODO(),
		f.FakeClient(),
		f.FakeDynClient(),
		sbr,
		[]string{},
	)

	require.NotNil(t, binder)

	t.Run("search", func(t *testing.T) {
		list, err := binder.search()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(list.Items))
	})

	t.Run("appendEnvFrom", func(t *testing.T) {
		secretName := "secret"
		d := mocks.DeploymentMock("binder", "binder", map[string]string{})
		list := binder.appendEnvFrom(d.Spec.Template.Spec.Containers[0].EnvFrom, secretName)

		assert.Equal(t, 1, len(list))
		assert.Equal(t, secretName, list[0].SecretRef.Name)
	})
}
