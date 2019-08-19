package servicebindingrequest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

	t.Run("appendEnv", func(t *testing.T) {
		d := mocks.DeploymentMock("binder", "binder", map[string]string{})
		list := binder.appendEnvVar(d.Spec.Template.Spec.Containers[0].Env, "name", "value")
		assert.Equal(t, 1, len(list))
		assert.Equal(t, "name", list[0].Name)
		assert.Equal(t, "value", list[0].Value)
	})

	t.Run("update", func(t *testing.T) {
		list, err := binder.search()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(list.Items))

		updatedObjects, err := binder.update(list)
		assert.NoError(t, err)
		assert.Len(t, updatedObjects, 1)

		containersPath := []string{"spec", "template", "spec", "containers"}
		containers, found, err := ustrv1.NestedSlice(list.Items[0].Object, containersPath...)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Len(t, containers, 1)

		c := corev1.Container{}
		u := containers[0].(map[string]interface{})
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, &c)
		assert.NoError(t, err)
		assert.NotEmpty(t, c.Env[0].Value)

		parsedTime, err := time.Parse(time.RFC3339, c.Env[0].Value)
		assert.NoError(t, err)
		assert.True(t, parsedTime.Before(time.Now()))
	})
}
