package service

import (
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/pkg/binding/registry"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes/mocks"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/testing"
	"strings"
)

var _ = Describe("Service", func() {

	var (
		client     *fake.FakeDynamicClient
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

	DescribeTable("CRD exist", func(version string, gr schema.GroupResource) {
		crd := crd(version, gr)
		client = fake.NewSimpleDynamicClient(runtime.NewScheme(), crd)
		gvr := gr.WithVersion(version)
		typeLookup.EXPECT().ResourceForKind(gomock.Any()).Return(&gvr, nil)

		builer := NewBuilder(typeLookup).WithClient(client)
		service, err := builer.Build(&unstructured.Unstructured{})
		Expect(err).NotTo(HaveOccurred())

		res, err := service.CustomResourceDefinition()
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Resource()).To(Equal(crd))
	},
		Entry("v1 crd", "v1", schema.GroupResource{Group: "foo", Resource: "bar"}),
		Entry("v1beta crd", "v1beta1", schema.GroupResource{Group: "foo", Resource: "bar"}),
	)

	It("should contain bindable annotations if listed in registry", func() {
		gvk := schema.GroupVersionKind{Group: "postgres-operator.crunchydata.com", Version: "v1beta1", Kind: "PostgresCluster"}
		gvr := schema.GroupVersionResource{Group: "postgres-operator.crunchydata.com", Version: "v1beta1", Resource: "postgresclusters"}
		crd := crd(gvr.Version, gvr.GroupResource())
		err := unstructured.SetNestedField(crd.Object, gvk.Kind, "spec", "names", "kind")
		Expect(err).NotTo(HaveOccurred())

		client = fake.NewSimpleDynamicClient(runtime.NewScheme(), crd)
		typeLookup.EXPECT().ResourceForKind(gvk).Return(&gvr, nil)

		builer := NewBuilder(typeLookup).WithClient(client)
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk)
		service, err := builer.Build(u)
		Expect(err).NotTo(HaveOccurred())

		serviceAnns, _ := registry.ServiceAnnotations.GetAnnotations(gvk)
		res, err := service.CustomResourceDefinition()
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Resource()).NotTo(Equal(crd))
		Expect(res.Resource().GetAnnotations()).To(Equal(serviceAnns))
	})

	It("should return nil when no crd exist", func() {
		client = fake.NewSimpleDynamicClient(runtime.NewScheme())
		typeLookup.EXPECT().ResourceForKind(gomock.Any()).Return(&schema.GroupVersionResource{Group: "app", Resource: "deployments", Version: "v1"}, nil)
		builer := NewBuilder(typeLookup).WithClient(client)
		service, err := builer.Build(&unstructured.Unstructured{})
		Expect(err).NotTo(HaveOccurred())

		res, err := service.CustomResourceDefinition()
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(BeNil())
	})

	Describe("OwnedResources", func() {
		It("should return owned resources", func() {
			id := uuid.NewUUID()
			id2 := uuid.NewUUID()
			ns := "ns1"
			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(schema.GroupVersionKind{Group: "foo.bar", Version: "v1", Kind: "Foo"})
			u.SetName("foo")
			u.SetNamespace(ns)
			u.SetUID(id)

			var children []interface{}

			client = fake.NewSimpleDynamicClient(runtime.NewScheme())

			for i := range bindableResourceGVRs {
				gvr := bindableResourceGVRs[i]

				if gvr.Resource == "configmaps" {
					client.PrependReactor("list", gvr.Resource, func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, k8serrors.NewNotFound(gvr.GroupResource(), "foo")
					})
					continue
				}
				ul := &unstructured.UnstructuredList{}
				// compute kind
				kind := strings.Title(gvr.Resource)[:len(gvr.Resource)-1]

				gvk := gvr.GroupVersion().WithKind(kind)
				ou := resource(gvk, "child1", ns, id)

				ou2 := resource(gvk, "child2", ns, id2)

				children = append(children, ou)
				ul.Items = append(ul.Items, *ou, *ou2)

				client.PrependReactor("list", gvr.Resource, func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, ul, nil
				})
			}

			impl := &service{client: client, resource: u, lookForOwnedResources: true, namespace: ns}

			ownedResources, err := impl.OwnedResources()
			Expect(err).NotTo(HaveOccurred())
			Expect(ownedResources).Should(HaveLen(len(bindableResourceGVRs) - 1))
			Expect(ownedResources).Should(ConsistOf(children...))
		})

		DescribeTable("return error if occurs at looking at owned resources", func(failingResourceName string) {
			id := uuid.NewUUID()
			id2 := uuid.NewUUID()
			ns := "ns1"
			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(schema.GroupVersionKind{Group: "foo.bar", Version: "v1", Kind: "Foo"})
			u.SetName("foo")
			u.SetNamespace(ns)
			u.SetUID(id)

			client = fake.NewSimpleDynamicClient(runtime.NewScheme())
			expectedErr := errors.New("foo")

			for _, gvr := range bindableResourceGVRs {
				ul := &unstructured.UnstructuredList{}
				// compute kind
				kind := strings.Title(gvr.Resource)[:len(gvr.Resource)-1]
				// fix for ConfigMap
				if kind == "Configmap" {
					kind = "ConfigMap"
				}
				gvk := gvr.GroupVersion().WithKind(kind)
				ou := resource(gvk, "child1", ns, id)

				ou2 := resource(gvk, "child2", ns, id2)

				ul.Items = append(ul.Items, *ou, *ou2)

				resourceName := gvr.Resource
				client.PrependReactor("list", resourceName, func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					if failingResourceName == resourceName {
						return true, nil, expectedErr
					}
					return true, ul, nil
				})
			}

			impl := &service{client: client, resource: u, lookForOwnedResources: true, namespace: ns}

			ownedResources, err := impl.OwnedResources()
			Expect(err).Should(Equal(expectedErr))
			Expect(ownedResources).Should(BeNil())
		},
			Entry("fail listing configmap", "configmaps"),
			Entry("fail listing services", "services"),
		)
	})

})

