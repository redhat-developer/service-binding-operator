package v1alpha2

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validation Webhook", func() {

	It("should allow updates on non-ready bindings", func() {
		sb := &ServiceBinding{
			Spec: ServiceBindingSpec{},
			Status: ServiceBindingStatus{
				Conditions: []v1.Condition{*apis.Conditions().NotBindingReady().Build()},
			},
		}
		Expect(sb.ValidateUpdate(sb)).ShouldNot(HaveOccurred())
	})

	It("should not allow spec updates on ready bindings", func() {
		old := &ServiceBinding{
			Spec: ServiceBindingSpec{},
			Status: ServiceBindingStatus{
				Conditions: []v1.Condition{*apis.Conditions().BindingReady().Build()},
			},
		}
		sb := old.DeepCopy()
		sb.Spec.Name = "foo"
		Expect(sb.ValidateUpdate(old)).Should(HaveOccurred())
	})

	It("should allow metadata updates on ready bindings", func() {
		old := &ServiceBinding{
			Spec: ServiceBindingSpec{},
			Status: ServiceBindingStatus{
				Conditions: []v1.Condition{*apis.Conditions().BindingReady().Build()},
			},
		}
		sb := old.DeepCopy()
		sb.Annotations = map[string]string{
			"foo": "bar",
		}
		Expect(sb.ValidateUpdate(old)).ShouldNot(HaveOccurred())
	})
})
