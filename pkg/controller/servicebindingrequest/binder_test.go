package servicebindingrequest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

var binderFakeClient client.Client
var binderFakeDynClient dynamic.Interface

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
}

func TestBinderAppendEnvFrom(t *testing.T) {
	binder := &Binder{}
	secretName := "secret"
	d := mocks.DeploymentMock("binder", "binder", map[string]string{})
	list := binder.appendEnvFrom(d.Spec.Template.Spec.Containers[0].EnvFrom, secretName)

	assert.Equal(t, 1, len(list))
	assert.Equal(t, secretName, list[0].SecretRef.Name)
}
