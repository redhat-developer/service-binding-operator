package mocks

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
