package context

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1alpha12 "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Application", func() {
	var (
		app pipeline.Application
	)
	It("should return empty secret path if binding path not set", func() {
		app = &application{}
		Expect(app.SecretPath()).To(BeEmpty())
	})

	It("should return secret path by provided binding path", func() {
		app = &application{bindingPath: &v1alpha12.BindingPath{SecretPath: "foo"}}
		Expect(app.SecretPath()).To(Equal("foo"))
	})

	It("should return default container path if binding path not set", func() {
		app = &application{}
		Expect(app.ContainersPath()).To(Equal(defaultContainerPath))
	})

	It("should return default container path if binding path container path is set to empty", func() {
		app = &application{bindingPath: &v1alpha12.BindingPath{}}
		Expect(app.ContainersPath()).To(Equal(defaultContainerPath))
	})

	It("should return container path set through binding path", func() {
		app = &application{bindingPath: &v1alpha12.BindingPath{ContainersPath: "foo"}}
		Expect(app.ContainersPath()).To(Equal("foo"))
	})

	It("should flag resource as updated if modified", func() {
		u := &unstructured.Unstructured{}
		u.SetNamespace("foo")
		app := &application{persistedResource: u}
		app.Resource().SetName("bar")
		Expect(app.IsUpdated()).To(BeTrue())
	})

	It("should flag resource as not updated if not modified", func() {
		u := &unstructured.Unstructured{}
		u.SetNamespace("foo")
		app := &application{persistedResource: u}
		_ = app.Resource()
		Expect(app.IsUpdated()).To(BeFalse())
	})
})
