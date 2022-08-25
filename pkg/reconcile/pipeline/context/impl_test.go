package context

import (
	"context"
	"encoding/base64"
	"encoding/json"
	e "errors"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis"
	bindingapi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/apis/spec/v1beta1"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes/mocks"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	pipelinemocks "github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/mocks"
	corev1 "k8s.io/api/core/v1"
	v1apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	fakeauth "k8s.io/client-go/kubernetes/typed/authorization/v1/fake"
	"k8s.io/client-go/testing"
)

var _ = Describe("Context", func() {

	var (
		mockCtrl   *gomock.Controller
		typeLookup *mocks.MockK8STypeLookup
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		typeLookup = mocks.NewMockK8STypeLookup(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Applications", func() {

		DescribeTable("should return slice of size 1 if application is specified", func(bindingPath *bindingapi.BindingPath) {
			ref := bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
				BindingPath: bindingPath,
			}

			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: ref,
				},
			}
			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			u := &unstructured.Unstructured{}
			u.SetName("app1")
			u.SetNamespace(sb.Namespace)
			u.SetGroupVersionKind(schema.GroupVersionKind{Group: "app", Version: "v1", Kind: "Foo"})
			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			applications, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(applications).To(HaveLen(1))
			Expect(applications[0].Resource()).To(Equal(u))
		},
			Entry("no binding path specified", nil),
			Entry("binding path specified", &bindingapi.BindingPath{ContainersPath: "foo.bar"}),
		)
		DescribeTable("should return slice of size 2 if 2 applications are specified through label seclector", func(bindingPath *bindingapi.BindingPath) {
			ls := &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
				},
				LabelSelector: ls,
				BindingPath:   bindingPath,
			}

			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: ref,
				},
			}
			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			u1 := &unstructured.Unstructured{}
			u1.SetName("app1")
			u1.SetNamespace(sb.Namespace)
			u1.SetGroupVersionKind(schema.GroupVersionKind{Group: "app", Version: "v1", Kind: "Foo"})
			u1.SetLabels(map[string]string{"env": "prod"})

			u2 := &unstructured.Unstructured{}
			u2.SetName("app2")
			u2.SetNamespace(sb.Namespace)
			u2.SetGroupVersionKind(schema.GroupVersionKind{Group: "app", Version: "v1", Kind: "Foo"})
			u2.SetLabels(map[string]string{"env": "prod"})

			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u1, u2)
			authClient := &fakeauth.FakeAuthorizationV1{}
			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			applications, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(applications).To(HaveLen(2))

			Expect(applications[0].Resource().GetName()).NotTo(Equal(applications[1].Resource().GetName()))
			Expect(applications[0].Resource()).Should(BeElementOf(u1, u2))
			Expect(applications[1].Resource()).Should(BeElementOf(u1, u2))
		},
			Entry("no binding path specified", nil),
			Entry("binding path specified", &bindingapi.BindingPath{ContainersPath: "foo.bar"}),
		)
		DescribeTable("should return slice of size 0 if no application is matching through label seclector", func(bindingPath *bindingapi.BindingPath) {
			ls := &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
				},
				LabelSelector: ls,
				BindingPath:   bindingPath,
			}

			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: ref,
				},
			}
			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			u := &unstructured.Unstructured{}
			u.SetName("app")
			u.SetNamespace(sb.Namespace)
			u.SetGroupVersionKind(schema.GroupVersionKind{Group: "app", Version: "v1", Kind: "Foo"})
			u.SetLabels(map[string]string{"env": "stage"})
			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			_, err = ctx.Applications()
			Expect(err).To(HaveOccurred())
		},
			Entry("no binding path specified", nil),
			Entry("binding path specified", &bindingapi.BindingPath{ContainersPath: "foo.bar"}),
		)

		It("should return error if application list returns error", func() {
			ls := &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
				},
				LabelSelector: ls,
			}

			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: ref,
				},
			}

			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			client := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{*gvr: "FooList"})
			expectedError := "Error listing foo"
			client.PrependReactor("list", "foos",
				func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, e.New(expectedError)
				})
			authClient := &fakeauth.FakeAuthorizationV1{}
			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			_, err = ctx.Applications()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError))
		})
		It("should return error if application is not found", func() {
			ref := bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}

			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: ref,
				},
			}
			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			client := fake.NewSimpleDynamicClient(runtime.NewScheme())

			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			_, err = ctx.Applications()
			Expect(err).To(HaveOccurred())
		})
		It("should report labels on service bindings when they exist", func() {
			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: bindingapi.Application{
						Ref: bindingapi.Ref{
							Group:   "app",
							Version: "v1",
							Kind:    "Foo",
						},
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"foo": "bar"},
						},
					},
				},
			}

			u := &unstructured.Unstructured{}
			u.SetName("app1")
			u.SetNamespace(sb.Namespace)
			u.SetGroupVersionKind(schema.GroupVersionKind{Group: "app", Version: "v1", Kind: "Foo"})
			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())
			Expect(ctx.HasLabelSelector()).To(BeTrue())
		})
		It("should not report labels on service bindings when they don't exist", func() {
			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: bindingapi.Application{
						Ref: bindingapi.Ref{
							Group:   "app",
							Version: "v1",
							Kind:    "Foo",
							Name:    "app1",
						},
					},
				},
			}

			u := &unstructured.Unstructured{}
			u.SetName("app1")
			u.SetNamespace(sb.Namespace)
			u.SetGroupVersionKind(schema.GroupVersionKind{Group: "app", Version: "v1", Kind: "Foo"})
			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())
			Expect(ctx.HasLabelSelector()).To(BeFalse())
		})
	})

	Describe("Services", func() {
		var (
			defServiceBinding = func(name string, namespace string, refs ...bindingapi.Ref) *bindingapi.ServiceBinding {
				var services []bindingapi.Service
				for idx, ref := range refs {
					id := fmt.Sprintf("id%v", idx)
					services = append(services, bindingapi.Service{
						NamespacedRef: bindingapi.NamespacedRef{
							Ref: ref,
						},
						Id: &id,
					})
				}
				sb := &bindingapi.ServiceBinding{
					Spec: bindingapi.ServiceBindingSpec{
						Services: services,
					},
				}
				return sb
			}
		)

		type testCase struct {
			serviceRefs []bindingapi.Ref
			serviceGVKs []schema.GroupVersionKind
			hasCrd      bool
		}

		DescribeTable("return successfully",
			func(tc *testCase) {
				sb := defServiceBinding("sb1", "ns1", tc.serviceRefs...)
				gvr := &schema.GroupVersionResource{Group: "foo", Version: "v1", Resource: "bars"}
				var objs []runtime.Object
				for i, gvk := range tc.serviceGVKs {
					u := &unstructured.Unstructured{}
					u.SetGroupVersionKind(gvk)
					u.SetName(fmt.Sprintf("s%d", i))
					u.SetNamespace(sb.Namespace)
					objs = append(objs, u)
					if tc.hasCrd {
						crd := &unstructured.Unstructured{}
						crd.SetGroupVersionKind(v1apiextensions.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
						crd.SetName(gvr.GroupResource().String())
						objs = append(objs, crd)
					}
				}
				client := fake.NewSimpleDynamicClient(scheme(objs...), objs...)
				authClient := &fakeauth.FakeAuthorizationV1{}

				typeLookup.EXPECT().ResourceForReferable(gomock.Any()).Return(gvr, nil).Times(len(tc.serviceGVKs))
				typeLookup.EXPECT().ResourceForKind(gomock.Any()).Return(gvr, nil).Times(len(tc.serviceGVKs))

				ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(sb)
				Expect(err).NotTo(HaveOccurred())

				services, err := ctx.Services()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(services)).To(Equal(len(tc.serviceGVKs)))
				for i, s := range services {
					Expect(s.Resource()).To(Equal(objs[i]))
					Expect(*s.Id()).To(Equal(fmt.Sprintf("id%v", i)))
					if tc.hasCrd {
						crd, err := s.CustomResourceDefinition()
						Expect(err).NotTo(HaveOccurred())
						Expect(crd).NotTo(BeNil())
						Expect(crd.Resource().GetName()).To(Equal(gvr.GroupResource().String()))
					}
				}
			},
			Entry("single service", &testCase{
				serviceRefs: []bindingapi.Ref{
					{
						Group:   "foo",
						Version: "v1",
						Kind:    "Bar",
						Name:    "s0",
					},
				},
				serviceGVKs: []schema.GroupVersionKind{
					{
						Group:   "foo",
						Version: "v1",
						Kind:    "Bar",
					},
				},
			}),
			Entry("single service + crd", &testCase{
				serviceRefs: []bindingapi.Ref{
					{
						Group:   "foo",
						Version: "v1",
						Kind:    "Bar",
						Name:    "s0",
					},
				},
				serviceGVKs: []schema.GroupVersionKind{
					{
						Group:   "foo",
						Version: "v1",
						Kind:    "Bar",
					},
				},
				hasCrd: true,
			}),
			Entry("two services", &testCase{
				serviceRefs: []bindingapi.Ref{
					{
						Group:   "foo",
						Version: "v1",
						Kind:    "Bar",
						Name:    "s0",
					},
					{
						Group:   "foo",
						Version: "v1",
						Kind:    "Bar",
						Name:    "s1",
					},
				},
				serviceGVKs: []schema.GroupVersionKind{
					{
						Group:   "foo",
						Version: "v1",
						Kind:    "Bar",
					},
					{
						Group:   "foo",
						Version: "v1",
						Kind:    "Bar",
					},
				},
			}),
		)
		It("Should return error when service not found", func() {
			sb := defServiceBinding("sb1", "ns1", bindingapi.Ref{
				Group:   "foo",
				Version: "v1",
				Kind:    "Bar",
				Name:    "bla",
			})
			client := fake.NewSimpleDynamicClient(runtime.NewScheme())
			authClient := &fakeauth.FakeAuthorizationV1{}

			gvr := &schema.GroupVersionResource{Group: "foo", Version: "v1", Resource: "bars"}
			typeLookup.EXPECT().ResourceForReferable(gomock.Any()).Return(gvr, nil)

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(sb)
			Expect(err).NotTo(HaveOccurred())

			_, err = ctx.Services()
			Expect(err).To(HaveOccurred())
		})

		It("Should return error when one service not found", func() {
			sb := defServiceBinding("sb1", "ns1", bindingapi.Ref{
				Group:   "foo",
				Version: "v1",
				Kind:    "Bar",
				Name:    "bla",
			},
				bindingapi.Ref{
					Group:   "foo",
					Version: "v1",
					Kind:    "Bar",
					Name:    "notfound",
				},
			)
			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(schema.GroupVersionKind{Group: "foo", Version: "v1", Kind: "Bar"})
			u.SetNamespace(sb.GetNamespace())
			u.SetName("bla")
			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			gvr := &schema.GroupVersionResource{Group: "foo", Version: "v1", Resource: "bars"}
			typeLookup.EXPECT().ResourceForReferable(gomock.Any()).Return(gvr, nil).Times(2)
			typeLookup.EXPECT().ResourceForKind(gomock.Any()).Return(gvr, nil).Times(1)

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(sb)
			Expect(err).NotTo(HaveOccurred())

			_, err = ctx.Services()
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Binding Name", func() {
		var testProvider = Provider(nil, nil, nil)
		It("should be equal on .spec.name if specified", func() {
			ctx, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Name: "sb2",
				},
			})

			Expect(ctx.BindingName()).To(Equal("sb2"))
		})

		It("should be equal on .name if .spec.name not specified", func() {
			ctx, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			})

			Expect(ctx.BindingName()).To(Equal("sb1"))
		})
	})

	Describe("Binding Secret Name", func() {
		var testProvider = Provider(nil, nil, nil)
		It("should not be empty string", func() {
			ctx, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			Expect(ctx.BindingSecretName()).NotTo(BeEmpty())
		})

		It("should not depend on binding item order", func() {
			ctx, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			ctx2, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			})
			ctx2.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})
			ctx2.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			Expect(ctx.BindingSecretName()).To(Equal(ctx2.BindingSecretName()))
		})

		It("should be equal to existing secret if additional binding items exist", func() {
			secretName := "foo"
			namespace := "ns1"
			ctx, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
			})
			secret := &unstructured.Unstructured{Object: map[string]interface{}{
				"data": map[string]interface{}{
					"foo1": base64.StdEncoding.EncodeToString([]byte("val1")),
					"foo2": base64.StdEncoding.EncodeToString([]byte("val2")),
				},
			}}
			secret.SetName(secretName)
			secret.SetNamespace(namespace)
			secret.SetAPIVersion("v1")
			secret.SetKind("Secret")

			ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret})

			Expect(ctx.BindingSecretName()).To(Equal(secretName))
		})

		It("should be generated if additional items are added", func() {
			secretName := "foo"
			namespace := "ns1"
			ctx, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
			})
			secret := &unstructured.Unstructured{Object: map[string]interface{}{
				"data": map[string]interface{}{
					"foo1": base64.StdEncoding.EncodeToString([]byte("val1")),
					"foo2": base64.StdEncoding.EncodeToString([]byte("val2")),
				},
			}}
			secret.SetName(secretName)
			secret.SetNamespace(namespace)
			secret.SetAPIVersion("v1")
			secret.SetKind("Secret")

			ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			bindingSecretName := ctx.BindingSecretName()
			Expect(bindingSecretName).NotTo(BeEmpty())
			Expect(bindingSecretName).NotTo(Equal(secretName))
		})

		It("should be generated if item key is modified", func() {
			secretName := "foo"
			namespace := "ns1"
			ctx, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
			})
			secret := &unstructured.Unstructured{Object: map[string]interface{}{
				"data": map[string]interface{}{
					"foo1": base64.StdEncoding.EncodeToString([]byte("val1")),
					"foo2": base64.StdEncoding.EncodeToString([]byte("val2")),
				},
			}}
			secret.SetName(secretName)
			secret.SetNamespace(namespace)
			secret.SetAPIVersion("v1")
			secret.SetKind("Secret")

			service := pipelinemocks.NewMockService(mockCtrl)
			b := &pipeline.SecretBackedBindings{Secret: secret, Service: service}
			ctx.AddBindings(b)
			items, err := b.Items()
			Expect(err).NotTo(HaveOccurred())
			items[0].Name = "bla"

			bindingSecretName := ctx.BindingSecretName()
			Expect(bindingSecretName).NotTo(BeEmpty())
			Expect(bindingSecretName).NotTo(Equal(secretName))
		})

		It("should be generated if two binding secrets are set", func() {
			secretNames := []string{"foo", "bar"}
			namespace := "ns1"
			ctx, _ := testProvider.Get(&bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
			})
			for _, sn := range secretNames {
				secret := &unstructured.Unstructured{Object: map[string]interface{}{
					"data": map[string]interface{}{
						"foo1": base64.StdEncoding.EncodeToString([]byte("val1")),
						"foo2": base64.StdEncoding.EncodeToString([]byte("val2")),
					},
				}}
				secret.SetName(sn)
				secret.SetNamespace(namespace)
				secret.SetAPIVersion("v1")
				secret.SetKind("Secret")

				ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret})
			}

			bindingSecretName := ctx.BindingSecretName()
			Expect(bindingSecretName).NotTo(BeEmpty())
			Expect(bindingSecretName).NotTo(Equal(secretNames[0]))
			Expect(bindingSecretName).NotTo(Equal(secretNames[1]))
		})
	})

	Describe("Close", func() {
		var (
			sb     *bindingapi.ServiceBinding
			ctx    pipeline.Context
			client dynamic.Interface
		)

		BeforeEach(func() {
			sb = &bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
					UID:       "uid1",
				},
			}
			sb.SetGroupVersionKind(bindingapi.GroupVersionKind)
			u, _ := converter.ToUnstructured(&sb)
			s := runtime.NewScheme()
			Expect(bindingapi.AddToScheme(s)).NotTo(HaveOccurred())
			Expect(corev1.AddToScheme(s)).NotTo(HaveOccurred())
			client = fake.NewSimpleDynamicClient(s, u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, _ = Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(sb)

		})

		It("should only persist context conditions on error", func() {

			err1 := "err1"
			err2 := "err2"
			ctx.SetCondition(apis.Conditions().NotInjectionReady().ServiceNotFound().Msg(err1).Build())
			ctx.SetCondition(apis.Conditions().NotCollectionReady().ServiceNotFound().Msg(err2).Build())

			ctx.Error(e.New(err1))

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(bindingapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := bindingapi.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).To(BeEmpty())
			Expect(updatedSB.Status.Conditions).To(HaveLen(3))

			cnd := meta.FindStatusCondition(updatedSB.Status.Conditions, apis.InjectionReady)
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))
			Expect(cnd.Reason).To(Equal(apis.ServiceNotFoundReason))
			Expect(cnd.Message).To(Equal(err1))

			cnd = meta.FindStatusCondition(updatedSB.Status.Conditions, apis.CollectionReady)
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))
			Expect(cnd.Reason).To(Equal(apis.ServiceNotFoundReason))
			Expect(cnd.Message).To(Equal(err2))

			cnd = meta.FindStatusCondition(updatedSB.Status.Conditions, apis.BindingReady)
			Expect(cnd.Type).To(Equal(apis.BindingReady))
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))

		})

		It("should create only secret if no application is defined", func() {
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(bindingapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := bindingapi.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Conditions).To(HaveLen(1))
			Expect(updatedSB.Status.Conditions[0].Type).To(Equal(apis.BindingReady))
			Expect(updatedSB.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
		})

		It("should update application if changed", func() {
			sb.Spec.Application = bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			gvr := schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&(sb.Spec.Application)).Return(&gvr, nil)

			u := &unstructured.Unstructured{}
			u.SetNamespace(sb.Namespace)
			u.SetName("app1")
			u.SetGroupVersionKind(gvr.GroupVersion().WithKind("Foo"))

			_, err := client.Resource(gvr).Namespace(sb.Namespace).Create(context.Background(), u, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			apps, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			specData := map[string]interface{}{
				"foo": "bar",
			}
			apps[0].Resource().Object["Spec"] = specData

			err = ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err = client.Resource(bindingapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := bindingapi.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Conditions).To(HaveLen(1))
			Expect(updatedSB.Status.Conditions[0].Type).To(Equal(apis.BindingReady))
			Expect(updatedSB.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))

			u, err = client.Resource(gvr).Namespace(sb.Namespace).Get(context.Background(), "app1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(u.Object["Spec"]).To(Equal(specData))

		})
		It("should not update service binding if its uid is unset", func() {
			sb.UID = ""
			sb.Name = "sb2"
			sb.Spec.Application = bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			gvr := schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&(sb.Spec.Application)).Return(&gvr, nil)

			u := &unstructured.Unstructured{}
			u.SetNamespace(sb.Namespace)
			u.SetName("app1")
			u.SetGroupVersionKind(gvr.GroupVersion().WithKind("Foo"))

			_, err := client.Resource(gvr).Namespace(sb.Namespace).Create(context.Background(), u, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			apps, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			specData := map[string]interface{}{
				"foo": "bar",
			}
			apps[0].Resource().Object["Spec"] = specData

			err = ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Resource(bindingapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).To(HaveOccurred())

			u, err = client.Resource(gvr).Namespace(sb.Namespace).Get(context.Background(), "app1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(u.Object["Spec"]).To(Equal(specData))

		})
	})

	Describe("Persist Secret", func() {
		var (
			sb     *bindingapi.ServiceBinding
			ctx    pipeline.Context
			client dynamic.Interface
		)

		BeforeEach(func() {
			sb = &bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
					UID:       "uid1",
				},
			}
			sb.SetGroupVersionKind(bindingapi.GroupVersionKind)
			u, _ := converter.ToUnstructured(&sb)
			s := runtime.NewScheme()
			Expect(bindingapi.AddToScheme(s)).NotTo(HaveOccurred())
			Expect(corev1.AddToScheme(s)).NotTo(HaveOccurred())
			client = fake.NewSimpleDynamicClient(s, u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, _ = Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(sb)

		})

		It("should reuse existing secret if no other bindings are added", func() {
			secret := &unstructured.Unstructured{Object: map[string]interface{}{
				"data": map[string]interface{}{
					"foo1": base64.StdEncoding.EncodeToString([]byte("val1")),
					"foo2": base64.StdEncoding.EncodeToString([]byte("val2")),
				},
			}}
			secret.SetName("foo")
			secret.SetNamespace("ns1")
			secret.SetAPIVersion("v1")
			secret.SetKind("Secret")
			service := pipelinemocks.NewMockService(mockCtrl)

			ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret, Service: service})

			err := ctx.PersistSecret()
			Expect(err).NotTo(HaveOccurred())

			Expect(sb.Status.Secret).Should(Equal(secret.GetName()))

			secretList, err := client.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}).Namespace(sb.Namespace).List(context.Background(), metav1.ListOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(secretList.Items).Should(BeEmpty())
		})

		It("Should create intermediate secret if additional bindings are added", func() {
			secret := &unstructured.Unstructured{Object: map[string]interface{}{
				"data": map[string]interface{}{
					"foo1": base64.StdEncoding.EncodeToString([]byte("val1")),
					"foo2": base64.StdEncoding.EncodeToString([]byte("val2")),
				},
			}}
			secret.SetName("foo")
			secret.SetNamespace("ns1")
			secret.SetAPIVersion("v1")
			secret.SetKind("Secret")
			service := pipelinemocks.NewMockService(mockCtrl)

			ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret, Service: service})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo3", Value: "val3"})

			err := ctx.PersistSecret()
			Expect(err).NotTo(HaveOccurred())

			Expect(sb.Status.Secret).ShouldNot(Equal(secret.GetName()))
			Expect(sb.Status.Secret).ShouldNot(BeEmpty())

			u, err := ctx.ReadSecret(sb.Namespace, sb.Status.Secret)
			Expect(err).NotTo(HaveOccurred())

			intermediateSecret := &corev1.Secret{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, intermediateSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(intermediateSecret.StringData).To(HaveLen(3))
			Expect(intermediateSecret.StringData).Should(HaveKeyWithValue("foo1", "val1"))
			Expect(intermediateSecret.StringData).Should(HaveKeyWithValue("foo2", "val2"))
			Expect(intermediateSecret.StringData).Should(HaveKeyWithValue("foo3", "val3"))
		})
	})

	Describe("Mapping template", func() {
		It("should not use a wildcard when a matching version is specified", func() {
			ref := bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}

			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: ref,
				},
			}
			gvr := schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}

			mappingSpec := v1beta1.ClusterWorkloadResourceMapping{
				Spec: v1beta1.ClusterWorkloadResourceMappingSpec{
					Versions: []v1beta1.ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
							Volumes: ".spec.volumeSpec",
						},
						{
							Version: "*",
							Volumes: ".spec.volumes",
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foos.app",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.WorkloadResourceMappingGroupVersionKind.Kind,
					APIVersion: "servicebinding.io/v1beta1",
				},
			}

			bytes, err := json.Marshal(mappingSpec)
			Expect(err).NotTo(HaveOccurred())

			var data map[string]interface{}
			err = json.Unmarshal(bytes, &data)
			Expect(err).NotTo(HaveOccurred())

			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(v1beta1.WorkloadResourceMappingGroupVersionKind)
			u.SetName("foos.app")
			u.SetUnstructuredContent(data)
			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			mapping, err := ctx.WorkloadResourceTemplate(&gvr, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(mapping.Volume).To(Equal([]string{"spec", "volumeSpec"}))
		})

		It("should use defaults when none are specified", func() {
			ref := bindingapi.Application{
				Ref: bindingapi.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}

			sb := bindingapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: bindingapi.ServiceBindingSpec{
					Application: ref,
				},
			}
			gvr := schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}

			mappingSpec := v1beta1.ClusterWorkloadResourceMapping{
				Spec: v1beta1.ClusterWorkloadResourceMappingSpec{
					Versions: []v1beta1.ClusterWorkloadResourceMappingTemplate{
						{
							Version: "*",
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "foos.app",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       v1beta1.WorkloadResourceMappingGroupVersionKind.Kind,
					APIVersion: "servicebinding.io/v1beta1",
				},
			}

			bytes, err := json.Marshal(mappingSpec)
			Expect(err).NotTo(HaveOccurred())

			var data map[string]interface{}
			err = json.Unmarshal(bytes, &data)
			Expect(err).NotTo(HaveOccurred())

			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(v1beta1.WorkloadResourceMappingGroupVersionKind)
			u.SetName("foos.app")
			u.SetUnstructuredContent(data)
			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u)
			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			mapping, err := ctx.WorkloadResourceTemplate(&gvr, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(mapping.Volume).To(Equal([]string{"spec", "template", "spec", "volumes"}))
			Expect(mapping.Containers).To(HaveLen(2))
			Expect(mapping.Containers[0].Env).To(Equal([]string{"env"}))
			Expect(mapping.Containers[0].EnvFrom).To(Equal([]string{"envFrom"}))
			Expect(mapping.Containers[0].Name).To(Equal([]string{"name"}))
			Expect(mapping.Containers[0].VolumeMounts).To(Equal([]string{"volumeMounts"}))
		})
	})
})
