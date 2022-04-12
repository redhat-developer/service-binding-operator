package context

import (
	"fmt"
	"strings"

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
	"k8s.io/client-go/util/jsonpath"
)

func jsonPath(path string) *jsonpath.JSONPath {
	jp := jsonpath.New("")
	str := fmt.Sprintf("{%s}", path)
	if err := jp.Parse(str); err != nil {
		Fail(fmt.Sprintf("Couldn't parse jsonpath {%s}: %v", path, err))
	}
	return jp
}

var deploymentWorkloadMapping pipeline.WorkloadMapping = pipeline.WorkloadMapping{
	Volume: strings.Split("spec.template.spec.volumes", "."),
	Containers: []pipeline.WorkloadContainer{
		{
			Path:         jsonPath(".spec.template.spec.containers[*]"),
			Name:         []string{"name"},
			Env:          []string{"env"},
			EnvFrom:      []string{"envFrom"},
			VolumeMounts: []string{"volumeMounts"},
		},
		{
			Path:         jsonPath(".spec.template.spec.initContainers[*]"),
			Name:         []string{"name"},
			Env:          []string{"env"},
			EnvFrom:      []string{"envFrom"},
			VolumeMounts: []string{"volumeMounts"},
		},
	},
}

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
			Name:  "c1",
		}
		c2 := corev1.Container{
			Image: "foo2",
			Name:  "c2",
		}
		d1 := deployment("d1", []corev1.Container{c1, c2})
		u, _ := converter.ToUnstructured(&d1)
		cu1, _ := converter.ToUnstructured(&c1)
		cu2, _ := converter.ToUnstructured(&c2)
		mct := pipeline.MetaPodSpec{
			Data:   u.Object,
			Volume: strings.Split("spec.template.spec.volumes", "."),
			Containers: []pipeline.MetaContainer{
				{
					Data:        cu1.Object,
					Name:        "c1",
					Env:         []string{"env"},
					EnvFrom:     []string{"envFrom"},
					VolumeMount: []string{"volumeMounts"},
				},
				{
					Data:        cu2.Object,
					Name:        "c2",
					Env:         []string{"env"},
					EnvFrom:     []string{"envFrom"},
					VolumeMount: []string{"volumeMounts"},
				},
			},
		}
		app := &application{persistedResource: u, resourceMapping: deploymentWorkloadMapping}
		containers, err := app.BindablePods()
		Expect(err).NotTo(HaveOccurred())
		Expect(*containers).To(Equal(mct))
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
		mct := pipeline.MetaPodSpec{
			Data:   u.Object,
			Volume: strings.Split("spec.template.spec.volumes", "."),
			Containers: []pipeline.MetaContainer{
				{
					Data:        cu2.Object,
					Name:        "c2",
					Env:         []string{"env"},
					EnvFrom:     []string{"envFrom"},
					VolumeMount: []string{"volumeMounts"},
				},
				{
					Data:        cu3.Object,
					Name:        "c3",
					Env:         []string{"env"},
					EnvFrom:     []string{"envFrom"},
					VolumeMount: []string{"volumeMounts"},
				},
			},
		}
		app := &application{persistedResource: u, bindableContainerNames: sets.NewString("c2", "c3", "c1"), resourceMapping: deploymentWorkloadMapping}
		containers, err := app.BindablePods()
		Expect(err).NotTo(HaveOccurred())
		Expect(*containers).To(Equal(mct))
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
