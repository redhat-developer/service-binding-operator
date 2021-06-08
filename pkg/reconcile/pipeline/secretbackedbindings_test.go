package pipeline_test

import (
	"encoding/base64"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/mocks"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("SecretBackedBindings", func() {
	var (
		mockCtrl *gomock.Controller
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	It("should return items contained in secret", func() {
		secret := &unstructured.Unstructured{Object: map[string]interface{}{
			"data": map[string]interface{}{
				"foo1": base64.StdEncoding.EncodeToString([]byte("val1")),
				"foo2": base64.StdEncoding.EncodeToString([]byte("val2")),
			},
		}}
		service := mocks.NewMockService(mockCtrl)

		b := &pipeline.SecretBackedBindings{Secret: secret, Service: service}

		items, err := b.Items()
		Expect(err).NotTo(HaveOccurred())
		Expect(items).Should(ConsistOf(&pipeline.BindingItem{Name: "foo1", Value: "val1", Source: service}, &pipeline.BindingItem{Name: "foo2", Value: "val2", Source: service}))
	})

	It("should return no items for secret with no data", func() {
		secret := &unstructured.Unstructured{Object: map[string]interface{}{
			"data": map[string]interface{}{},
		}}

		b := &pipeline.SecretBackedBindings{Secret: secret}
		items, err := b.Items()
		Expect(err).NotTo(HaveOccurred())
		Expect(items).Should(BeEmpty())
	})

	It("should return reference to secret when Items() was not invoked", func() {
		secret := &unstructured.Unstructured{}
		secret.SetName("foo")
		secret.SetNamespace("ns1")
		secret.SetAPIVersion("v1")
		secret.SetKind("Secret")

		b := &pipeline.SecretBackedBindings{Secret: secret}

		ref := b.Source()

		Expect(ref).Should(Equal(&v1.ObjectReference{APIVersion: "v1", Kind: "Secret", Name: secret.GetName(), Namespace: secret.GetNamespace()}))
	})

	It("should return no ref when an item key is modified", func() {
		secret := &unstructured.Unstructured{Object: map[string]interface{}{
			"data": map[string]interface{}{
				"foo1": base64.StdEncoding.EncodeToString([]byte("val1")),
				"foo2": base64.StdEncoding.EncodeToString([]byte("val2")),
			},
		}}
		service := mocks.NewMockService(mockCtrl)

		b := &pipeline.SecretBackedBindings{Secret: secret, Service: service}

		items, _ := b.Items()
		items[0].Name = "bla"

		ref := b.Source()

		Expect(ref).To(BeNil())
	})

	It("should return ref if no item key is modified", func() {
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
		service := mocks.NewMockService(mockCtrl)

		b := &pipeline.SecretBackedBindings{Secret: secret, Service: service}

		items, err := b.Items()
		Expect(err).NotTo(HaveOccurred())
		Expect(items).Should(HaveLen(2))

		ref := b.Source()

		Expect(ref).To(Equal(&v1.ObjectReference{APIVersion: "v1", Kind: "Secret", Name: secret.GetName(), Namespace: secret.GetNamespace()}))
	})

})
