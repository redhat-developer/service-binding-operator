package controllers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
)

func init() {
	log.SetLogger(zap.New(zap.UseDevMode((true))))
}

func TestBindingCustomSecretPath(t *testing.T) {
	ns := "custombinder"
	name := "service-binding-request-custom"
	matchLabels := map[string]string{
		"appx": "x",
	}

	f := mocks.NewFake(t, ns)
	sbrSecretPath := f.AddMockedServiceBinding(name, &ns, "ref-custom-podspec", "deployment", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredDeployment("deployment", matchLabels)

	customSecretPath := "metadata.clusterName"
	sbrSecretPath.Spec.Application.BindingPath = &v1alpha1.BindingPath{
		SecretPath: customSecretPath,
	}
	binderForsbrSecretPath := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbrSecretPath,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)
	require.NotNil(t, binderForsbrSecretPath)

	t.Run("custom secret field path", func(t *testing.T) {
		secretPath := binderForsbrSecretPath.getSecretFieldPath()
		expectedSecretPath := []string{"metadata", "clusterName"}
		require.Equal(t, expectedSecretPath, secretPath)
	})

	t.Run("update custom secret field path ", func(t *testing.T) {
		list, err := binderForsbrSecretPath.search()
		require.NoError(t, err)
		require.Len(t, list.Items, 1)

		obj := list.Items[0]
		err = binderForsbrSecretPath.updateSecretField(&obj)
		require.NoError(t, err)
		customSecretPathSlice := strings.Split(customSecretPath, ".")

		customSecretInMeta, found, err := unstructured.NestedFieldCopy(obj.Object, customSecretPathSlice...)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, binderForsbrSecretPath.sbr.Status.Secret, customSecretInMeta)
	})
}

