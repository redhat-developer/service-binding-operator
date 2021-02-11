package controllers

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func TestSBRRequestMapperMap(t *testing.T) {
	sbr := &v1alpha1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBinding",
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "mapper-unit",
			Name:      "mapper-unit-sbr",
		},
		Spec: v1alpha1.ServiceBindingSpec{
			Application: &v1alpha1.Application{
				Ref: v1alpha1.Ref{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
					Name:     "mapper-unit-deployment",
				},
			},
			Services: []v1alpha1.Service{
				{
					NamespacedRef: v1alpha1.NamespacedRef{
						Ref: v1alpha1.Ref{
							Group:   "",
							Version: "v1",
							Kind:    "Secret",
							Name:    "mapper-unit-secret",
						},
					},
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
			typeLookup := &ServiceBindingReconciler{restMapper: testutils.BuildTestRESTMapper()}
			mapper := &sbrRequestMapper{
				client:     client,
				typeLookup: typeLookup,
			}
			mappedRequests := mapper.Map(mapObject)
			require.Len(t, mappedRequests, tc.expectedRequestsLen)
		})
	}
}
