package servicebindingrequest

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	knativev1 "knative.dev/serving/pkg/apis/serving/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
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

func TestBindingCustomSecretPath(t *testing.T) {
	ns := "custombinder"
	name := "service-binding-request-custom"
	matchLabels := map[string]string{
		"appx": "x",
	}

	f := mocks.NewFake(t, ns)
	sbrSecretPath := f.AddMockedServiceBindingRequest(name, &ns, "ref-custom-podspec", "deployment", deploymentsGVR, matchLabels)
	f.AddMockedUnstructuredDeployment("deployment", matchLabels)

	customSecretPath := "metadata.clusterName"
	sbrSecretPath.Spec.ApplicationSelector.BindingPath = &v1alpha1.BindingPath{
		PodSpecPath: &v1alpha1.PodSpecPath{
			Containers: v1alpha1.DefaultPathToContainers,
			Volumes:    v1alpha1.DefaultPathToVolumes,
		},
		CustomSecretPath: &customSecretPath,
	}
	binderForsbrSecretPath := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbrSecretPath,
		[]string{},
		testutils.BuildTestRESTMapper(),
	)
	require.NotNil(t, binderForsbrSecretPath)

	t.Run("custom secret field path", func(t *testing.T) {
		secretPath := binderForsbrSecretPath.getSecretFieldPath()
		expectedSecretPath := []string{"metadata", "clusterName"}
		require.True(t, reflect.DeepEqual(secretPath, expectedSecretPath))
	})

	t.Run("update custom secret field path ", func(t *testing.T) {
		list, err := binderForsbrSecretPath.search()
		require.NoError(t, err)
		require.Len(t, list.Items, 1)

		updatedDeployment, err := binderForsbrSecretPath.updateSecretField(&list.Items[0])
		require.NoError(t, err)
		require.NotNil(t, updatedDeployment)

		customSecretPathSlice := strings.Split(customSecretPath, ".")

		customSecretInMeta, found, err := unstructured.NestedFieldCopy(list.Items[0].Object, customSecretPathSlice...)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, name, customSecretInMeta)
	})
}

func TestBinderNew(t *testing.T) {
	ns := "binder"
	name := "service-binding-request"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}

	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, "ref", "", deploymentsGVR, matchLabels)
	sbr.Spec.ApplicationSelector.BindingPath = &v1alpha1.BindingPath{
		PodSpecPath: &v1alpha1.PodSpecPath{
			Containers: v1alpha1.DefaultPathToContainers,
			Volumes:    v1alpha1.DefaultPathToVolumes,
		},
	}
	f.AddMockedUnstructuredDeployment("ref", matchLabels)

	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		[]string{},
		testutils.BuildTestRESTMapper(),
	)
	require.NotNil(t, binder)

	sbrWithResourceRef := f.AddMockedServiceBindingRequest(
		"service-binding-request-with-ref",
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
			[]string{},
			testutils.BuildTestRESTMapper(),
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
			[]string{},
			testutils.BuildTestRESTMapper(),
		)

		require.NotNil(t, binder)
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
	})

	t.Run("appendEnvFrom-removeEnvFrom", func(t *testing.T) {
		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			[]string{},
			testutils.BuildTestRESTMapper(),
		)

		require.NotNil(t, binder)
		secretName := "secret"
		d := mocks.DeploymentMock("binder", "binder", map[string]string{})
		envFrom := d.Spec.Template.Spec.Containers[0].EnvFrom

		list := binder.appendEnvFrom(envFrom, secretName)
		require.Equal(t, 1, len(list))
		require.Equal(t, secretName, list[0].SecretRef.Name)

		list = binder.removeEnvFrom(envFrom, secretName)
		require.Equal(t, 0, len(list))
	})

	t.Run("appendEnv", func(t *testing.T) {

		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			[]string{},
			testutils.BuildTestRESTMapper(),
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
			[]string{},
			testutils.BuildTestRESTMapper(),
		)

		require.NotNil(t, binder)
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))

		updatedObjects, err := binder.update(list)
		require.NoError(t, err)
		require.Len(t, updatedObjects, 1)

		// make sure SBR annonation is added
		deployment := appsv1.Deployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(updatedObjects[0].Object, &deployment)
		require.NoError(t, err)

		sbrName, err := getSBRNamespacedNameFromObject(&deployment)
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, sbrName)

		containers, found, err := unstructured.NestedSlice(list.Items[0].Object, binder.getContainersPath()...)
		require.NoError(t, err)
		require.True(t, found)
		require.Len(t, containers, 1)

		c := corev1.Container{}
		u := containers[0].(map[string]interface{})
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, &c)
		require.NoError(t, err)

		// special env-var should exist to trigger a side effect such as Pod restart when the
		// intermediate secret has been modified
		envVar := getEnvVar(c.Env, changeTriggerEnv)
		require.NotNil(t, envVar)
		require.NotEmpty(t, envVar.Value)

		uSecret, err := fakeDynClient.Resource(secretsGVR).Get(name, metav1.GetOptions{})
		require.NoError(t, err)
		s := corev1.Secret{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(uSecret.Object, &s)
		require.NoError(t, err)
		require.Equal(t, s.ObjectMeta.ResourceVersion, envVar.Value)

	})

	t.Run("update with extra modifier present", func(t *testing.T) {
		binder := newBinder(
			context.TODO(),
			f.FakeDynClient(),
			sbr,
			[]string{},
			testutils.BuildTestRESTMapper(),
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
			[]string{},
			testutils.BuildTestRESTMapper(),
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

		sbrName, err := getSBRNamespacedNameFromObject(&deployment)
		require.NoError(t, err)
		require.Equal(t, types.NamespacedName{}, sbrName)

		c := corev1.Container{}
		u := containers[0].(map[string]interface{})
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u, &c)
		require.NoError(t, err)

		// making sure envFrom directive is removed
		require.Empty(t, c.EnvFrom)
		// making sure no volume mounts are present
		require.Nil(t, c.VolumeMounts)
	})

	t.Run("podspec-path-default", func(t *testing.T) {
		containersPath := binder.getContainersPath()
		expectedContainersPath := []string{"spec", "template", "spec", "containers"}
		require.True(t, reflect.DeepEqual(containersPath, expectedContainersPath))

		volumesPath := binder.getVolumesPath()
		expectedVolumesPath := []string{"spec", "template", "spec", "volumes"}
		require.True(t, reflect.DeepEqual(volumesPath, expectedVolumesPath))
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
	name := "service-binding-request"
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, "backingServiceResourceRef", "applicationResourceRef", deploymentsGVR, nil)
	f.AddMockedUnstructuredDeployment("applicationResourceRef", nil)

	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		[]string{},
		testutils.BuildTestRESTMapper(),
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
	f := mocks.NewFake(t, ns)
	sbr := f.AddMockedServiceBindingRequest(name, nil, "backingServiceResourceRef", "applicationResourceRef", deploymentConfigsGVR, nil)
	f.AddMockedUnstructuredDeploymentConfig("applicationResourceRef", nil)

	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		[]string{},
		testutils.BuildTestRESTMapper(),
	)

	require.NotNil(t, binder)

	t.Run("deploymentconfig", func(t *testing.T) {
		list, err := binder.search()
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
		require.Equal(t, "DeploymentConfig", list.Items[0].Object["kind"])
	})

}

