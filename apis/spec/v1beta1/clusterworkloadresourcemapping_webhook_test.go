package v1beta1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Validation Webhook", func() {
	invalidEntries := []TableEntry{
		Entry("Duplicate versions",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
						},
						{
							Version: "v1",
						},
					},
				},
			},
			field.ErrorList{
				field.Duplicate(field.NewPath("spec", "versions[1]"), "v1"),
			}.ToAggregate(),
		),
		Entry("Invalid restricted jsonpath - annotations",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version:     "v1",
							Annotations: ".spec.template.spec.annotations[*]",
						},
					},
				},
			},
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "versions[0]", "annotations"), ".spec.template.spec.annotations[*]", "Invalid fixed JSONPath"),
			}.ToAggregate(),
		),
		Entry("Invalid restricted jsonpath - volumes",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
							Volumes: ".spec.template.spec.volumes[*]",
						},
					},
				},
			},
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "versions[0]", "volumes"), ".spec.template.spec.volumes[*]", "Invalid fixed JSONPath"),
			}.ToAggregate(),
		),
		Entry("Invalid restricted jsonpath - container name",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
							Containers: []ClusterWorkloadResourceMappingContainer{
								{
									Path: ".spec.template.spec.containers[*]",
									Name: ".name[*]",
								},
							},
						},
					},
				},
			},
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "versions[0]", "containers[0]", "name"), ".name[*]", "Invalid fixed JSONPath"),
			}.ToAggregate(),
		),
		Entry("Invalid restricted jsonpath - container env",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
							Containers: []ClusterWorkloadResourceMappingContainer{
								{
									Path: ".spec.template.spec.containers[*]",
									Env:  ".env[*]",
								},
							},
						},
					},
				},
			},
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "versions[0]", "containers[0]", "env"), ".env[*]", "Invalid fixed JSONPath"),
			}.ToAggregate(),
		),
		Entry("Invalid restricted jsonpath - container volumeMounts",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
							Containers: []ClusterWorkloadResourceMappingContainer{
								{
									Path:         ".spec.template.spec.containers[*]",
									VolumeMounts: ".volumeMounts[*]",
								},
							},
						},
					},
				},
			},
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "versions[0]", "containers[0]", "volumeMounts"), ".volumeMounts[*]", "Invalid fixed JSONPath"),
			}.ToAggregate(),
		),
		Entry("Invalid restricted jsonpath - failed to parse",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
							Containers: []ClusterWorkloadResourceMappingContainer{
								{
									Path:         ".spec.template.spec.containers[*]",
									VolumeMounts: ".volumeMounts[*",
								},
							},
						},
					},
				},
			},
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "versions[0]", "containers[0]", "volumeMounts"), ".volumeMounts[*", "Unable to parse fixed JSONPath"),
			}.ToAggregate(),
		),
		Entry("Invalid jsonpath - path",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
							Containers: []ClusterWorkloadResourceMappingContainer{
								{
									Path: ".spec.template.spec.containers[*",
								},
							},
						},
					},
				},
			},
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "versions[0]", "containers[0]", "path"), ".spec.template.spec.containers[*", "Invalid JSONPath"),
			}.ToAggregate(),
		),
		Entry("Required version field",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Annotations: "",
						},
					},
				},
			},
			field.ErrorList{
				field.Required(field.NewPath("spec", "versions[0]", "version"), "field \"version\" required"),
			}.ToAggregate(),
		),
		Entry("Required version field",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: " \t\n",
						},
					},
				},
			},
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "versions[0]", "version"), " \t\n", "Whitespace-only version field forbidden"),
			}.ToAggregate(),
		),
		Entry("Required path field",
			ClusterWorkloadResourceMapping{
				Spec: ClusterWorkloadResourceMappingSpec{
					Versions: []ClusterWorkloadResourceMappingTemplate{
						{
							Version: "v1",
							Containers: []ClusterWorkloadResourceMappingContainer{
								{
									Name: "",
								},
							},
						},
					},
				},
			},
			field.ErrorList{
				field.Required(field.NewPath("spec", "versions[0]", "containers[0]", "path"), "field \"path\" required"),
			}.ToAggregate(),
		),
	}
	DescribeTable("Reporting errors on invalid mappings",
		func(mapping ClusterWorkloadResourceMapping, expected error) {
			Expect(mapping.validate()).To(Equal(expected))
		},
		invalidEntries...,
	)
	It("should accept valid resources", func() {
		mapping := ClusterWorkloadResourceMapping{
			Spec: ClusterWorkloadResourceMappingSpec{
				Versions: []ClusterWorkloadResourceMappingTemplate{
					DefaultTemplate,
				},
			},
		}
		Expect(mapping.validate()).To(BeNil())
	})
})