func resource(gvk schema.GroupVersionKind, name string, namespace string, owner types.UID) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	u.SetName(name)
	u.SetNamespace(namespace)
	u.SetOwnerReferences([]v1.OwnerReference{
		{
			UID: owner,
		},
	})
	return u
}

func crd(version string, gr schema.GroupResource) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: version, Kind: "CustomResourceDefinition"})
	u.SetName(gr.String())
	return u
}

var _ = Describe("Builder", func() {
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

	It("should use custom CRD reader set on builder", func() {
		gvk := schema.GroupVersionKind{Kind: "Foo"}
		gvr := schema.GroupVersionResource{Resource: "foo"}
		typeLookup.EXPECT().ResourceForKind(gvk).Return(&gvr, nil)
		crdResource := &unstructured.Unstructured{}
		crdResource.SetName("crdfoo1")
		u := &unstructured.Unstructured{}
		u.SetName("foo1")
		u.SetGroupVersionKind(gvk)
		client := fake.NewSimpleDynamicClient(runtime.NewScheme())
		builder := NewBuilder(typeLookup).WithClient(client).WithCrdReader(func(gvk *schema.GroupVersionResource) (*unstructured.Unstructured, error) {
			return crdResource, nil
		})
		s, err := builder.Build(u)

		Expect(err).NotTo(HaveOccurred())
		Expect(s.Resource()).To(Equal(u))
		crd, err := s.CustomResourceDefinition()
		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Resource()).To(Equal(crdResource))
	})

	It("should use custom CRD reader set on build", func() {
		gvk := schema.GroupVersionKind{Kind: "Foo"}
		gvr := schema.GroupVersionResource{Resource: "foo"}
		typeLookup.EXPECT().ResourceForKind(gvk).Return(&gvr, nil)
		crdResource := &unstructured.Unstructured{}
		crdResource.SetName("crdfoo1")
		u := &unstructured.Unstructured{}
		u.SetName("foo1")
		u.SetGroupVersionKind(gvk)
		client := fake.NewSimpleDynamicClient(runtime.NewScheme())
		builder := NewBuilder(typeLookup).WithClient(client)

		s, err := builder.Build(u, CrdReaderOption(func(gvk *schema.GroupVersionResource) (*unstructured.Unstructured, error) {
			return crdResource, nil
		}))

		Expect(err).NotTo(HaveOccurred())
		Expect(s.Resource()).To(Equal(u))
		crd, err := s.CustomResourceDefinition()
		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Resource()).To(Equal(crdResource))
	})
})
