package context

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

var _ = Describe("Service", func() {

	var (
		client dynamic.Interface
	)

	DescribeTable("CRD exist", func(version string, gr schema.GroupResource) {
		crd := crd(version, gr)
		client = fake.NewSimpleDynamicClient(runtime.NewScheme(), crd)
		gvr := gr.WithVersion(version)
		ns := "n1"
		service := &service{client: client, serviceRef: &v1alpha1.Service{NamespacedRef: v1alpha1.NamespacedRef{
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

})

func crd(version string, gr schema.GroupResource) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: version, Kind: "CustomResourceDefinition"})
	u.SetName(gr.String())
	return u
}
