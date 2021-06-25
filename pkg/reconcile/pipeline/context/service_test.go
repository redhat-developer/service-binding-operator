package context

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	v1alpha12 "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
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
		client *fake.FakeDynamicClient
	)

	DescribeTable("CRD exist", func(version string, gr schema.GroupResource) {
		crd := crd(version, gr)
		client = fake.NewSimpleDynamicClient(runtime.NewScheme(), crd)
		gvr := gr.WithVersion(version)
		ns := "n1"
		service := &service{client: client, serviceRef: &v1alpha12.Service{NamespacedRef: v1alpha12.NamespacedRef{
			Namespace: &ns,
		}}, groupVersionResource: &gvr}

		res, err := service.CustomResourceDefinition()
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Resource()).To(Equal(crd))
	},
		Entry("v1 crd", "v1", schema.GroupResource{Group: "foo", Resource: "bar"}),
		Entry("v1beta crd", "v1beta1", schema.GroupResource{Group: "foo", Resource: "bar"}),
	)
	It("should return nil when no crd exist", func() {
		client = fake.NewSimpleDynamicClient(runtime.NewScheme())

		service := &service{client: client, groupVersionResource: &schema.GroupVersionResource{Group: "app", Resource: "deployments", Version: "v1"}}
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

			impl := &service{client: client, resource: u, lookForOwnedResources: true, serviceRef: &v1alpha12.Service{NamespacedRef: v1alpha12.NamespacedRef{Namespace: &ns}}}

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

			impl := &service{client: client, resource: u, lookForOwnedResources: true, serviceRef: &v1alpha12.Service{NamespacedRef: v1alpha12.NamespacedRef{Namespace: &ns}}}

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