func TestBindTwoApplications(t *testing.T) {
	ns := "binder"
	f := mocks.NewFake(t, ns)

	name1 := "service-binding-request-1"
	matchLabels1 := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}
	f.AddMockedUnstructuredDeployment("applicationResourceRef1", matchLabels1)
	sbr1 := f.AddMockedServiceBindingRequest(name1, nil, "backingServiceResourceRef", "", deploymentsGVR, matchLabels1)
	binder1 := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr1,
		[]string{},
		testutils.BuildTestRESTMapper(),
	)
	require.NotNil(t, binder1)

	name2 := "service-binding-request-2"
	matchLabels2 := map[string]string{
		"connects-to": "database",
		"environment": "demo",
	}
	f.AddMockedUnstructuredDeployment("applicationResourceRef2", matchLabels2)
	sbr2 := f.AddMockedServiceBindingRequest(name2, nil, "backingServiceResourceRef", "", deploymentsGVR, matchLabels2)
	binder2 := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr2,
		[]string{},
		testutils.BuildTestRESTMapper(),
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
	name := "service-binding-request"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}

	f := mocks.NewFake(t, ns)
	gvr := knativev1.SchemeGroupVersion.WithResource("services") // Group/Version/Resource for sbr
	sbr := f.AddMockedServiceBindingRequest(name, nil, "", "knative-app", gvr, matchLabels)
	f.AddMockedUnstructuredKnativeService("knative-app", matchLabels)

	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		[]string{},
		testutils.BuildTestRESTMapper(),
	)

	require.NotNil(t, binder)
	require.NotNil(t, binder.modifier)

	t.Run("Knative service contract with service binding operator", func(t *testing.T) {
		list, err := binder.search()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(list.Items))

	})

	ksvc := mocks.KnativeServiceMock(ns, "knative-app-with-rev-name", matchLabels)
	ksvc.Spec.Template.Name = "knative-app-with-rev-name-revision-1"

}

func Test_extraFieldsModifier(t *testing.T) {
	ns := "binder"
	name := "service-binding-request"
	matchLabels := map[string]string{
		"connects-to": "database",
		"environment": "binder",
	}

	f := mocks.NewFake(t, ns)
	deploy := mocks.DeploymentMock(ns, "deployment-fake", matchLabels)
	sbr := mocks.ServiceBindingRequestMock(ns, name, nil, "", deploy.Name, deploymentsGVR, matchLabels)
	binder := newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		[]string{},
		testutils.BuildTestRESTMapper(),
	)

	require.NotNil(t, binder)
	require.Nil(t, binder.modifier)

	gvr := knativev1.SchemeGroupVersion.WithResource("services")
	ksvc := mocks.KnativeServiceMock(ns, "knative-app-with-rev-name", matchLabels)
	sbr = mocks.ServiceBindingRequestMock(ns, name, nil, "", ksvc.Name, gvr, matchLabels)

	binder = newBinder(
		context.TODO(),
		f.FakeDynClient(),
		sbr,
		[]string{},
		testutils.BuildTestRESTMapper(),
	)

	require.NotNil(t, binder)
	require.NotNil(t, binder.modifier)

	t.Run("ksvc revision name is empty", func(t *testing.T) {
		u, err := converter.ToUnstructured(&ksvc)
		require.NoError(t, err)

		err = binder.modifier.ModifyExtraFields(u)
		require.NoError(t, err)

		var modified knativev1.Service
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &modified)
		require.NoError(t, err)
		assert.Equal(t, ksvc, modified)
	})

	t.Run("ksvc revision name is not empty", func(t *testing.T) {
		ksvc.Spec.Template.Name = fmt.Sprintf("%s-%s", ksvc.Name, "rev-1")

		u, err := converter.ToUnstructured(&ksvc)
		require.NoError(t, err)

		err = binder.modifier.ModifyExtraFields(u)
		require.NoError(t, err)

		var modified knativev1.Service
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &modified)
		require.NoError(t, err)
		assert.Equal(t, "", modified.Spec.Template.Name)

		ksvc.Spec.Template.Name = ""
		// the rest fields shoud not be modified
		assert.Equal(t, ksvc, modified)
	})

}
