/*
Copyright 2022.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipeline_test

import (
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis/spec/v1alpha3"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
)

func jsonPath(path string) *jsonpath.JSONPath {
	jpath := jsonpath.New("")
	err := jpath.Parse(fmt.Sprintf("{%s}", path))
	Expect(err).NotTo(HaveOccurred())
	return jpath
}

var _ = Describe("Mapping workloads", func() {
	var (
		mockCtrl *gomock.Controller
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("Parsing", func() {
		It("should parse valid mappings", func() {
			template := v1alpha3.ClusterWorkloadResourceMappingTemplate{
				Version:     "*",
				Annotations: ".spec.annotations",
				Containers: []v1alpha3.ClusterWorkloadResourceMappingContainer{
					{
						Name:         ".name",
						Env:          ".env",
						Path:         ".spec.spec.containers[*]",
						VolumeMounts: ".volumeMounts",
					},
				},
				Volumes: ".spec.spec.volumes",
			}
			mapping := pipeline.WorkloadMapping{
				Volume: []string{"spec", "spec", "volumes"},
				Containers: []pipeline.WorkloadContainer{
					{
						Name:         []string{"name"},
						Env:          []string{"env"},
						EnvFrom:      []string{"envFrom"},
						VolumeMounts: []string{"volumeMounts"},
						Path:         jsonPath(".spec.spec.containers[*]"),
					},
				},
			}

			result, err := pipeline.FromWorkloadResourceMappingTemplate(template)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Volume).To(Equal(mapping.Volume))
			Expect(result.Containers[0].Name).To(Equal(mapping.Containers[0].Name))
			Expect(result.Containers[0].Env).To(Equal(mapping.Containers[0].Env))
			Expect(result.Containers[0].EnvFrom).To(Equal(mapping.Containers[0].EnvFrom))
			Expect(result.Containers[0].VolumeMounts).To(Equal(mapping.Containers[0].VolumeMounts))
		})

		It("should use defaults when not specified", func() {
			template := v1alpha3.ClusterWorkloadResourceMappingTemplate{
				Version:     "*",
				Annotations: ".spec.annotations",
				Containers: []v1alpha3.ClusterWorkloadResourceMappingContainer{
					{
						Path: ".spec.spec.containers[*]",
					},
				},
				Volumes: ".spec.spec.volumes",
			}
			mapping := pipeline.WorkloadMapping{
				Volume: []string{"spec", "spec", "volumes"},
				Containers: []pipeline.WorkloadContainer{
					{
						Name:         []string{"name"},
						Env:          []string{"env"},
						EnvFrom:      []string{"envFrom"},
						VolumeMounts: []string{"volumeMounts"},
						Path:         jsonPath(".spec.spec.containers[*]"),
					},
				},
			}

			result, err := pipeline.FromWorkloadResourceMappingTemplate(template)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Volume).To(Equal(mapping.Volume))
			Expect(result.Containers[0].Name).To(Equal(mapping.Containers[0].Name))
			Expect(result.Containers[0].Env).To(Equal(mapping.Containers[0].Env))
			Expect(result.Containers[0].EnvFrom).To(Equal(mapping.Containers[0].EnvFrom))
			Expect(result.Containers[0].VolumeMounts).To(Equal(mapping.Containers[0].VolumeMounts))
		})

		It("should return an error on invalid mappings", func() {
			template := v1alpha3.ClusterWorkloadResourceMappingTemplate{
				Version:     "*",
				Annotations: ".spec.annotations",
				Containers: []v1alpha3.ClusterWorkloadResourceMappingContainer{
					{
						Path: ".spec.spec.containers[",
					},
				},
				Volumes: ".spec.spec.volumes",
			}

			_, err := pipeline.FromWorkloadResourceMappingTemplate(template)
			Expect(err).To(HaveOccurred())
		})

		It("should not allow jsonpaths in restricted jsonpath contexts", func() {
			templates := []v1alpha3.ClusterWorkloadResourceMappingTemplate{
				{
					Version:     "*",
					Annotations: ".spec.annotations",
					Containers: []v1alpha3.ClusterWorkloadResourceMappingContainer{
						{
							Path: ".spec.spec.containers[*]",
							Name: ".name[*]",
						},
					},
				},
				{
					Version:     "*",
					Annotations: ".spec.annotations",
					Containers: []v1alpha3.ClusterWorkloadResourceMappingContainer{
						{
							Path: ".spec.spec.containers[*]",
							Env:  ".env[*]",
						},
					},
				},
				{
					Version:     "*",
					Annotations: ".spec.annotations",
					Containers: []v1alpha3.ClusterWorkloadResourceMappingContainer{
						{
							Path:         ".spec.spec.containers[*]",
							VolumeMounts: ".volumeMounts[*]",
						},
					},
				},
				{
					Version:     "*",
					Annotations: ".spec.annotations",
					Containers: []v1alpha3.ClusterWorkloadResourceMappingContainer{
						{
							Path: ".spec.spec.containers[*]",
						},
					},
					Volumes: ".spec.spec.volumes[*]",
				},
			}

			for _, template := range templates {
				_, err := pipeline.FromWorkloadResourceMappingTemplate(template)
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Context("Environment variables", func() {
		It("should use $SERVICE_BINDING_ROOT if set", func() {
			container := pipeline.MetaContainer{
				Data: map[string]interface{}{
					"env": []map[string]interface{}{
						{
							"name":  "SERVICE_BINDING_ROOT",
							"value": "/tmp/bindings",
						},
					},
				},
				Env: []string{"env"},
			}

			result, err := container.MountPath("foo-bindings")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("/tmp/bindings/foo-bindings"))
			Expect(container.Data["env"]).To(Equal([]map[string]interface{}{{"name": "SERVICE_BINDING_ROOT", "value": "/tmp/bindings"}}))
		})

		It("should set $SERVICE_BINDING_ROOT if not set", func() {
			container := pipeline.MetaContainer{
				Data: map[string]interface{}{},
				Env:  []string{"env"},
			}

			result, err := container.MountPath("foo-bindings")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("/bindings/foo-bindings"))
			Expect(container.Data["env"]).To(Equal([]map[string]interface{}{{"name": "SERVICE_BINDING_ROOT", "value": "/bindings"}}))
		})

		It("should add new environment variables", func() {
			container := pipeline.MetaContainer{
				Data: map[string]interface{}{},
				Env:  []string{"env"},
			}
			envVars := []v1.EnvVar{
				{
					Name:  "FOO",
					Value: "foo",
				},
				{
					Name: "BAR",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "spam",
							},
							Key: "eggs",
						},
					},
				},
			}
			var envs []map[string]interface{}
			for _, env := range envVars {
				data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&env)
				Expect(err).NotTo(HaveOccurred())
				envs = append(envs, data)
			}

			err := container.AddEnvVars(envVars)
			Expect(err).NotTo(HaveOccurred())

			val, found, err := converter.NestedResources(&v1.EnvVar{}, container.Data, container.Env...)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(val).To(ConsistOf(envs))
		})

		It("should remove service binding environment variables", func() {
			envVars := []v1.EnvVar{
				{
					Name:  "FOO",
					Value: "foo",
				},
				{
					Name: "BAR",
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "spam",
							},
							Key: "eggs",
						},
					},
				},
			}
			envs, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&map[string]interface{}{"env": envVars})
			Expect(err).NotTo(HaveOccurred())
			container := pipeline.MetaContainer{
				Data: envs,
				Env:  []string{"env"},
			}

			err = container.RemoveEnvVars("BAR")
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"env": []map[string]interface{}{
					{
						"name":  "FOO",
						"value": "foo",
					},
				},
			}))

			err = container.RemoveEnvVars("FOO")
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"env": []map[string]interface{}{},
			}))
		})

		It("should not set service binding environment variables more than once", func() {
			envVars := []v1.EnvVar{
				{
					Name:  "FOO",
					Value: "foo",
				},
			}

			envs, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&map[string]interface{}{"env": envVars})
			Expect(err).NotTo(HaveOccurred())
			container := pipeline.MetaContainer{
				Data: envs,
				Env:  []string{"env"},
			}

			err = container.AddEnvVars(envVars)
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"env": []map[string]interface{}{
					{
						"name":  "FOO",
						"value": "foo",
					},
				},
			}))
		})

		It("should add new environment variables from a secret", func() {
			container := pipeline.MetaContainer{
				Data:    map[string]interface{}{},
				EnvFrom: []string{"envFrom"},
			}
			envFromVars := v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "spam",
					},
				},
			}
			env, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&envFromVars)
			Expect(err).NotTo(HaveOccurred())

			err = container.AddEnvFromVar(envFromVars)
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{"envFrom": []map[string]interface{}{env}}))
		})

		It("should not set environment variables from a secret more than once", func() {
			container := pipeline.MetaContainer{
				Data:    map[string]interface{}{},
				EnvFrom: []string{"envFrom"},
			}
			envFromVars := v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "spam",
					},
				},
			}
			env, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&envFromVars)
			Expect(err).NotTo(HaveOccurred())

			err = container.AddEnvFromVar(envFromVars)
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{"envFrom": []map[string]interface{}{env}}))

			err = container.AddEnvFromVar(envFromVars)
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{"envFrom": []map[string]interface{}{env}}))
		})

		It("should not add new environment variables from a secret when envFrom is not set", func() {
			container := pipeline.MetaContainer{}
			envFromVars := v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "spam",
					},
				},
			}
			err := container.AddEnvFromVar(envFromVars)
			Expect(err).To(HaveOccurred())
		})

		It("should remove environment variables from a secret", func() {
			container := pipeline.MetaContainer{
				Data: map[string]interface{}{
					"envFrom": []map[string]interface{}{
						{
							"secretRef": map[string]interface{}{
								"name": "spam",
							},
						},
						{
							"secretRef": map[string]interface{}{
								"name": "eggs",
							},
						},
					},
				},
				EnvFrom: []string{"envFrom"},
			}

			err := container.RemoveEnvFromVars("spam")
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"envFrom": []map[string]interface{}{
					{
						"secretRef": map[string]interface{}{
							"name": "eggs",
						},
					},
				},
			}))

			err = container.RemoveEnvFromVars("eggs")
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"envFrom": []map[string]interface{}{},
			}))
		})
	})

	Context("Files", func() {
		It("should add volumes", func() {
			volume := v1.Volume{
				Name: "foo",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "service-binding-data",
					},
				},
			}

			template := pipeline.MetaPodSpec{
				Volume: []string{"spec", "volumes"},
				Data:   map[string]interface{}{},
			}

			err := template.AddVolume(volume)
			Expect(err).NotTo(HaveOccurred())
			Expect(template.Data).To(Equal(map[string]interface{}{
				"spec": map[string]interface{}{
					"volumes": []map[string]interface{}{
						{
							"name": "foo",
							"secret": map[string]interface{}{
								"secretName": "service-binding-data",
							},
						},
					},
				},
			}))
		})

		It("should not alter existing volumes when adding new volumes", func() {
			volume := v1.Volume{
				Name: "foo",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: "service-binding-data",
					},
				},
			}

			template := pipeline.MetaPodSpec{
				Volume: []string{"volumes"},
				Data: map[string]interface{}{
					"volumes": []map[string]interface{}{
						{
							"name": "spam",
							"secret": map[string]interface{}{
								"secretName": "eggs",
							},
						},
					},
				},
			}

			err := template.AddVolume(volume)
			Expect(err).NotTo(HaveOccurred())
			Expect(template.Data).To(Equal(map[string]interface{}{
				"volumes": []map[string]interface{}{
					{
						"name": "spam",
						"secret": map[string]interface{}{
							"secretName": "eggs",
						},
					},
					{
						"name": "foo",
						"secret": map[string]interface{}{
							"secretName": "service-binding-data",
						},
					},
				},
			}))
		})

		It("should remove volumes", func() {
			template := pipeline.MetaPodSpec{
				Volume: []string{"volumes"},
				Data: map[string]interface{}{
					"volumes": []map[string]interface{}{
						{
							"name": "foo",
							"secret": map[string]interface{}{
								"secretName": "service-binding-data",
							},
						},
					},
				},
			}

			err := template.RemoveVolume("foo")
			Expect(err).NotTo(HaveOccurred())
			Expect(template.Data).To(Equal(map[string]interface{}{
				"volumes": []map[string]interface{}{},
			}))
		})

		It("should not alter non-service binding volumes on removal", func() {
			template := pipeline.MetaPodSpec{
				Volume: []string{"volumes"},
				Data: map[string]interface{}{
					"volumes": []map[string]interface{}{
						{
							"name": "spam",
							"secret": map[string]interface{}{
								"secretName": "eggs",
							},
						},
						{
							"name": "foo",
							"secret": map[string]interface{}{
								"secretName": "service-binding-data",
							},
						},
					},
				},
			}

			err := template.RemoveVolume("foo")
			Expect(err).NotTo(HaveOccurred())
			Expect(template.Data).To(Equal(map[string]interface{}{
				"volumes": []map[string]interface{}{
					{
						"name": "spam",
						"secret": map[string]interface{}{
							"secretName": "eggs",
						},
					},
				},
			}))
		})

		It("should add volume mounts", func() {
			volume := v1.VolumeMount{
				Name:      "foo",
				MountPath: "/bindings/foo-binding",
			}

			container := pipeline.MetaContainer{
				Data:        map[string]interface{}{},
				VolumeMount: []string{"spec", "volumeMounts"},
			}

			err := container.AddVolumeMount(volume)
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"spec": map[string]interface{}{
					"volumeMounts": []map[string]interface{}{
						{
							"name":      volume.Name,
							"mountPath": volume.MountPath,
						},
					},
				},
			}))
		})

		It("should not alter non-service binding volume mounts when adding volume mounts", func() {
			volume := v1.VolumeMount{
				Name:      "foo",
				MountPath: "/bindings/foo-binding",
			}

			container := pipeline.MetaContainer{
				Data: map[string]interface{}{
					"spec": map[string]interface{}{
						"volumeMounts": []map[string]interface{}{
							{
								"name":      "bar",
								"mountPath": "/some/other/path",
							},
						},
					},
				},
				VolumeMount: []string{"spec", "volumeMounts"},
			}

			err := container.AddVolumeMount(volume)
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"spec": map[string]interface{}{
					"volumeMounts": []map[string]interface{}{
						{
							"name":      "bar",
							"mountPath": "/some/other/path",
						},
						{
							"name":      volume.Name,
							"mountPath": volume.MountPath,
						},
					},
				},
			}))
		})

		It("should remove volume mounts", func() {
			container := pipeline.MetaContainer{
				Data: map[string]interface{}{
					"volumeMounts": []map[string]interface{}{
						{
							"name":      "foo",
							"mountPath": "/bindings/foo-binding",
						},
					},
				},
				VolumeMount: []string{"volumeMounts"},
			}

			err := container.RemoveVolumeMount("foo")
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"volumeMounts": []map[string]interface{}{},
			}))
		})

		It("should only remove service binding volume mounts", func() {
			container := pipeline.MetaContainer{
				Data: map[string]interface{}{
					"volumeMounts": []map[string]interface{}{
						{
							"name":      "bar",
							"mountPath": "/some/other/path",
						},
						{
							"name":      "foo",
							"mountPath": "/bindings/foo-binding",
						},
					},
				},
				VolumeMount: []string{"volumeMounts"},
			}

			err := container.RemoveVolumeMount("foo")
			Expect(err).NotTo(HaveOccurred())
			Expect(container.Data).To(Equal(map[string]interface{}{
				"volumeMounts": []map[string]interface{}{
					{
						"name":      "bar",
						"mountPath": "/some/other/path",
					},
				},
			}))
		})
	})
})
