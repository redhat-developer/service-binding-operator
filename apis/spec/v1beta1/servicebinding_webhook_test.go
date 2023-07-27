package v1beta1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		_, err := sb.ValidateUpdate(sb)
		Expect(err).ToNot(HaveOccurred())
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
		_, err := sb.ValidateUpdate(old)
		Expect(err).To(HaveOccurred())
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
		_, err := sb.ValidateUpdate(old)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return error if both application name and selecter is specified during creation", func() {

		ls := &metav1.LabelSelector{
			MatchLabels: map[string]string{"env": "prod"},
		}

		ref := ServiceBindingWorkloadReference{
			APIVersion: "app/v1",
			Kind:       "Foo",
			Name:       "app1",
			Selector:   ls,
		}

		sb := &ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sb1",
				Namespace: "ns1",
			},
			Spec: ServiceBindingSpec{
				Workload: ref,
			},
		}
		_, err := sb.ValidateCreate()
		Expect(err).To(HaveOccurred())

	})

	It("should return error if both application name and selecter is specified during update", func() {

		ls := &metav1.LabelSelector{
			MatchLabels: map[string]string{"env": "prod"},
		}

		ref := ServiceBindingWorkloadReference{
			APIVersion: "app/v1",
			Kind:       "Foo",
			Name:       "app1",
			Selector:   ls,
		}

		sb := &ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sb1",
				Namespace: "ns1",
			},
			Spec: ServiceBindingSpec{
				Workload: ref,
			},
		}
		_, err := sb.ValidateUpdate(sb)
		Expect(err).To(HaveOccurred())

	})

})
