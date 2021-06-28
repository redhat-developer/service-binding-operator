package context

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
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
		app = &application{bindingPath: &v1alpha1.BindingPath{SecretPath: "foo"}}
		Expect(app.SecretPath()).To(Equal("foo"))
	})

	It("should return default container path if binding path not set", func() {
		app = &application{}
		Expect(app.ContainersPath()).To(Equal(defaultContainerPath))
	})

	It("should return default container path if binding path container path is set to empty", func() {
		app = &application{bindingPath: &v1alpha1.BindingPath{}}
		Expect(app.ContainersPath()).To(Equal(defaultContainerPath))
	})

	It("should return container path set through binding path", func() {
		app = &application{bindingPath: &v1alpha1.BindingPath{ContainersPath: "foo"}}
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

	It("should return all containers if bindable position are not specified", func() {
		c1 := corev1.Container{
			Image: "foo",
		}
		c2 := corev1.Container{
			Image: "foo2",
		}
		d1 := deployment("d1", []corev1.Container{c1, c2})
		u, _ := converter.ToUnstructured(&d1)
		cu1, _ := converter.ToUnstructured(&c1)
		cu2, _ := converter.ToUnstructured(&c2)
		app := &application{persistedResource: u}
		containers, err := app.BindableContainers()
		Expect(err).NotTo(HaveOccurred())
		Expect(containers).To(ConsistOf(cu1.Object, cu2.Object))
	})

	It("should return only containers which names are specified in bindable names", func() {
		c1 := corev1.Container{
			Image: "foo",
		}
		c2 := corev1.Container{
			Name:  "c2",
			Image: "foo2",
		}
		c3 := corev1.Container{
			Name:  "c3",
			Image: "foo3",
		}
		d1 := deployment("d1", []corev1.Container{c1, c2, c3})
		u, _ := converter.ToUnstructured(&d1)
		cu2, _ := converter.ToUnstructured(&c2)
		cu3, _ := converter.ToUnstructured(&c3)
		app := &application{persistedResource: u, bindableContainerNames: sets.NewString("c2", "c3", "c1")}
		containers, err := app.BindableContainers()
		Expect(err).NotTo(HaveOccurred())
		Expect(containers).To(ConsistOf(cu2.Object, cu3.Object))
	})
})

func deployment(name string, containers []corev1.Container) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: containers,
				},
			},
		},
	}
}
