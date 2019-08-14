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
				Description: "ApplicationSelector defines the selector based on labels and resource kind",
				Properties: map[string]spec.Schema{
					"matchLabels": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
					"resourceKind": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"matchLabels", "resourceKind"},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_apps_v1alpha1_BackingServiceSelector(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BackingServiceSelector defines the selector based on resource name, version, and resource kind",
				Properties: map[string]spec.Schema{
					"resourceKind": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"resourceVersion": {
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
				Required: []string{"resourceKind", "resourceVersion", "resourceRef"},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_apps_v1alpha1_ServiceBindingRequest(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceBindingRequest is the Schema for the servicebindings API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
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
					"backingServiceSelector": {
						SchemaProps: spec.SchemaProps{
							Description: "BackingServiceSelector is used to identify the backing service operator.\n\nRefer: https://12factor.net/backing-services A backing service is any service the app consumes over the network as part of its normal operation. Examples include datastores (such as MySQL or CouchDB), messaging/queueing systems (such as RabbitMQ or Beanstalkd), SMTP services for outbound email (such as Postfix), and caching systems (such as Memcached).\n\nExample 1:\n\tbackingServiceSelector:\n\t\tresourceKind: databases.example.org\n     resourceRef: mysql-database\nExample 2:\n\tbackingServiceSelector:\n\t\tresourceKind: databases.example.org\n\t\tresourceVersion: v1alpha1\n     resourceRef: mysql-database",
							Ref:         ref("./pkg/apis/apps/v1alpha1.BackingServiceSelector"),
						},
					},
					"applicationSelector": {
						SchemaProps: spec.SchemaProps{
							Description: "ApplicationSelector is used to identify the application connecting to the backing service operator. Example 1:\n\tapplicationSelector:\n\t\tmatchLabels:\n\t\t\tconnects-to: postgres\n\t\t\tenvironment: stage\n\t\tresourceKind: Deployment\nExample 2:\n\tapplicationSelector:\n\t\tmatchLabels:\n\t\t\tconnects-to: postgres\n\t\t\tenvironment: stage",
							Ref:         ref("./pkg/apis/apps/v1alpha1.ApplicationSelector"),
						},
					},
				},
				Required: []string{"mountPathPrefix", "backingServiceSelector", "applicationSelector"},
			},
		},
		Dependencies: []string{
			"./pkg/apis/apps/v1alpha1.ApplicationSelector", "./pkg/apis/apps/v1alpha1.BackingServiceSelector"},
	}
}

func schema_pkg_apis_apps_v1alpha1_ServiceBindingRequestStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceBindingRequestStatus defines the observed state of ServiceBindingRequest",
				Properties: map[string]spec.Schema{
					"bindingStatus": {
						SchemaProps: spec.SchemaProps{
							Description: "BindingStatus is the status of the service binding request. Possible values are Success, Failure, InProgress.",
							Type:        []string{"string"},
							Format:      "",
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
				Required: []string{"bindingStatus", "secret", "applicationObjects"},
			},
		},
		Dependencies: []string{},
	}
}
