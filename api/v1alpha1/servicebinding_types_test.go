package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec", func() {
	var sbSpec ServiceBindingSpec

	Context("Binding as files", func() {
		BeforeEach(func() {
			sbSpec = ServiceBindingSpec{
				BindAsFiles: true,
			}
		})
		It("should use none naming strategy as default", func() {
			Expect(sbSpec.NamingTemplate()).To(Equal(templates["none"]))
		})
		It("should use by name referred strategy", func() {
			sbSpec.NamingStrategy = "uppercase"
			Expect(sbSpec.NamingTemplate()).To(Equal(templates["uppercase"]))
		})
		It("should use provided template", func() {
			sbSpec.NamingStrategy = "{{ foo }}"
			Expect(sbSpec.NamingTemplate()).To(Equal(sbSpec.NamingStrategy))
		})
	})
	Context("Binding as env vars", func() {
		BeforeEach(func() {
			sbSpec = ServiceBindingSpec{
				BindAsFiles: false,
			}
		})
		It("should use uppercase naming strategy as default", func() {
			Expect(sbSpec.NamingTemplate()).To(Equal(templates["uppercase"]))
		})
		It("should use by name referred strategy", func() {
			sbSpec.NamingStrategy = "lowercase"
			Expect(sbSpec.NamingTemplate()).To(Equal(templates["lowercase"]))
		})
		It("should use provided template", func() {
			sbSpec.NamingStrategy = "{{ foo }}"
			Expect(sbSpec.NamingTemplate()).To(Equal(sbSpec.NamingStrategy))
		})
	})
})
