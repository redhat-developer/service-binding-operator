package context

import (
	"context"
	e "errors"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context/mocks"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
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
		It("Should return empty slice if not application declared in service binding", func() {
			sb := v1alpha1.ServiceBinding{
				Spec: v1alpha1.ServiceBindingSpec{},
			}
			ctx := &impl{serviceBinding: &sb}
			applications, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(applications).To(BeEmpty())
		})

		DescribeTable("should return slice of size 1 if application is specified", func(bindingPath *v1alpha1.BindingPath, expectedContainerPath string) {
			ref := v1alpha1.Application{
				Ref: v1alpha1.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
				BindingPath: bindingPath,
			}

			sb := v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: v1alpha1.ServiceBindingSpec{
					Application: &ref,
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
			Entry("binding path specified", &v1alpha1.BindingPath{ContainersPath: "foo.bar"}, "foo.bar"),
		)

		It("should return error if application is not found", func() {
			ref := v1alpha1.Application{
				Ref: v1alpha1.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}

			sb := v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: v1alpha1.ServiceBindingSpec{
					Application: &ref,
				},
			}
			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			client := fake.NewSimpleDynamicClient(runtime.NewScheme())

			ctx := &impl{client: client, serviceBinding: &sb, typeLookup: typeLookup}

			_, err := ctx.Applications()
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("Services", func() {
		var (
			defServiceBinding = func(name string, namespace string, refs ...v1alpha1.Ref) *v1alpha1.ServiceBinding {
				var services []v1alpha1.Service
				for _, ref := range refs {
					services = append(services, v1alpha1.Service{
						NamespacedRef: v1alpha1.NamespacedRef{
							Ref: ref,
						},
					})
				}
				sb := &v1alpha1.ServiceBinding{
					Spec: v1alpha1.ServiceBindingSpec{
						Services: services,
					},
				}
				return sb
			}
		)

		type testCase struct {
			serviceRefs []v1alpha1.Ref
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
				}
			},
			Entry("single service", &testCase{
				serviceRefs: []v1alpha1.Ref{
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
				serviceRefs: []v1alpha1.Ref{
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
			sb := defServiceBinding("sb1", "ns1", v1alpha1.Ref{
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
			sb := defServiceBinding("sb1", "ns1", v1alpha1.Ref{
				Group:   "foo",
				Version: "v1",
				Kind:    "Bar",
				Name:    "bla",
			},
				v1alpha1.Ref{
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
			ctx := &impl{serviceBinding: &v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			}}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			Expect(ctx.BindingSecretName()).NotTo(BeEmpty())
		})

		It("should not depend on binding item order", func() {
			ctx := &impl{serviceBinding: &v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			}}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			ctx2 := &impl{serviceBinding: &v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			}}
			ctx2.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})
			ctx2.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			Expect(ctx.BindingSecretName()).To(Equal(ctx2.BindingSecretName()))
		})
	})

	Describe("Close", func() {
		var (
			sb  *v1alpha1.ServiceBinding
			ctx pipeline.Context
		)

		BeforeEach(func() {
			sb = &v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
					UID: "uid1",
				},
			}
			sb.SetGroupVersionKind(v1alpha1.GroupVersionKind)
			u, _ := converter.ToUnstructured(&sb)
			client = fake.NewSimpleDynamicClient(runtime.NewScheme(), u)

			ctx, _ = Provider(client, typeLookup).Get(sb)

		})

		It("should only persist context conditions on error", func() {

			err1 := "err1"
			err2 := "err2"
			ctx.SetCondition(v1alpha1.Conditions().NotInjectionReady().ServiceNotFound().Msg(err1).Build())
			ctx.SetCondition(v1alpha1.Conditions().NotCollectionReady().ServiceNotFound().Msg(err2).Build())

			ctx.Error(e.New(err1))

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(v1alpha1.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := v1alpha1.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).To(BeEmpty())
			Expect(updatedSB.Status.Conditions).To(HaveLen(3))

			cnd := meta.FindStatusCondition(updatedSB.Status.Conditions, v1alpha1.InjectionReady)
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))
			Expect(cnd.Reason).To(Equal(v1alpha1.ServiceNotFoundReason))
			Expect(cnd.Message).To(Equal(err1))

			cnd = meta.FindStatusCondition(updatedSB.Status.Conditions, v1alpha1.CollectionReady)
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))
			Expect(cnd.Reason).To(Equal(v1alpha1.ServiceNotFoundReason))
			Expect(cnd.Message).To(Equal(err2))

			cnd = meta.FindStatusCondition(updatedSB.Status.Conditions, v1alpha1.BindingReady)
			Expect(cnd.Type).To(Equal(v1alpha1.BindingReady))
			Expect(cnd.Status).To(Equal(metav1.ConditionFalse))

		})

		It("should create only secret if no application is defined", func() {
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(v1alpha1.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := v1alpha1.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).NotTo(BeEmpty())
			Expect(updatedSB.Status.Conditions).To(HaveLen(1))
			Expect(updatedSB.Status.Conditions[0].Type).To(Equal(v1alpha1.BindingReady))
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
			sb.Spec.Application = &v1alpha1.Application{
				Ref: v1alpha1.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			gvr := schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(sb.Spec.Application).Return(&gvr, nil)

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

			u, err = client.Resource(v1alpha1.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := v1alpha1.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Secret).NotTo(BeEmpty())
			Expect(updatedSB.Status.Conditions).To(HaveLen(1))
			Expect(updatedSB.Status.Conditions[0].Type).To(Equal(v1alpha1.BindingReady))
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
			sb.Spec.Application = &v1alpha1.Application{
				Ref: v1alpha1.Ref{
					Group:   "app",
					Version: "v1",
					Kind:    "Foo",
					Name:    "app1",
				},
			}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			gvr := schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(sb.Spec.Application).Return(&gvr, nil)

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

			_, err = client.Resource(v1alpha1.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
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
	})
})
