package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validation Webhook", func() {

	It("should allow updates on non-ready bindings", func() {
		sb := &ServiceBinding{
			Status: ServiceBindingStatus{
				Conditions: []v1.Condition{*apis.Conditions().NotBindingReady().Build()},
			},
		}
		Expect(sb.ValidateUpdate(sb)).ShouldNot(HaveOccurred())
	})

	It("should not allow updates on ready bindings", func() {
		sb := &ServiceBinding{
			Status: ServiceBindingStatus{
				Conditions: []v1.Condition{*apis.Conditions().BindingReady().Build()},
			},
		}
		Expect(sb.ValidateUpdate(sb)).Should(HaveOccurred())
	})
})
