package service

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/pkg/binding"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	v1apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/golang/mock/gomock"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

var _ = Describe("CRD", func() {

	var (
		client   *fake.FakeDynamicClient
		mockCtrl *gomock.Controller
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		schema := runtime.NewScheme()
		Expect(olmv1alpha1.AddToScheme(schema)).NotTo(HaveOccurred())
		client = fake.NewSimpleDynamicClient(schema)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should be bindable if marked as provisioned service", func() {
		u := &unstructured.Unstructured{}
		annotations := map[string]string{
			binding.ProvisionedServiceAnnotationKey: "true",
		}
		u.SetAnnotations(annotations)
		crd := &customResourceDefinition{resource: u, client: client}

		Expect(crd.IsBindable()).To(BeTrue())
	})

	DescribeTable("should be bindable if has binding annotation", func(annKey string) {
		u := &unstructured.Unstructured{}
		annotations := map[string]string{
			annKey: "path={.spec}",
			"foo":  "bar",
		}
		u.SetAnnotations(annotations)
		crd := &customResourceDefinition{resource: u, client: client}

		Expect(crd.IsBindable()).To(BeTrue())
	},
		Entry("service.binding", "service.binding"),
		Entry("service.binding/foo", "service.binding/foo"),
	)
	It("should not be bindable if there are no annotations", func() {
		crd := &customResourceDefinition{resource: &unstructured.Unstructured{}, client: client}
		Expect(crd.IsBindable()).To(BeFalse())
	})

	It("should not be bindable if there are no service binding annotations", func() {
		u := &unstructured.Unstructured{}
		annotations := map[string]string{
			"foo": "bar",
		}
		u.SetAnnotations(annotations)
		crd := &customResourceDefinition{resource: u, client: client}
		Expect(crd.IsBindable()).To(BeFalse())
	})

	It("should be bindable if there are service binding descriptors on CSV", func() {
		res := &v1apiextensions.CustomResourceDefinition{}
		res.Spec.Names.Kind = "Foo"
		res.Name = "foo"
		u, err := converter.ToUnstructured(res)
		Expect(err).NotTo(HaveOccurred())

		serviceGVR := schema.GroupVersionResource{Version: "v1"}

		csv := &olmv1alpha1.ClusterServiceVersion{}
		csv.Spec.CustomResourceDefinitions.Owned = []olmv1alpha1.CRDDescription{
			{
				Name:    u.GetName(),
				Version: serviceGVR.Version,
				Kind:    "Foo",
				StatusDescriptors: []olmv1alpha1.StatusDescriptor{
					{
						Path:         "foo",
						XDescriptors: []string{"urn:alm:descriptor:io.kubernetes:Secret", "service.binding:username:sourceValue=username"},
					},
					{
						Path:         "bar",
						XDescriptors: []string{"bar"},
					},
					{
						Path:         "foo2",
						XDescriptors: []string{"urn:alm:descriptor:io.kubernetes:Secret", "service.binding:username2:sourceValue=username"},
					},
				},
			},
		}
		ucsv, err := converter.ToUnstructured(csv)
		Expect(err).NotTo(HaveOccurred())
		ucsv.SetGroupVersionKind(olmv1alpha1.SchemeGroupVersion.WithKind("ClusterServiceVersion"))
		ns := "ns1"
		ucsv.SetName("foo")
		ucsv.SetNamespace(ns)
		client = fake.NewSimpleDynamicClient(runtime.NewScheme(), ucsv)

		crd := &customResourceDefinition{resource: u, client: client, ns: ns, serviceGVR: &serviceGVR, useOlmAnnotations: true}

		Expect(crd.IsBindable()).To(BeTrue())
	})
})
