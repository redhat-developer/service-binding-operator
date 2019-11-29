package servicebindingrequest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	ustrv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	applicationGVR := schema.GroupVersionResource{"apps", "v1", "deployments"}

	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, "ref", "", applicationGVR, matchLabels)
	f.AddMockedUnstructuredDeployment("ref", matchLabels)

	binder := NewBinder(
		context.TODO(),
		f.FakeClient(),
		f.FakeDynClient(),
		sbr,
		[]string{},
	)

	require.NotNil(t, binder)

	sbrWithResourceRef := f.AddMockedServiceBindingRequest("service-binding-request-with-ref", "ref", "ref", applicationGVR, make(map[string]string))

	binderForSBRWithResourceRef := NewBinder(
		context.TODO(),
		f.FakeClient(),
		f.FakeDynClient(),
		sbrWithResourceRef,
		[]string{},
	)

	require.NotNil(t, binderForSBRWithResourceRef)

	t.Run("search target object by resource name", func(t *testing.T) {
		list, err := binderForSBRWithResourceRef.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
	})

	t.Run("search", func(t *testing.T) {
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
	})

	t.Run("appendEnvFrom", func(t *testing.T) {
		secretName := "secret"
		d := mocks.DeploymentMock("binder", "binder", map[string]string{})
		list := binder.appendEnvFrom(d.Spec.Template.Spec.Containers[0].EnvFrom, secretName)

		require.Equal(t, 1, len(list))
		require.Equal(t, secretName, list[0].SecretRef.Name)
	})

	t.Run("appendEnv", func(t *testing.T) {
		d := mocks.DeploymentMock("binder", "binder", map[string]string{})
		list := binder.appendEnvVar(d.Spec.Template.Spec.Containers[0].Env, "name", "value")
		require.Equal(t, 1, len(list))
		require.Equal(t, "name", list[0].Name)
		require.Equal(t, "value", list[0].Value)
	})

	t.Run("update", func(t *testing.T) {
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))

		updatedObjects, err := binder.update(list)
		require.NoError(t, err)
		require.Len(t, updatedObjects, 1)

		containersPath := []string{"spec", "template", "spec", "containers"}
		containers, found, err := ustrv1.NestedSlice(list.Items[0].Object, containersPath...)
		require.NoError(t, err)
		require.True(t, found)
		require.Len(t, containers, 1)

		c := corev1.Container{}
		u := containers[0].(map[string]interface{})
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, &c)
		require.NoError(t, err)

		// ServiceBindingOperatorChangeTriggerEnvVar should exist to trigger a side effect such as Pod restart when the
		// intermediate secret has been modified
		envVar := getEnvVar(c.Env, ServiceBindingOperatorChangeTriggerEnvVar)
		require.NotNil(t, envVar)
		require.NotEmpty(t, envVar.Value)

		parsedTime, err := time.Parse(time.RFC3339, envVar.Value)
		require.NoError(t, err)
		require.True(t, parsedTime.Before(time.Now()))
	})
}

// getEnvVar returns an EnvVar with given name if exists in the given envVars.
func getEnvVar(envVars []corev1.EnvVar, name string) *corev1.EnvVar {
	for _, v := range envVars {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func TestAppendEnvVar(t *testing.T) {
	envName := "lastbound"
	envList := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  envName,
			Value: "lastboundvalue",
		},
	}

	binder := &Binder{}
	updatedEnvVarList := binder.appendEnvVar(envList, envName, "someothervalue")

	// validate that no new key is added.
	// the existing key should be overwritten with the new value.

	require.Len(t, updatedEnvVarList, 1)
	require.Equal(t, updatedEnvVarList[0].Value, "someothervalue")
}

func TestBinderApplicationName(t *testing.T) {
	ns := "binder"
	name := "service-binding-request"
	applicationGVR := schema.GroupVersionResource{"apps", "v1", "deployments"}

	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, "backingServiceResourceRef", "applicationResourceRef", applicationGVR, nil)
	f.AddMockedUnstructuredDeployment("ref", nil)

	binder := NewBinder(
		context.TODO(),
		f.FakeClient(),
		f.FakeDynClient(),
		sbr,
		[]string{},
	)

	require.NotNil(t, binder)

	t.Run("search by application name", func(t *testing.T) {
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
	})

}

func TestBindingWithDeploymentConfig(t *testing.T) {
	ns := "service-binding-demo-with-deploymentconfig"
	name := "service-binding-request"
	applicationGVR := schema.GroupVersionResource{"apps", "v1", "deploymentconfigs"}

	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, "backingServiceResourceRef", "applicationResourceRef", applicationGVR, nil)
	f.AddMockedUnstructuredDeploymentConfig("ref", nil)

	binder := NewBinder(
		context.TODO(),
		f.FakeClient(),
		f.FakeDynClient(),
		sbr,
		[]string{},
	)

	require.NotNil(t, binder)

	t.Run("deploymentconfig", func(t *testing.T) {
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
		require.Equal(t, "DeploymentConfig", list.Items[0].Object["kind"])
	})

}
