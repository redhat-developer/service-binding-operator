package context

import (
	"context"
	"encoding/base64"
	e "errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis"
	specapi "github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha2"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context/mocks"
	pipelinemocks "github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/mocks"
	corev1 "k8s.io/api/core/v1"
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

var _ = Describe("Spec API Context", func() {

	var (
		mockCtrl   *gomock.Controller
		typeLookup *mocks.MockK8STypeLookup
		Provider   = SpecProvider
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		typeLookup = mocks.NewMockK8STypeLookup(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Applications", func() {

		It("should return slice of size 1", func() {
			ref := specapi.ServiceBindingApplicationReference{
				APIVersion: "app/v1",
				Kind:       "Foo",
				Name:       "app1",
			}

			sb := specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: specapi.ServiceBindingSpec{
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

			ctx, err := SpecProvider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			applications, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(applications).To(HaveLen(1))
			Expect(applications[0].Resource()).To(Equal(u))
			Expect(applications[0].ContainersPath()).To(Equal(defaultContainerPath))
		})
		It("should return slice of size 2 if 2 applications are specified through label selector", func() {
			ls := metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := specapi.ServiceBindingApplicationReference{
				APIVersion: "app/v1",
				Kind:       "Foo",
				Selector:   ls,
			}

			sb := specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: specapi.ServiceBindingSpec{
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

			ctx, err := SpecProvider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			applications, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(applications).To(HaveLen(2))

			Expect(applications[0].Resource().GetName()).NotTo(Equal(applications[1].Resource().GetName()))
			Expect(applications[0].Resource()).Should(BeElementOf(u1, u2))
			Expect(applications[1].Resource()).Should(BeElementOf(u1, u2))
			Expect(applications[0].ContainersPath()).To(Equal(defaultContainerPath))
			Expect(applications[1].ContainersPath()).To(Equal(defaultContainerPath))
		})
		It("should return error if no application is matching through label selector", func() {
			ls := metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := specapi.ServiceBindingApplicationReference{
				APIVersion: "app/v1",
				Kind:       "Foo",
				Selector:   ls,
			}

			sb := specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: specapi.ServiceBindingSpec{
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

			ctx, err := SpecProvider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			_, err = ctx.Applications()
			Expect(err).To(HaveOccurred())
		})
		It("should return error if application list returns error", func() {
			ls := metav1.LabelSelector{
				MatchLabels: map[string]string{"env": "prod"},
			}

			ref := specapi.ServiceBindingApplicationReference{
				APIVersion: "app/v1",
				Kind:       "Foo",
				Selector:   ls,
			}

			sb := specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: specapi.ServiceBindingSpec{
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
			authClient := &fakeauth.FakeAuthorizationV1{}
			ctx, err := SpecProvider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			_, err = ctx.Applications()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedError))
		})
		It("should return error if application is not found", func() {
			ref := specapi.ServiceBindingApplicationReference{
				APIVersion: "app/v1",
				Kind:       "Foo",
				Name:       "app1",
			}

			sb := specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: specapi.ServiceBindingSpec{
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
		It("should return application with bindable containers", func() {
			ref := specapi.ServiceBindingApplicationReference{
				APIVersion: "app/v1",
				Kind:       "Foo",
				Name:       "app1",
				Containers: []string{"c2", "c3", "c1"},
			}

			sb := specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
				},
				Spec: specapi.ServiceBindingSpec{
					Application: ref,
				},
			}
			gvr := &schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(gvr, nil)

			c1 := corev1.Container{
				Image: "foo",
			}
			c2 := corev1.Container{
				Name:  "c2",
				Image: "foo2",
			}
			c3 := corev1.Container{
				Name:  "c3",
				Image: "foo3",
			}
			d1 := deployment("app1", []corev1.Container{c1, c2, c3})

			u, _ := converter.ToUnstructured(&d1)
			cu2, _ := converter.ToUnstructured(&c2)
			cu3, _ := converter.ToUnstructured(&c3)
			u.SetName("app1")
			u.SetNamespace(sb.Namespace)
			u.SetGroupVersionKind(schema.GroupVersionKind{Group: "app", Version: "v1", Kind: "Foo"})
			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), u)

			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := SpecProvider(client, authClient.SubjectAccessReviews(), typeLookup).Get(&sb)
			Expect(err).NotTo(HaveOccurred())

			applications, err := ctx.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(applications).To(HaveLen(1))
			containers, err := applications[0].BindableContainers()
			Expect(err).NotTo(HaveOccurred())
			Expect(containers).To(ConsistOf(cu2.Object, cu3.Object))
		})
	})

	Describe("Services", func() {
		var (
			defServiceBinding = func(name string, namespace string, ref specapi.ServiceBindingServiceReference) *specapi.ServiceBinding {
				sb := &specapi.ServiceBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
					Spec: specapi.ServiceBindingSpec{
						Service: ref,
					},
				}
				return sb
			}
		)

		It("return successfully", func() {
			sb := defServiceBinding("sb1", "ns1", specapi.ServiceBindingServiceReference{
				APIVersion: "foo/v1",
				Kind:       "Bar",
				Name:       "s0",
			})
			var objs []runtime.Object
			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "foo",
				Version: "v1",
				Kind:    "Bar",
			})
			u.SetName("s0")
			u.SetNamespace(sb.Namespace)
			objs = append(objs, u)

			client := fake.NewSimpleDynamicClient(runtime.NewScheme(), objs...)
			gvr := &schema.GroupVersionResource{Group: "foo", Version: "v1", Resource: "bars"}
			typeLookup.EXPECT().ResourceForReferable(gomock.Any()).Return(gvr, nil)

			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(sb)
			Expect(err).NotTo(HaveOccurred())

			services, err := ctx.Services()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(services)).To(Equal(1))
			s := services[0]

			Expect(s.Resource()).To(Equal(u))
			serviceImpl, ok := s.(*service)
			if !ok {
				Fail("not service impl")
			}
			Expect(serviceImpl.client).To(Equal(client))
			Expect(serviceImpl.groupVersionResource).To(Equal(gvr))
			Expect(serviceImpl.namespace).To(Equal(u.GetNamespace()))
			Expect(serviceImpl.id).To(BeNil())
		})
		It("Should return error when service not found", func() {
			sb := defServiceBinding("sb1", "ns1", specapi.ServiceBindingServiceReference{
				APIVersion: "foo/v1",
				Kind:       "Bar",
				Name:       "bla",
			})
			client := fake.NewSimpleDynamicClient(runtime.NewScheme())
			gvr := &schema.GroupVersionResource{Group: "foo", Version: "v1", Resource: "bars"}
			typeLookup.EXPECT().ResourceForReferable(gomock.Any()).Return(gvr, nil)

			authClient := &fakeauth.FakeAuthorizationV1{}

			ctx, err := Provider(client, authClient.SubjectAccessReviews(), typeLookup).Get(sb)
			Expect(err).NotTo(HaveOccurred())

			_, err = ctx.Services()
			Expect(err).To(HaveOccurred())
		})

	})

	Describe("Binding Name", func() {
		var testProvider = Provider(nil, nil, nil)
		It("should be equal on .spec.name if specified", func() {
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
				Spec: specapi.ServiceBindingSpec{
					Name: "sb2",
				},
			})

			Expect(ctx.BindingName()).To(Equal("sb2"))
		})

		It("should be equal on .name if .spec.name not specified", func() {
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
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
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			Expect(ctx.BindingSecretName()).NotTo(BeEmpty())
		})

		It("should not depend on binding item order", func() {
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			ctx2, _ := testProvider.Get(&specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "sb1",
				},
			})
			ctx2.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})
			ctx2.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})

			Expect(ctx.BindingSecretName()).To(Equal(ctx2.BindingSecretName()))
		})

		It("should be equal to existing secret if no additional binding items exist", func() {
			secretName := "foo"
			namespace := "ns1"
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
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
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
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

		DescribeTable("should be generated if type is set on binding", func(spec specapi.ServiceBindingSpec, key string, value string) {
			secretName := "foo"
			namespace := "ns1"
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: namespace,
				},
				Spec: spec,
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

			bindingSecretName := ctx.BindingSecretName()
			Expect(bindingSecretName).NotTo(BeEmpty())
			Expect(bindingSecretName).NotTo(Equal(secretName))

			bindingItems := ctx.BindingItems()
			items := bindingItems.AsMap()
			Expect(items).To(HaveLen(3))
			Expect(items).Should(HaveKeyWithValue(key, value))
			Expect(items).Should(HaveKeyWithValue("foo1", "val1"))
			Expect(items).Should(HaveKeyWithValue("foo2", "val2"))
		},
			Entry("if type is set on binding",
				specapi.ServiceBindingSpec{
					Type: "mysql",
				}, "type", "mysql"),
			Entry("if provider is set on binding",
				specapi.ServiceBindingSpec{
					Provider: "mysql",
				}, "provider", "mysql"))

		It("should be generated if item key is modified", func() {
			secretName := "foo"
			namespace := "ns1"
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
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
			ctx, _ := testProvider.Get(&specapi.ServiceBinding{
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
			sb     *specapi.ServiceBinding
			ctx    pipeline.Context
			client dynamic.Interface
		)

		BeforeEach(func() {
			sb = &specapi.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sb1",
					Namespace: "ns1",
					UID:       "uid1",
				},
			}
			sb.SetGroupVersionKind(specapi.GroupVersionKind)
			u, _ := converter.ToUnstructured(&sb)
			client = fake.NewSimpleDynamicClient(runtime.NewScheme(), u)

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

			u, err := client.Resource(specapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := specapi.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Binding).To(BeNil())
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

		It("should update application if changed", func() {
			sb.Spec.Application = specapi.ServiceBindingApplicationReference{
				APIVersion: "app/v1",
				Kind:       "Foo",
				Name:       "app1",
			}
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			gvr := schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			ref := sb.Spec.Application
			typeLookup.EXPECT().ResourceForReferable(&ref).Return(&gvr, nil)

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

			u, err = client.Resource(specapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := specapi.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Binding.Name).NotTo(BeEmpty())
			Expect(updatedSB.Status.Conditions).To(HaveLen(1))
			Expect(updatedSB.Status.Conditions[0].Type).To(Equal(apis.BindingReady))
			Expect(updatedSB.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))

			u, err = ctx.ReadSecret(sb.Namespace, updatedSB.Status.Binding.Name)
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
			app := specapi.ServiceBindingApplicationReference{
				APIVersion: "app/v1",
				Kind:       "Foo",
				Name:       "app1",
			}
			sb.Spec.Application = app
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo", Value: "v1"})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo2", Value: "v2"})

			gvr := schema.GroupVersionResource{Group: "app", Version: "v1", Resource: "foos"}
			typeLookup.EXPECT().ResourceForReferable(&app).Return(&gvr, nil)

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

			_, err = client.Resource(specapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).To(HaveOccurred())

			u, err = ctx.ReadSecret(sb.Namespace, sb.Status.Binding.Name)
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
			service := pipelinemocks.NewMockService(mockCtrl)

			ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret, Service: service})

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(specapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := specapi.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Binding.Name).Should(Equal(secret.GetName()))

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
			service := pipelinemocks.NewMockService(mockCtrl)

			ctx.AddBindings(&pipeline.SecretBackedBindings{Secret: secret, Service: service})
			ctx.AddBindingItem(&pipeline.BindingItem{Name: "foo3", Value: "val3"})

			err := ctx.Close()
			Expect(err).NotTo(HaveOccurred())

			u, err := client.Resource(specapi.GroupVersionResource).Namespace(sb.Namespace).Get(context.Background(), sb.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			updatedSB := specapi.ServiceBinding{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)
			Expect(err).NotTo(HaveOccurred())

			Expect(updatedSB.Status.Binding.Name).ShouldNot(Equal(secret.GetName()))
			Expect(updatedSB.Status.Binding.Name).ShouldNot(BeEmpty())

			u, err = ctx.ReadSecret(sb.Namespace, sb.Status.Binding.Name)
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

	Describe("EnvBindings", func() {
		var (
			sb  *specapi.ServiceBinding
			ctx pipeline.Context
		)

		It("should return what specified in .spec.enf", func() {
			sb = &specapi.ServiceBinding{
				Spec: specapi.ServiceBindingSpec{
					Env: []specapi.EnvMapping{
						{
							Name: "e1",
							Key:  "b1",
						},
						{
							Name: "e2",
							Key:  "b2",
						},
					},
				},
			}
			ctx, _ = Provider(nil, nil, nil).Get(sb)

			Expect(ctx.EnvBindings()).To(ConsistOf(&pipeline.EnvBinding{Var: "e1", Name: "b1"}, &pipeline.EnvBinding{Var: "e2", Name: "b2"}))
		})

	})
})