func TestBinderNew(t *testing.T) {
	ns := "binder"
	name := "service-binding"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}

	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBinding(name, nil, "ref", "", deploymentsGVR, matchLabels)
	ensureDefaults(sbr.Spec.Application)
	f.AddMockedUnstructuredDeployment("ref", matchLabels)

	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)
	require.NotNil(t, binder)

	sbrWithResourceRef := f.AddMockedServiceBinding(
		"service-binding-with-ref",
		nil,
		"ref",
		"ref",
		deploymentsGVR,
		matchLabels,
	)

	f.AddMockedUnstructuredSecretRV(name)
	fakeDynClient := f.FakeDynClient()

	t.Run("search-using-resourceref", func(t *testing.T) {
		binderForSBRWithResourceRef := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbrWithResourceRef,
			&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
		)

		require.NotNil(t, binderForSBRWithResourceRef)
		list, err := binderForSBRWithResourceRef.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
	})

	t.Run("search", func(t *testing.T) {
		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
		)

		require.NotNil(t, binder)
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
	})

	t.Run("updateEnvFrom-removeEnvFrom-with-empty-envFrom", func(t *testing.T) {
		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
		)

		require.NotNil(t, binder)
		secretName := "secret"
		d := mocks.DeploymentMock("binder", "binder", map[string]string{})
		envFrom := d.Spec.Template.Spec.Containers[0].EnvFrom

		list := binder.updateEnvFromList(envFrom, secretName)
		require.Equal(t, 1, len(list))
		require.Equal(t, secretName, list[0].SecretRef.Name)

		list = binder.removeEnvFrom(envFrom, secretName)
		require.Equal(t, 0, len(list))
	})

	t.Run("updateEnvFrom-removeEnvFrom-with-configMapRef", func(t *testing.T) {
		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
		)

		require.NotNil(t, binder)
		secretName := "secret"
		configMapName := "configmap"
		d := mocks.DeploymentMock("binder", "binder", map[string]string{})
		envFrom := d.Spec.Template.Spec.Containers[0].EnvFrom
		envFrom = append(envFrom, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		})

		list := binder.updateEnvFromList(envFrom, secretName)
		require.Equal(t, 2, len(list))
		require.Equal(t, configMapName, list[0].ConfigMapRef.Name)
		require.Equal(t, secretName, list[1].SecretRef.Name)

		list = binder.removeEnvFrom(envFrom, secretName)
		require.Equal(t, 0, len(list))
	})

	t.Run("appendEnv", func(t *testing.T) {

		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
		)

		require.NotNil(t, binder)
		d := mocks.DeploymentMock("binder", "binder", map[string]string{})
		list := binder.appendEnvVar(d.Spec.Template.Spec.Containers[0].Env, "name", "value")
		require.Equal(t, 1, len(list))
		require.Equal(t, "name", list[0].Name)
		require.Equal(t, "value", list[0].Value)
	})

	t.Run("update", func(t *testing.T) {

		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
		)

		require.NotNil(t, binder)
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))

		updatedObjects, err := binder.update(list)
		require.NoError(t, err)
		require.Len(t, updatedObjects, 1)

		containers, found, err := unstructured.NestedSlice(updatedObjects[0].Object, binder.getContainersPath()...)
		require.NoError(t, err)
		require.True(t, found)
		require.Len(t, containers, 1)

		c := corev1.Container{}
		u := containers[0].(map[string]interface{})
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, &c)
		require.NoError(t, err)

		uSecret, err := fakeDynClient.Resource(secretsGVR).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{})
		require.NoError(t, err)
		s := corev1.Secret{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(uSecret.Object, &s)
		require.NoError(t, err)
	})

	t.Run("update with extra modifier present", func(t *testing.T) {
		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
		)
		// test binder with extra modifier present
		ch := make(chan struct{})
		binder.modifier = extraFieldsModifierFunc(func(u *unstructured.Unstructured) error {
			close(ch)
			return nil
		})

		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))

		updatedObjects, err := binder.update(list)
		require.NoError(t, err)
		require.Len(t, updatedObjects, 1)
		<-ch

		list, err = binder.search()
		require.NoError(t, err)
		// call another update as object is already updated, modifier func should not be called
		updatedObjects, err = binder.update(list)
		require.NoError(t, err)
		require.Len(t, updatedObjects, 0)
	})

	t.Run("remove", func(t *testing.T) {

		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
		)

		require.NotNil(t, binder)
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))

		updatedObjects, err := binder.update(list)
		require.NoError(t, err)
		require.Len(t, updatedObjects, 1)

		err = binder.remove(list)
		require.NoError(t, err)

		containers, found, err := unstructured.NestedSlice(list.Items[0].Object, binder.getContainersPath()...)
		require.NoError(t, err)
		require.True(t, found)
		require.Len(t, containers, 1)

		// make sure SBR annonation is removed
		deployment := appsv1.Deployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(list.Items[0].Object, &deployment)
		require.NoError(t, err)

		c := corev1.Container{}
		u := containers[0].(map[string]interface{})
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, &c)
		require.NoError(t, err)

		// making sure envFrom directive is removed
		require.Empty(t, c.EnvFrom)
		// making sure no volume mounts are present
		require.Nil(t, c.VolumeMounts)
	})
}

func TestBinderAppendEnvVar(t *testing.T) {
	envName := "lastbound"
	envList := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  envName,
			Value: "lastboundvalue",
		},
	}

	b := &binder{}
	updatedEnvVarList := b.appendEnvVar(envList, envName, "someothervalue")

	// validate that no new key is added.
	// the existing key should be overwritten with the new value.

	require.Len(t, updatedEnvVarList, 1)
	require.Equal(t, updatedEnvVarList[0].Value, "someothervalue")
}

func TestBinderApplicationName(t *testing.T) {
	ns := "binder"
	name := "service-binding"
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBinding(name, nil, "backingServiceResourceRef", "applicationResourceRef", deploymentsGVR, nil)
	f.AddMockedUnstructuredDeployment("applicationResourceRef", nil)

	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
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
	name := "service-binding"
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBinding(name, nil, "backingServiceResourceRef", "applicationResourceRef", deploymentConfigsGVR, nil)
	f.AddMockedUnstructuredDeploymentConfig("applicationResourceRef", nil)

	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)

	require.NotNil(t, binder)

	t.Run("deploymentconfig", func(t *testing.T) {
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
		require.Equal(t, "DeploymentConfig", list.Items[0].Object["kind"])
	})

}

