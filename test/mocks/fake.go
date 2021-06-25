package mocks

import (
	v1alpha12 "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"testing"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

// Fake defines all the elements to fake a kubernetes api client.
type Fake struct {
	t    *testing.T       // testing instance
	ns   string           // namespace
	S    *runtime.Scheme  // runtime client scheme
	objs []runtime.Object // all fake objects
}

// AddMockedServiceBinding add mocked object from ServiceBindingMock.
func (f *Fake) AddMockedServiceBinding(
	name string,
	backingServiceNamespace *string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
) *v1alpha12.ServiceBinding {
	f.S.AddKnownTypes(v1alpha12.GroupVersion, &v1alpha12.ServiceBinding{})
	sbr := ServiceBindingMock(f.ns, name, backingServiceNamespace, backingServiceResourceRef, applicationResourceRef, applicationGVR, matchLabels)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedServiceBindingWithUnannotated add mocked object from ServiceBindingMock with DetectBindingResources.
func (f *Fake) AddMockedServiceBindingWithUnannotated(
	name string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
) *v1alpha12.ServiceBinding {
	f.S.AddKnownTypes(v1alpha12.GroupVersion, &v1alpha12.ServiceBinding{})
	sbr := ServiceBindingMock(f.ns, name, nil, backingServiceResourceRef, applicationResourceRef, applicationGVR, matchLabels)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedUnstructuredServiceBindingWithoutApplication creates a mock ServiceBinding object
func (f *Fake) AddMockedUnstructuredServiceBindingWithoutApplication(
	name string,
	backingServiceResourceRef string,
) *unstructured.Unstructured {
	f.S.AddKnownTypes(v1alpha12.GroupVersion, &v1alpha12.ServiceBinding{})
	var emptyGVR = schema.GroupVersionResource{}
	sbr, err := UnstructuredServiceBindingMock(f.ns, name, backingServiceResourceRef, "", emptyGVR, nil)
	require.NoError(f.t, err)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedUnstructuredServiceBindingWithoutApplication creates a mock ServiceBinding object
func (f *Fake) AddMockedUnstructuredServiceBindingWithoutService(
	name string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
) *unstructured.Unstructured {
	f.S.AddKnownTypes(v1alpha12.GroupVersion, &v1alpha12.ServiceBinding{})
	sbr, err := UnstructuredServiceBindingMock(f.ns, name, "", applicationResourceRef, applicationGVR, nil)
	require.NoError(f.t, err)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedUnstructuredServiceBinding creates a mock ServiceBinding object
func (f *Fake) AddMockedUnstructuredServiceBinding(
	name string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	applicationGVR schema.GroupVersionResource,
	matchLabels map[string]string,
) *unstructured.Unstructured {
	f.S.AddKnownTypes(v1alpha12.GroupVersion, &v1alpha12.ServiceBinding{})
	sbr, err := UnstructuredServiceBindingMock(f.ns, name, backingServiceResourceRef, applicationResourceRef, applicationGVR, matchLabels)
	require.NoError(f.t, err)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedUnstructuredCSV add mocked unstructured CSV.
func (f *Fake) AddMockedUnstructuredCSV(name string) {
	require.NoError(f.t, olmv1alpha1.AddToScheme(f.S))
	csv, err := UnstructuredClusterServiceVersionMock(f.ns, name)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, csv)
}

// AddMockedCSVList add mocked object from ClusterServiceVersionListMock.
func (f *Fake) AddMockedCSVList(name string) {
	require.NoError(f.t, olmv1alpha1.AddToScheme(f.S))
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, ClusterServiceVersionListMock(f.ns, name))
}

// AddMockedCSVWithVolumeMountList add mocked object from ClusterServiceVersionListVolumeMountMock.
func (f *Fake) AddMockedCSVWithVolumeMountList(name string) {
	require.NoError(f.t, olmv1alpha1.AddToScheme(f.S))
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, ClusterServiceVersionListVolumeMountMock(f.ns, name))
}

// AddMockedUnstructuredCSVWithVolumeMount same than AddMockedCSVWithVolumeMountList but using
// unstructured object.
func (f *Fake) AddMockedUnstructuredCSVWithVolumeMount(name string) {
	require.NoError(f.t, olmv1alpha1.AddToScheme(f.S))
	csv, err := UnstructuredClusterServiceVersionVolumeMountMock(f.ns, name)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, csv)
}

// AddMockedDatabaseCR add mocked object from DatabaseCRMock.
func (f *Fake) AddMockedDatabaseCR(ref string, namespace string) runtime.Object {
	mock := UnstructuredDatabaseCRMock(namespace, ref)
	f.S.AddKnownTypeWithName(mock.GroupVersionKind(), &unstructured.Unstructured{})
	f.objs = append(f.objs, mock)
	return mock
}

func (f *Fake) AddMockedUnstructuredDatabaseCR(ref string) {
	d := UnstructuredDatabaseCRMock(f.ns, ref)
	f.S.AddKnownTypeWithName(d.GroupVersionKind(), &unstructured.Unstructured{})
	f.objs = append(f.objs, d)
}

// AddMockedUnstructuredDeploymentConfig adds mocked object from UnstructuredDeploymentConfigMock.
func (f *Fake) AddMockedUnstructuredDeploymentConfig(name string, matchLabels map[string]string) {
	d := UnstructuredDeploymentConfigMock(f.ns, name)
	f.S.AddKnownTypeWithName(schema.GroupVersionKind{Group: "apps.openshift.io", Version: "v1", Kind: "DeploymentConfig"}, &unstructured.Unstructured{})
	f.objs = append(f.objs, d)
}

// AddMockedUnstructuredDeployment add mocked object from UnstructuredDeploymentMock.
func (f *Fake) AddMockedUnstructuredDeployment(name string, matchLabels map[string]string) *unstructured.Unstructured {
	require.NoError(f.t, appsv1.AddToScheme(f.S))
	d, err := UnstructuredDeploymentMock(f.ns, name, matchLabels)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.Deployment{})
	f.objs = append(f.objs, d)
	return d
}

// AddMockedUnstructuredKnativeService add mocked object from UnstructuredKnativeService.
func (f *Fake) AddMockedUnstructuredKnativeService(name string, matchLabels map[string]string) {
	d := UnstructuredKnativeServiceMock(f.ns, name, matchLabels)
	f.S.AddKnownTypeWithName(d.GroupVersionKind(), &unstructured.Unstructured{})
	f.objs = append(f.objs, d)
}

func (f *Fake) AddMockedUnstructuredDatabaseCRD() *unstructured.Unstructured {
	require.NoError(f.t, apiextensionv1beta1.AddToScheme(f.S))
	c, err := UnstructuredDatabaseCRDMock(f.ns)
	require.NoError(f.t, err)
	f.S.AddKnownTypes(apiextensionv1beta1.SchemeGroupVersion, &apiextensionv1beta1.CustomResourceDefinition{})
	f.objs = append(f.objs, c)
	return c
}

func (f *Fake) AddMockedUnstructuredPostgresDatabaseCR(ref string) *unstructured.Unstructured {
	d, err := UnstructuredPostgresDatabaseCRMock(f.ns, ref)
	require.NoError(f.t, err)
	f.objs = append(f.objs, d)
	return d
}

// AddMockedUnstructuredSecret add mocked object from SecretMock.
func (f *Fake) AddMockedUnstructuredSecret(name string) *unstructured.Unstructured {
	s, err := UnstructuredSecretMock(f.ns, name)
	require.NoError(f.t, err)
	f.objs = append(f.objs, s)
	return s
}

// AddMockedUnstructuredSecret add mocked object from SecretMock. This secret is created with a resourceVersion
func (f *Fake) AddMockedUnstructuredSecretRV(name string) *unstructured.Unstructured {
	s, err := UnstructuredSecretMockRV(f.ns, name)
	require.NoError(f.t, err)
	f.objs = append(f.objs, s)
	return s
}

// AddNamespacedMockedSecret add mocked object from SecretMock in a namespace
// which isn't necessarily same as that of the ServiceBinding namespace.
func (f *Fake) AddNamespacedMockedSecret(name string, namespace string, data map[string][]byte) {
	f.objs = append(f.objs, SecretMock(namespace, name, data))
}

// AddMockedUnstructuredConfigMap add mocked object from ConfigMapMock.
func (f *Fake) AddMockedUnstructuredConfigMap(name string) {
	mock := ConfigMapMock(f.ns, name)
	uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(mock)
	require.NoError(f.t, err)
	f.objs = append(f.objs, &unstructured.Unstructured{Object: uObj})
}

func (f *Fake) AddMockResource(resource runtime.Object) {
	f.objs = append(f.objs, resource)
}

// FakeDynClient returns fake dynamic api client.
func (f *Fake) FakeDynClient() *fakedynamic.FakeDynamicClient {
	return fakedynamic.NewSimpleDynamicClient(f.S, f.objs...)
}

// NewFake instantiate Fake type.
func NewFake(t *testing.T, ns string) *Fake {
	return &Fake{t: t, ns: ns, S: scheme.Scheme}
}
