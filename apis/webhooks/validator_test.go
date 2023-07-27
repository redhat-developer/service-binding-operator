package webhooks

import (
	"context"
	"encoding/json"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/apis/spec/v1beta1"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var _ = Describe("Validate", func() {
	var (
		mockCtrl  *gomock.Controller
		validator MappingValidator
		ctx       context.Context
		mapping   v1beta1.ClusterWorkloadResourceMapping
		scheme    *runtime.Scheme
		lookup    *mocks.MockK8STypeLookup
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		scheme = runtime.NewScheme()
		Expect(v1beta1.AddToScheme(scheme)).NotTo(HaveOccurred())
		Expect(v1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
		ctx = context.Background()
		mapping = v1beta1.ClusterWorkloadResourceMapping{
			Spec: v1beta1.ClusterWorkloadResourceMappingSpec{
				Versions: []v1beta1.ClusterWorkloadResourceMappingTemplate{
					v1beta1.DefaultTemplate,
				},
			},
		}
		lookup = mocks.NewMockK8STypeLookup(mockCtrl)
		validator = MappingValidator{
			client: nil,
			lookup: lookup,
		}
	})

	var _ = Describe("Create", func() {
		It("should accept a valid mapping", func() {
			msg, err := validator.ValidateCreate(ctx, &mapping)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(BeEmpty())
		})

		It("should reject an invalid mapping", func() {
			mapping.Spec.Versions[0].Volumes = "foo.bar"
			msg, err := validator.ValidateCreate(ctx, &mapping)
			Expect(err).To(HaveOccurred())
			Expect(msg).To(BeEmpty())
		})
	})

	var _ = Describe("Update", func() {
		It("should reject an invalid mapping", func() {
			oldMapping := mapping.DeepCopy()
			mapping.Spec.Versions[0].Volumes = "foo.bar"
			msg, err := validator.ValidateUpdate(ctx, oldMapping, &mapping)
			Expect(err).To(HaveOccurred())
			Expect(msg).To(BeEmpty())
		})

		It("should accept a valid mapping", func() {
			oldMapping := mapping.DeepCopy()
			mapping.Spec.Versions[0].Volumes = ".spec.volumes"
			validator.client = fake.NewSimpleDynamicClient(scheme)
			msg, err := validator.ValidateUpdate(ctx, oldMapping, &mapping)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(BeEmpty())
		})

		It("should serialize old mappings into relevant service bindings", func() {
			mapping.Name = "foos.bar"
			oldMapping := mapping.DeepCopy()
			mapping.Spec.Versions[0].Volumes = ".spec.volumes"
			relevantSB := v1beta1.ServiceBinding{
				ObjectMeta: v1.ObjectMeta{
					Name: "relevant",
				},
				Spec: v1beta1.ServiceBindingSpec{
					Workload: v1beta1.ServiceBindingWorkloadReference{
						APIVersion: "bar/v1",
						Kind:       "Foo",
						Name:       "x",
					},
				},
			}
			irrelevantSB := v1beta1.ServiceBinding{
				ObjectMeta: v1.ObjectMeta{
					Name: "irrelevant",
				},
				Spec: v1beta1.ServiceBindingSpec{
					Workload: v1beta1.ServiceBindingWorkloadReference{
						APIVersion: "bar/v1",
						Kind:       "Spam",
						Name:       "x",
					},
				},
			}
			validator.client = fake.NewSimpleDynamicClient(scheme, &relevantSB, &irrelevantSB)
			spamGVK := schema.GroupVersionKind{Group: "bar", Version: "v1", Kind: "Spam"}
			spamGVR := schema.GroupVersionResource{Group: "bar", Version: "v1", Resource: "spams"}
			lookup.EXPECT().ResourceForKind(spamGVK).Return(&spamGVR, nil).Times(2)
			fooGVK := schema.GroupVersionKind{Group: "bar", Version: "v1", Kind: "Foo"}
			fooGVR := schema.GroupVersionResource{Group: "bar", Version: "v1", Resource: "foos"}
			lookup.EXPECT().ResourceForKind(fooGVK).Return(&fooGVR, nil).Times(2)

			msg, err := validator.ValidateUpdate(ctx, oldMapping, &mapping)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(BeEmpty())

			oldData, err := json.Marshal(oldMapping)
			Expect(err).NotTo(HaveOccurred())

			data, err := validator.client.Resource(v1beta1.GroupVersionResource).
				Get(ctx, "relevant", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(data.GetAnnotations()).To(Equal(map[string]string{apis.MappingAnnotationKey: string(oldData)}))

			data, err = validator.client.Resource(v1beta1.GroupVersionResource).
				Get(ctx, "irrelevant", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(data.GetAnnotations()).To(BeEmpty())
		})
	})

	var _ = Describe("Delete", func() {
		It("should not need to validate deletes", func() {
			msg, err := validator.ValidateDelete(ctx, &mapping)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(BeEmpty())
		})
	})
})

var _ = Describe("Serialize", func() {

	var (
		mockCtrl *gomock.Controller
		binding1 v1beta1.ServiceBinding
		binding2 v1alpha1.ServiceBinding
		mapping  v1beta1.ClusterWorkloadResourceMapping
		lookup   *mocks.MockK8STypeLookup
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		binding1 = v1beta1.ServiceBinding{
			ObjectMeta: v1.ObjectMeta{Name: "foo", Namespace: "bar"},
			Spec: v1beta1.ServiceBindingSpec{
				Workload: v1beta1.ServiceBindingWorkloadReference{
					APIVersion: "foo.bar/v1",
					Kind:       "Baz",
					Name:       "spam",
				},
			},
		}

		binding2 = v1alpha1.ServiceBinding{
			ObjectMeta: v1.ObjectMeta{Name: "foo", Namespace: "bar"},
			Spec: v1alpha1.ServiceBindingSpec{
				Application: v1alpha1.Application{
					Ref: v1alpha1.Ref{
						Group:   "foo.bar",
						Version: "v1",
						Kind:    "Baz",
						Name:    "spam",
					},
				},
			},
		}

		mapping = v1beta1.ClusterWorkloadResourceMapping{
			ObjectMeta: v1.ObjectMeta{Name: "baz.foo.bar"},
			Spec: v1beta1.ClusterWorkloadResourceMappingSpec{
				Versions: []v1beta1.ClusterWorkloadResourceMappingTemplate{
					v1beta1.DefaultTemplate,
				},
			},
		}
		lookup = mocks.NewMockK8STypeLookup(mockCtrl)
		lookup.EXPECT().ResourceForKind(schema.GroupVersionKind{
			Group:   "foo.bar",
			Kind:    "Baz",
			Version: "v1",
		}).Return(&schema.GroupVersionResource{
			Group:    "foo.bar",
			Resource: "baz",
			Version:  "v1",
		}, nil).AnyTimes()
	})

	It("should serialize the mapping into annotations", func() {
		scheme := runtime.NewScheme()
		Expect(v1beta1.AddToScheme(scheme)).NotTo(HaveOccurred())
		Expect(v1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
		client := fake.NewSimpleDynamicClient(scheme, &binding1, &binding2, &mapping)

		data, err := json.Marshal(mapping)
		Expect(err).NotTo(HaveOccurred())

		err = Serialize(context.Background(), &mapping, client, lookup)
		Expect(err).NotTo(HaveOccurred())

		bindingUnstructured, err := client.
			Resource(v1beta1.GroupVersionResource).
			Namespace("bar").
			Get(context.Background(), "foo", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(bindingUnstructured.GetAnnotations()[apis.MappingAnnotationKey]).To(Equal(string(data)))

		bindingUnstructured, err = client.
			Resource(v1alpha1.GroupVersionResource).
			Namespace("bar").
			Get(context.Background(), "foo", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(bindingUnstructured.GetAnnotations()[apis.MappingAnnotationKey]).To(Equal(string(data)))
	})

	It("should not modify existing annotations", func() {
		annotations := map[string]string{
			"foo": "bar",
		}
		binding1.SetAnnotations(annotations)
		binding2.SetAnnotations(annotations)

		scheme := runtime.NewScheme()
		Expect(v1beta1.AddToScheme(scheme)).NotTo(HaveOccurred())
		Expect(v1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
		client := fake.NewSimpleDynamicClient(scheme, &binding1, &binding2, &mapping)

		data, err := json.Marshal(mapping)
		Expect(err).NotTo(HaveOccurred())

		err = Serialize(context.Background(), &mapping, client, lookup)
		Expect(err).NotTo(HaveOccurred())

		bindingUnstructured, err := client.
			Resource(v1beta1.GroupVersionResource).
			Namespace("bar").
			Get(context.Background(), "foo", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(bindingUnstructured.GetAnnotations()[apis.MappingAnnotationKey]).To(Equal(string(data)))
		Expect(bindingUnstructured.GetAnnotations()["foo"]).To(Equal("bar"))

		bindingUnstructured, err = client.
			Resource(v1alpha1.GroupVersionResource).
			Namespace("bar").
			Get(context.Background(), "foo", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(bindingUnstructured.GetAnnotations()[apis.MappingAnnotationKey]).To(Equal(string(data)))
		Expect(bindingUnstructured.GetAnnotations()["foo"]).To(Equal("bar"))
	})

	It("should ignore irrelevant bindings", func() {
		mapping.Name = "spam.foo.bar"
		scheme := runtime.NewScheme()
		Expect(v1beta1.AddToScheme(scheme)).NotTo(HaveOccurred())
		Expect(v1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())
		client := fake.NewSimpleDynamicClient(scheme, &binding1, &binding2, &mapping)

		err := Serialize(context.Background(), &mapping, client, lookup)
		Expect(err).NotTo(HaveOccurred())

		bindingUnstructured, err := client.
			Resource(v1beta1.GroupVersionResource).
			Namespace("bar").
			Get(context.Background(), "foo", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(bindingUnstructured.GetAnnotations()).NotTo(HaveKey(apis.MappingAnnotationKey))

		bindingUnstructured, err = client.
			Resource(v1alpha1.GroupVersionResource).
			Namespace("bar").
			Get(context.Background(), "foo", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(bindingUnstructured.GetAnnotations()).NotTo(HaveKey(apis.MappingAnnotationKey))
	})

	It("should use the resource field from the service binding when available", func() {
		mapping.Name = "spam.foo.bar"
		scheme := runtime.NewScheme()
		Expect(v1beta1.AddToScheme(scheme)).NotTo(HaveOccurred())
		Expect(v1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

		binding2.Spec.Application.Kind = ""
		binding2.Spec.Application.Resource = "baz"

		client := fake.NewSimpleDynamicClient(scheme, &binding2, &mapping)

		err := Serialize(context.Background(), &mapping, client, lookup)
		Expect(err).NotTo(HaveOccurred())

		bindingUnstructured, err := client.
			Resource(v1alpha1.GroupVersionResource).
			Namespace("bar").
			Get(context.Background(), "foo", v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(bindingUnstructured.GetAnnotations()).NotTo(HaveKey(apis.MappingAnnotationKey))
	})
})
