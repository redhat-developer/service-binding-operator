package binding

import (
	"testing"

	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSpecHandler(t *testing.T) {
	type args struct {
		name            string
		value           string
		service         map[string]interface{}
		resources       []runtime.Object
		expectedData    interface{}
		expectedRawData map[string]interface{}
	}

	assertHandler := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			f := mocks.NewFake(t, "test")

			for _, r := range args.resources {
				f.AddMockResource(r)
			}

			restMapper := testutils.BuildTestRESTMapper()

			handler, err := NewSpecHandler(
				f.FakeDynClient(),
				args.name,
				args.value,
				unstructured.Unstructured{Object: args.service},
				restMapper,
			)
			require.NoError(t, err)
			got, err := handler.Handle()
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, args.expectedData, got.Data, "Data does not match expected")
			require.Equal(t, args.expectedRawData, got.RawData, "RawData does not match expected")
		}
	}

	t.Run("should return password from the resource", assertHandler(args{
		name:  "service.binding/password",
		value: "path={.status.dbCredentials.password}",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"password": "hunter2",
				},
			},
		},
		resources: []runtime.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind: "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},

				Data: map[string]string{
					"username": "AzureDiamond",
					"password": "hunter2",
				},
			},
		},
		expectedData: map[string]interface{}{
			"password": "hunter2",
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"password": "hunter2",
				},
			},
		},
	}))

	t.Run("should return only password from related secret", assertHandler(args{
		name:  "service.binding/password",
		value: "path={.status.dbCredentials},objectType=Secret,sourceValue=password",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "the-secret-resource-name",
			},
		},
		resources: []runtime.Object{
			&corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind: "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},

				Data: map[string][]byte{
					"username": []byte("AzureDiamond"),
					"password": []byte("hunter2"),
				},
			},
		},
		expectedData: map[string]interface{}{
			"password": "hunter2",
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"password": "hunter2",
				},
			},
		},
	}))

	t.Run("should return all data from related secret", assertHandler(args{
		name:  "service.binding",
		value: "path={.status.dbCredentials},objectType=Secret,elementType=map",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "the-secret-resource-name",
			},
		},
		resources: []runtime.Object{
			&corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind: "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},

				Data: map[string][]byte{
					"username": []byte("AzureDiamond"),
					"password": []byte("hunter2"),
				},
			},
		},
		expectedData: map[string]interface{}{
			"password": "hunter2",
			"username": "AzureDiamond",
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"username": "AzureDiamond",
					"password": "hunter2",
				},
			},
		},
	}))

	t.Run("should return only password from related config map", assertHandler(args{
		name:  "service.binding/password",
		value: "path={.status.dbCredentials},objectType=ConfigMap,sourceValue=password",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "the-secret-resource-name",
			},
		},
		resources: []runtime.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind: "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},

				Data: map[string]string{
					"password": "hunter2",
				},
			},
		},
		expectedData: map[string]interface{}{
			"password": "hunter2",
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"password": "hunter2",
				},
			},
		},
	}))

	t.Run("should return all data from related config map", assertHandler(args{
		name:  "service.binding",
		value: "path={.status.dbCredentials},objectType=ConfigMap,elementType=map",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"dbCredentials": "the-secret-resource-name",
			},
		},
		resources: []runtime.Object{
			&corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind: "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "the-namespace",
					Name:      "the-secret-resource-name",
				},

				Data: map[string]string{
					"username": "AzureDiamond",
					"password": "hunter2",
				},
			},
		},
		expectedData: map[string]interface{}{
			"username": "AzureDiamond",
			"password": "hunter2",
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"dbCredentials": map[string]interface{}{
					"username": "AzureDiamond",
					"password": "hunter2",
				},
			},
		},
	}))

	t.Run("should a map with type as key and url as value", assertHandler(args{
		name:  "service.binding",
		value: "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"bootstrap": []interface{}{
					map[string]interface{}{"type": "https", "url": "secure.example.com"},
					map[string]interface{}{"type": "http", "url": "www.example.com"},
				},
			},
		},
		expectedData: map[string]interface{}{
			"bootstrap": map[string]interface{}{
				"https": "secure.example.com",
				"http":  "www.example.com",
			},
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"bootstrap": map[string]interface{}{
					"https": "secure.example.com",
					"http":  "www.example.com",
				},
			},
		},
	}))

	t.Run("should return a map with type as key and url as value", assertHandler(args{
		name:  "service.binding/urls",
		value: "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"bootstrap": []interface{}{
					map[string]interface{}{"type": "https", "url": "secure.example.com"},
					map[string]interface{}{"type": "http", "url": "www.example.com"},
				},
			},
		},
		expectedData: map[string]interface{}{
			"urls": map[string]interface{}{
				"https": "secure.example.com",
				"http":  "www.example.com",
			},
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"urls": map[string]interface{}{
					"https": "secure.example.com",
					"http":  "www.example.com",
				},
			},
		},
	}))

	t.Run("should return a map with type as key and url as value", assertHandler(args{
		name:  "service.binding/urls",
		value: "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"bootstrap": []interface{}{
					map[string]interface{}{"type": "https", "url": "secure.example.com"},
					map[string]interface{}{"type": "http", "url": "www.example.com"},
				},
			},
		},
		expectedData: map[string]interface{}{
			"urls": map[string]interface{}{
				"https": "secure.example.com",
				"http":  "www.example.com",
			},
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"urls": map[string]interface{}{
					"https": "secure.example.com",
					"http":  "www.example.com",
				},
			},
		},
	}))

	t.Run("should return a slice of strings with all urls", assertHandler(args{
		name:  "service.binding",
		value: "path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"bootstrap": []interface{}{
					map[string]interface{}{"type": "https", "url": "secure.example.com"},
					map[string]interface{}{"type": "http", "url": "www.example.com"},
				},
			},
		},
		expectedData: map[string]interface{}{
			"bootstrap": []string{"secure.example.com", "www.example.com"},
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"bootstrap": []string{"secure.example.com", "www.example.com"},
			},
		},
	}))

	t.Run("should return a slice of strings with all urls", assertHandler(args{
		name:  "service.binding/urls",
		value: "path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url",
		service: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace": "the-namespace",
			},
			"status": map[string]interface{}{
				"bootstrap": []interface{}{
					map[string]interface{}{"type": "https", "url": "secure.example.com"},
					map[string]interface{}{"type": "http", "url": "www.example.com"},
				},
			},
		},
		expectedData: map[string]interface{}{
			"urls": []string{"secure.example.com", "www.example.com"},
		},
		expectedRawData: map[string]interface{}{
			"status": map[string]interface{}{
				"urls": []string{"secure.example.com", "www.example.com"},
			},
		},
	}))
}
