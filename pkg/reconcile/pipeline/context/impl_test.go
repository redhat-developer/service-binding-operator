package context

import (
	"context"
	"encoding/base64"
	e "errors"
	"fmt"
	v1alpha12 "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context/mocks"
	mocks2 "github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/mocks"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/testing"
)

var _ = Describe("Context", func() {

	var (
		mockCtrl   *gomock.Controller
		typeLookup *mocks.MockK8STypeLookup
		client     dynamic.Interface
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		typeLookup = mocks.NewMockK8STypeLookup(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Applications", func() {

		DescribeTable("should return slice of size 1 if application is specified", func(bindingPath *v1alpha12.BindingPath, expectedContainerPath string) {
			ref := v1alpha12.Application{
				Ref: v1alpha12.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
				BindingPath: bindingPath,
			}

			sb := v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: v1alpha12.ServiceBindingSpec{
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

			ctx := &impl{client: client, serviceBinding: &sb, typeLookup: typeLookup}

			applications, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(applications).To(HaveLen(1))
			Expect(applications[0].Resource()).To(Equal(u))
			Expect(applications[0].ContainersPath()).To(Equal(expectedContainerPath))
		},
			Entry("no binding path specified", nil, defaultContainerPath),
			Entry("binding path specified", &v1alpha12.BindingPath{ContainersPath: "foo.bar"}, "foo.bar"),
		)
		DescribeTable("should return slice of size 2 if 2 applications are specified through label seclector", func(bindingPath *v1alpha12.BindingPath, expectedContainerPath string) {
			ls := &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := v1alpha12.Application{
				Ref: v1alpha12.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
				},
				LabelSelector: ls,
				BindingPath:   bindingPath,
			}

			sb := v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: v1alpha12.ServiceBindingSpec{
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

			ctx := &impl{client: client, serviceBinding: &sb, typeLookup: typeLookup}

			applications, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(applications).To(HaveLen(2))

			Expect(applications[0].Resource().GetName()).NotTo(Equal(applications[1].Resource().GetName()))
			Expect(applications[0].Resource()).Should(BeElementOf(u1, u2))
			Expect(applications[1].Resource()).Should(BeElementOf(u1, u2))
			Expect(applications[0].ContainersPath()).To(Equal(expectedContainerPath))
			Expect(applications[1].ContainersPath()).To(Equal(expectedContainerPath))
		},
			Entry("no binding path specified", nil, defaultContainerPath),
			Entry("binding path specified", &v1alpha12.BindingPath{ContainersPath: "foo.bar"}, "foo.bar"),
		)
		DescribeTable("should return slice of size 0 if no application is matching through label seclector", func(bindingPath *v1alpha12.BindingPath, expectedContainerPath string) {
			ls := &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := v1alpha12.Application{
				Ref: v1alpha12.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
				},
				LabelSelector: ls,
				BindingPath:   bindingPath,
			}

			sb := v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: v1alpha12.ServiceBindingSpec{
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

			ctx := &impl{client: client, serviceBinding: &sb, typeLookup: typeLookup}

			_, err := ctx.Applications()
			Expect(err).To(HaveOccurred())
		},
			Entry("no binding path specified", nil, defaultContainerPath),
			Entry("binding path specified", &v1alpha12.BindingPath{ContainersPath: "foo.bar"}, "foo.bar"),
		)

		It("should return error if application list returns error", func() {
			ls := &metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := v1alpha12.Application{
				Ref: v1alpha12.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
				},
				LabelSelector: ls,
			}

			sb := v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: v1alpha12.ServiceBindingSpec{
					Application: ref,
				},
			}

			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			client := fake.NewSimpleDynamicClient(runtime.NewScheme())
			expectedError := "Error listing foo"
			client.PrependReactor("list", "foos",
				func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, e.New(expectedError)
				})

			ctx := &impl{client: client, serviceBinding: &sb, typeLookup: typeLookup}

			_, err := ctx.Applications()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError))
		})
		It("should return error if application is not found", func() {
			ref := v1alpha12.Application{
				Ref: v1alpha12.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}

			sb := v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: v1alpha12.ServiceBindingSpec{
					Application: ref,
				},
			}
			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			client := fake.NewSimpleDynamicClient(runtime.NewScheme())

			ctx := &impl{client: client, serviceBinding: &sb, typeLookup: typeLookup}

			_, err := ctx.Applications()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Services", func() {
		var (
			defServiceBinding = func(name string, namespace string, refs ...v1alpha12.Ref) *v1alpha12.ServiceBinding {
				var services []v1alpha12.Service
				for idx, ref := range refs {
					id := fmt.Sprintf("id%v", idx)
					services = append(services, v1alpha12.Service{
						NamespacedRef: v1alpha12.NamespacedRef{
							Ref: ref,
						},
						Id: &id,
					})
				}
				sb := &v1alpha12.ServiceBinding{
					Spec: v1alpha12.ServiceBindingSpec{
						Services: services,
					},
				}
				return sb
			}
		)

		type testCase struct {
			serviceRefs []v1alpha12.Ref
			serviceGVKs []schema.GroupVersionKind
		}

		DescribeTable("return successfully",
			func(tc *testCase) {
				sb := defServiceBinding("sb1", "ns1", tc.serviceRefs...)
				var objs []runtime.Object
				for i, gvk := range tc.serviceGVKs {
					u := &unstructured.Unstructured{}
					u.SetGroupVersionKind(gvk)
					u.SetName(fmt.Sprintf("s%d", i))
					u.SetNamespace(sb.Namespace)
					objs = append(objs, u)
				}
				client = fake.NewSimpleDynamicClient(runtime.NewScheme(), objs...)
				gvr := &schema.GroupVersionResource{Group: "foo", Version: "v1", Resource: "bars"}
				typeLookup.EXPECT().ResourceForReferable(gomock.Any()).Return(gvr, nil).Times(len(tc.serviceGVKs))

				ctx := &impl{
					serviceBinding: sb,
					client:         client,
					typeLookup:     typeLookup,
				}

				services, err := ctx.Services()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(services)).To(Equal(len(tc.serviceGVKs)))
				for i, s := range services {
					Expect(s.Resource()).To(Equal(objs[i]))
					serviceImpl, ok := s.(*service)
					if !ok {
						Fail("not service impl")
					}
					Expect(serviceImpl.client).To(Equal(client))
					Expect(serviceImpl.groupVersionResource).To(Equal(gvr))
					Expect(serviceImpl.serviceRef.Name).To(Equal(tc.serviceRefs[i].Name))
					Expect(*serviceImpl.serviceRef.Id).To(Equal(fmt.Sprintf("id%v", i)))
				}
			},
			Entry("single service", &testCase{
				serviceRefs: []v1alpha12.Ref{
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
			Entry("two services", &testCase{
				serviceRefs: []v1alpha12.Ref{
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
			sb := defServiceBinding("sb1", "ns1", v1alpha12.Ref{
				Group:   "foo",
				Version: "v1",
				Kind:    "Bar",
				Name:    "bla",
			})
			client = fake.NewSimpleDynamicClient(runtime.NewScheme())
			gvr := &schema.GroupVersionResource{Group: "foo", Version: "v1", Resource: "bars"}
			typeLookup.EXPECT().ResourceForReferable(gomock.Any()).Return(gvr, nil)

			ctx := &impl{
				serviceBinding: sb,
				client:         client,
				typeLookup:     typeLookup,
			}

			_, err := ctx.Services()
			Expect(err).To(HaveOccurred())
		})

		It("Should return error when one service not found", func() {
			sb := defServiceBinding("sb1", "ns1", v1alpha12.Ref{
				Group:   "foo",
				Version: "v1",
				Kind:    "Bar",
				Name:    "bla",
			},
				v1alpha12.Ref{
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
			client = fake.NewSimpleDynamicClient(runtime.NewScheme(), u)
			gvr := &schema.GroupVersionResource{Group: "foo", Version: "v1", Resource: "bars"}
			typeLookup.EXPECT().ResourceForReferable(gomock.Any()).Return(gvr, nil).Times(2)

			ctx := &impl{
				serviceBinding: sb,
				client:         client,
				typeLookup:     typeLookup,
			}

			_, err := ctx.Services()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Binding Secret Name", func() {
		It("should not be empty string", func() {
			ctx := &impl{serviceBinding: &v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			}}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			Expect(ctx.BindingSecretName()).NotTo(BeEmpty())
		})

		It("should not depend on binding item order", func() {
			ctx := &impl{serviceBinding: &v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			}}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			ctx2 := &impl{serviceBinding: &v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			}}
			ctx2.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})
			ctx2.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			Expect(ctx.BindingSecretName()).To(Equal(ctx2.BindingSecretName()))
		})

		It("should be equal to existing secret if additional binding items exist", func() {
			secretName := "foo"
			namespace := "ns1"
			ctx := &impl{serviceBinding: &v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
			}}
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
			ctx := &impl{serviceBinding: &v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
			}}
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
			ctx := &impl{serviceBinding: &v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
			}}
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

			service := mocks2.NewMockService(mockCtrl)
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
			ctx := &impl{serviceBinding: &v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
			}}
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
			sb  *v1alpha12.ServiceBinding
			ctx pipeline.Context
		)

		BeforeEach(func() {
			sb = &v1alpha12.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
					UID:       "uid1",
				},
			}
			sb.SetGroupVersionKind(v1alpha12.GroupVersionKind)
			u, _ := converter.ToUnstructured(&sb)
			client = fake.NewSimpleDynamicClient(runtime.NewScheme(), u)

			ctx, _ = Provider(client, typeLookup).Get(sb)

		})

		It("should only persist context conditions on error", func() {

			err1 := "err1"
			err2 := "err2"
			ctx.SetCondition(v1alpha12.Conditions().NotInjectionReady().ServiceNotFound().Msg(err1).Build())
			ctx.SetCondition(v1alpha12.Conditions().NotCollectionReady().ServiceNotFound().Msg(err2).Build())

			ctx.Error(e.New(err1))

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(v1alpha12.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := v1alpha12.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).To(BeEmpty())
			Expect(updatedSB.Status.Conditions).To(HaveLen(3))

			cnd := meta.FindStatusCondition(updatedSB.Status.Conditions, v1alpha12.InjectionReady)
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))
			Expect(cnd.Reason).To(Equal(v1alpha12.ServiceNotFoundReason))
			Expect(cnd.Message).To(Equal(err1))

			cnd = meta.FindStatusCondition(updatedSB.Status.Conditions, v1alpha12.CollectionReady)
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))
			Expect(cnd.Reason).To(Equal(v1alpha12.ServiceNotFoundReason))
			Expect(cnd.Message).To(Equal(err2))

			cnd = meta.FindStatusCondition(updatedSB.Status.Conditions, v1alpha12.BindingReady)
			Expect(cnd.Type).To(Equal(v1alpha12.BindingReady))
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))

		})

		It("should create only secret if no application is defined", func() {
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(v1alpha12.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := v1alpha12.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).NotTo(BeEmpty())
			Expect(updatedSB.Status.Conditions).To(HaveLen(1))
			Expect(updatedSB.Status.Conditions[0].Type).To(Equal(v1alpha12.BindingReady))
			Expect(updatedSB.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))

			u, err = ctx.ReadSecret(sb.Namespace, updatedSB.Status.Secret)
			Expect(err).NotTo(HaveOccurred())

			secret := &corev1.Secret{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, secret)
			Expect(err).NotTo(HaveOccurred())
			items := ctx.BindingItems()
			Expect(secret.StringData).To(Equal(items.AsMap()))
		})

		It("should update application if changed", func() {
			sb.Spec.Application = v1alpha12.Application{
				Ref: v1alpha12.Ref{
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

			u, err = client.Resource(v1alpha12.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := v1alpha12.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).NotTo(BeEmpty())
			Expect(updatedSB.Status.Conditions).To(HaveLen(1))
			Expect(updatedSB.Status.Conditions[0].Type).To(Equal(v1alpha12.BindingReady))
			Expect(updatedSB.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))

			u, err = ctx.ReadSecret(sb.Namespace, updatedSB.Status.Secret)
			Expect(err).NotTo(HaveOccurred())

			secret := &corev1.Secret{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, secret)
			Expect(err).NotTo(HaveOccurred())
			bindingItems := ctx.BindingItems()
			Expect(secret.StringData).To(Equal(bindingItems.AsMap()))
			Expect(secret.OwnerReferences[0].UID).To(Equal(sb.UID))

			u, err = client.Resource(gvr).Namespace(sb.Namespace).Get(context.Background(), "app1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(u.Object["Spec"]).To(Equal(specData))

		})
		It("should not update service binding if its uid is unset", func() {
			sb.UID = ""
			sb.Name = "sb2"
			sb.Spec.Application = v1alpha12.Application{
				Ref: v1alpha12.Ref{
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

			_, err = client.Resource(v1alpha12.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).To(HaveOccurred())

			u, err = ctx.ReadSecret(sb.Namespace, sb.Status.Secret)
			Expect(err).NotTo(HaveOccurred())

			secret := &corev1.Secret{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, secret)
			Expect(err).NotTo(HaveOccurred())
			bindingItems := ctx.BindingItems()
			Expect(secret.StringData).To(Equal(bindingItems.AsMap()))
			Expect(secret.OwnerReferences).To(HaveLen(0))

			u, err = client.Resource(gvr).Namespace(sb.Namespace).Get(context.Background(), "app1", metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			Expect(u.Object["Spec"]).To(Equal(specData))

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
			service := mocks2.NewMockService(mockCtrl)

			ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret, Service: service})

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(v1alpha12.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := v1alpha12.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).Should(Equal(secret.GetName()))

			secretList, err := client.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}).List(context.Background(), metav1.ListOptions{})
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
			service := mocks2.NewMockService(mockCtrl)

			ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret, Service: service})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo3", Value: "val3"})

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(v1alpha12.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := v1alpha12.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).ShouldNot(Equal(secret.GetName()))
			Expect(updatedSB.Status.Secret).ShouldNot(BeEmpty())

			u, err = ctx.ReadSecret(sb.Namespace, sb.Status.Secret)
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
})
