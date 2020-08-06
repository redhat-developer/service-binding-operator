package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestSBRRequestMapperMap(t *testing.T) {
	sbr := &v1alpha1.ServiceBindingRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingRequest",
			APIVersion: "apps.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "mapper-unit",
			Name:      "mapper-unit-sbr",
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			ApplicationSelector: v1alpha1.ApplicationSelector{
				GroupVersionResource: metav1.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				ResourceRef: "mapper-unit-deployment",
			},
			BackingServiceSelectors: &[]v1alpha1.BackingServiceSelector{
				{
					GroupVersionKind: metav1.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Secret",
					},
					ResourceRef: "mapper-unit-secret",
				},
			},
		},
	}

	type testCase struct {
		description         string
		expectedRequestsLen int
		buildMapObjectFn    func(*mocks.Fake) handler.MapObject
		buildFakeFn         func() *mocks.Fake
	}

	testCases := []testCase{
		{
			description: "no service bindings declared in namespace",
			buildFakeFn: func() *mocks.Fake {
				f := mocks.NewFake(t, reconcilerNs)
				return f
			},
			buildMapObjectFn: func(f *mocks.Fake) handler.MapObject {
				return handler.MapObject{
					Meta: &metav1.ObjectMeta{
						Namespace: "mapper-unit",
						Name:      "mapper-unit-secret",
					},
					Object: &corev1.Secret{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "Secret",
						},
					},
				}
			},
			expectedRequestsLen: 0,
		},
		{
			description: "resource declared as application of a service binding",
			buildFakeFn: func() *mocks.Fake {
				f := mocks.NewFake(t, reconcilerNs)
				uSbr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sbr)
				require.NoError(t, err)
				f.AddMockResource(&unstructured.Unstructured{Object: uSbr})
				return f
			},
			buildMapObjectFn: func(f *mocks.Fake) handler.MapObject {
				return handler.MapObject{
					Meta: &metav1.ObjectMeta{
						Namespace: "mapper-unit",
						Name:      "mapper-unit-deployment",
					},
					Object: &appsv1.Deployment{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "apps/v1",
							Kind:       "Deployment",
						},
					},
				}
			},
			expectedRequestsLen: 1,
		},
		{
			description: "resource declared as service of a service binding",
			buildFakeFn: func() *mocks.Fake {
				f := mocks.NewFake(t, reconcilerNs)
				uSbr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sbr)
				require.NoError(t, err)
				f.AddMockResource(&unstructured.Unstructured{Object: uSbr})
				return f
			},
			buildMapObjectFn: func(f *mocks.Fake) handler.MapObject {
				return handler.MapObject{
					Meta: &metav1.ObjectMeta{
						Namespace: "mapper-unit",
						Name:      "mapper-unit-secret",
					},
					Object: &corev1.Secret{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "Secret",
						},
					},
				}
			},
			expectedRequestsLen: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			f := tc.buildFakeFn()
			mapObject := tc.buildMapObjectFn(f)
			client := f.FakeDynClient()
			restMapper := testutils.BuildTestRESTMapper()
			mapper := &sbrRequestMapper{
				client:     client,
				restMapper: restMapper,
			}
			mappedRequests := mapper.Map(mapObject)
			require.Len(t, mappedRequests, tc.expectedRequestsLen)
		})
	}
}
