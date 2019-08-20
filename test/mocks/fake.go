package mocks

import (
	"testing"

	pgapis "github.com/baijum/postgresql-operator/pkg/apis"
	pgv1alpha1 "github.com/baijum/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// Fake defines all the elements to fake a kubernetes api client.
type Fake struct {
	t    *testing.T       // testing instance
	ns   string           // namespace
	S    *runtime.Scheme  // runtime client scheme
	objs []runtime.Object // all fake objects
}

// AddMockedServiceBindingRequest add mocked object from ServiceBindingRequestMock.
func (f *Fake) AddMockedServiceBindingRequest(name, ref string, matchLabels map[string]string) *v1alpha1.ServiceBindingRequest {
	f.S.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.ServiceBindingRequest{})
	sbr := ServiceBindingRequestMock(f.ns, name, ref, matchLabels)
	f.objs = append(f.objs, sbr)
	return sbr
}

// AddMockedUnstructuredCSV add mocked unstructured CSV.
func (f *Fake) AddMockedUnstructuredCSV(name string) {
	require.Nil(f.t, olmv1alpha1.AddToScheme(f.S))
	csv, err := UnstructuredClusterServiceVersionMock(f.ns, name)
	require.Nil(f.t, err)
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, csv)
}

// AddMockedCSVList add mocked object from ClusterServiceVersionListMock.
func (f *Fake) AddMockedCSVList(name string) {
	require.Nil(f.t, olmv1alpha1.AddToScheme(f.S))
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, ClusterServiceVersionListMock(f.ns, name))
}

// AddMockedCSVWithVolumeMountList add mocked object from ClusterServiceVersionListVolumeMountMock.
func (f *Fake) AddMockedCSVWithVolumeMountList(name string) {
	require.Nil(f.t, olmv1alpha1.AddToScheme(f.S))
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, ClusterServiceVersionListVolumeMountMock(f.ns, name))
}

// AddMockedUnstructuredCSVWithVolumeMount same than AddMockedCSVWithVolumeMountList but using
// unstructured object.
func (f *Fake) AddMockedUnstructuredCSVWithVolumeMount(name string) {
	require.Nil(f.t, olmv1alpha1.AddToScheme(f.S))
	csv, err := UnstructuredClusterServiceVersionVolumeMountMock(f.ns, name)
	require.Nil(f.t, err)
	f.S.AddKnownTypes(olmv1alpha1.SchemeGroupVersion, &olmv1alpha1.ClusterServiceVersion{})
	f.objs = append(f.objs, csv)
}

// AddMockedDatabaseCRList add mocked object from DatabaseCRListMock.
func (f *Fake) AddMockedDatabaseCRList(ref string) {
	require.Nil(f.t, pgapis.AddToScheme(f.S))
	f.S.AddKnownTypes(pgv1alpha1.SchemeGroupVersion, &pgv1alpha1.Database{})
	f.objs = append(f.objs, DatabaseCRListMock(f.ns, ref))
}

// AddMockedUnstructuredDeployment add mocked object from UnstructuredDeploymentMock.
func (f *Fake) AddMockedUnstructuredDeployment(name string, matchLabels map[string]string) {
	require.Nil(f.t, appsv1.AddToScheme(f.S))
	d, err := UnstructuredDeploymentMock(f.ns, name, matchLabels)
	require.Nil(f.t, err)
	f.S.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.Deployment{})
	f.objs = append(f.objs, d)
}

// AddMockedSecret add mocked object from SecretMock.
func (f *Fake) AddMockedSecret(name string) {
	f.objs = append(f.objs, SecretMock(f.ns, name))
}

// FakeClient returns fake structured api client.
func (f *Fake) FakeClient() client.Client {
	return fake.NewFakeClientWithScheme(f.S, f.objs...)
}

// FakeDynClient returns fake dynamic api client.
func (f *Fake) FakeDynClient() dynamic.Interface {
	return fakedynamic.NewSimpleDynamicClient(f.S, f.objs...)
}

// NewFake instantiate Fake type.
func NewFake(t *testing.T, ns string) *Fake {
	return &Fake{t: t, ns: ns, S: scheme.Scheme}
}