func TestBindEmptyApplication(t *testing.T) {
	ns := "binder"
	f := mocks.NewFake(t, ns)

	name := "service-binding-empty-application"
	sbr := f.AddMockedServiceBinding(name, nil, "backingServiceResourceRef", "", deploymentsGVR, nil)

	binder1 := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)
	require.NotNil(t, binder1)

	t.Run("Binder search should return errEmptyApplication error for service binding with empty application", func(t *testing.T) {
		applicationList, err := binder1.search()
		assert.Error(t, errEmptyApplication, err)
		assert.Empty(t, applicationList)
	})
}

func TestBindTwoApplications(t *testing.T) {
	ns := "binder"
	f := mocks.NewFake(t, ns)

	name1 := "service-binding-1"
	matchLabels1 := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}
	f.AddMockedUnstructuredDeployment("applicationResourceRef1", matchLabels1)
	sbr1 := f.AddMockedServiceBinding(name1, nil, "backingServiceResourceRef", "", deploymentsGVR, matchLabels1)
	binder1 := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr1,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)
	require.NotNil(t, binder1)

	name2 := "service-binding-2"
	matchLabels2 := map[string]string{
		"connects-to": "database",
		"environment": "demo",
	}
	f.AddMockedUnstructuredDeployment("applicationResourceRef2", matchLabels2)
	sbr2 := f.AddMockedServiceBinding(name2, nil, "backingServiceResourceRef", "", deploymentsGVR, matchLabels2)
	binder2 := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr2,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)
	require.NotNil(t, binder2)

	t.Run("two applications with one backing service", func(t *testing.T) {
		list1, err := binder1.search()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(list1.Items))

		list2, err := binder2.search()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(list2.Items))
	})
}

func TestKnativeServicesContractWithBinder(t *testing.T) {
	ns := "binder"
	name := "service-binding"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}

	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBinding(name, nil, "", "knative-app", knativeServiceGVR, matchLabels)
	f.AddMockedUnstructuredKnativeService("knative-app", matchLabels)

	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)

	require.NotNil(t, binder)
	require.NotNil(t, binder.modifier)

	t.Run("Knative service contract with service binding operator", func(t *testing.T) {
		list, err := binder.search()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(list.Items))

	})
}

func Test_extraFieldsModifier(t *testing.T) {
	ns := "binder"
	name := "service-binding"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}

	f := mocks.NewFake(t, ns)
	deploy := mocks.DeploymentMock(ns, "deployment-fake", matchLabels)
	sbr := mocks.ServiceBindingMock(ns, name, nil, "", deploy.Name, deploymentsGVR, matchLabels)
	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)

	require.NotNil(t, binder)
	require.Nil(t, binder.modifier)

	ksvc := mocks.UnstructuredKnativeServiceMock(ns, "knative-app-with-rev-name", matchLabels)
	sbr = mocks.ServiceBindingMock(ns, name, nil, "", ksvc.GetName(), knativeServiceGVR, matchLabels)

	binder = newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		&ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()},
	)

	require.NotNil(t, binder)
	require.NotNil(t, binder.modifier)

	t.Run("ksvc revision name is empty", func(t *testing.T) {
		path := []string{"spec", "template", "metadata", "name"}
		err := binder.modifier.ModifyExtraFields(ksvc)
		require.NoError(t, err)

		_, found, err := unstructured.NestedString(ksvc.Object, path...)
		require.NoError(t, err)
		require.False(t, found)
	})

	t.Run("ksvc revision name is not empty", func(t *testing.T) {
		path := []string{"spec", "template", "metadata", "name"}
		err := unstructured.SetNestedField(ksvc.Object, fmt.Sprintf("%s-%s", ksvc.GetName(), "rev-1"), path...)
		require.NoError(t, err)

		err = binder.modifier.ModifyExtraFields(ksvc)
		require.NoError(t, err)

		_, found, err := unstructured.NestedString(ksvc.Object, path...)
		require.NoError(t, err)
		require.False(t, found)

	})

}
