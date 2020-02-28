// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"./pkg/apis/apps/v1alpha1.ApplicationSelector":         schema_pkg_apis_apps_v1alpha1_ApplicationSelector(ref),
		"./pkg/apis/apps/v1alpha1.BackingServiceSelector":      schema_pkg_apis_apps_v1alpha1_BackingServiceSelector(ref),
		"./pkg/apis/apps/v1alpha1.ServiceBindingRequest":       schema_pkg_apis_apps_v1alpha1_ServiceBindingRequest(ref),
		"./pkg/apis/apps/v1alpha1.ServiceBindingRequestSpec":   schema_pkg_apis_apps_v1alpha1_ServiceBindingRequestSpec(ref),
		"./pkg/apis/apps/v1alpha1.ServiceBindingRequestStatus": schema_pkg_apis_apps_v1alpha1_ServiceBindingRequestStatus(ref),
	}
}

func schema_pkg_apis_apps_v1alpha1_ApplicationSelector(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ApplicationSelector defines the selector based on labels and GVR",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"labelSelector": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.LabelSelector"),
						},
					},
					"group": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"version": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"resource": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"resourceRef": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"bindingPath": {
						SchemaProps: spec.SchemaProps{
							Description: "BindingPath refers to the path in the application workload's schema where the binding workload would be referenced.",
							Ref:         ref("./pkg/apis/apps/v1alpha1.BindingPath"),
						},
					},
				},
				Required: []string{"group", "version", "resource"},
			},
		},
		Dependencies: []string{
			"./pkg/apis/apps/v1alpha1.BindingPath", "k8s.io/apimachinery/pkg/apis/meta/v1.LabelSelector"},
	}
}

func schema_pkg_apis_apps_v1alpha1_BackingServiceSelector(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BackingServiceSelector defines the selector based on resource name, version, and resource kind",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"group": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"version": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"kind": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"resourceRef": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"group", "version", "kind", "resourceRef"},
			},
		},
	}
}

func schema_pkg_apis_apps_v1alpha1_ServiceBindingRequest(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Expresses intent to bind an operator-backed service with a Deployment",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.ServiceBindingRequestSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.ServiceBindingRequestStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/apps/v1alpha1.ServiceBindingRequestSpec", "./pkg/apis/apps/v1alpha1.ServiceBindingRequestStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_apps_v1alpha1_ServiceBindingRequestSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceBindingRequestSpec defines the desired state of ServiceBindingRequest",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"mountPathPrefix": {
						SchemaProps: spec.SchemaProps{
							Description: "MountPathPrefix is the prefix for volume mount",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"envVarPrefix": {
						SchemaProps: spec.SchemaProps{
							Description: "EnvVarPrefix is the prefix for environment variables",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"customEnvVar": {
						SchemaProps: spec.SchemaProps{
							Description: "Custom env variables",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.EnvVar"),
									},
								},
							},
						},
					},
					"backingServiceSelector": {
						SchemaProps: spec.SchemaProps{
							Description: "BackingServiceSelector is used to identify the backing service operator.",
							Ref:         ref("./pkg/apis/apps/v1alpha1.BackingServiceSelector"),
						},
					},
					"backingServiceSelectors": {
						SchemaProps: spec.SchemaProps{
							Description: "BackingServiceSelectors is used to identify multiple backing services.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("./pkg/apis/apps/v1alpha1.BackingServiceSelector"),
									},
								},
							},
						},
					},
					"applicationSelector": {
						SchemaProps: spec.SchemaProps{
							Description: "ApplicationSelector is used to identify the application connecting to the backing service operator.",
							Ref:         ref("./pkg/apis/apps/v1alpha1.ApplicationSelector"),
						},
					},
					"detectBindingResources": {
						SchemaProps: spec.SchemaProps{
							Description: "DetectBindingResources is flag used to bind all non-bindable variables from different subresources owned by backing operator CR.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
				},
				Required: []string{"applicationSelector"},
			},
		},
		Dependencies: []string{
			"./pkg/apis/apps/v1alpha1.ApplicationSelector", "./pkg/apis/apps/v1alpha1.BackingServiceSelector", "k8s.io/api/core/v1.EnvVar"},
	}
}

func schema_pkg_apis_apps_v1alpha1_ServiceBindingRequestStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceBindingRequestStatus defines the observed state of ServiceBindingRequest",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"bindingStatus": {
						SchemaProps: spec.SchemaProps{
							Description: "BindingStatus is the status of the service binding request.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"conditions": {
						SchemaProps: spec.SchemaProps{
							Description: "Conditions describes the state of the operator's reconciliation functionality.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/openshift/custom-resource-status/conditions/v1.Condition"),
									},
								},
							},
						},
					},
					"secret": {
						SchemaProps: spec.SchemaProps{
							Description: "Secret is the name of the intermediate secret",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"applicationObjects": {
						SchemaProps: spec.SchemaProps{
							Description: "ApplicationObjects contains all the application objects filtered by label",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/openshift/custom-resource-status/conditions/v1.Condition"},
	}
}
