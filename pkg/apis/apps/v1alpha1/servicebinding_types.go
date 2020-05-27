package v1alpha1

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// ServiceBindingSpec defines the desired state of ServiceBindingSpec
type ServiceBindingSpec struct {
	// MountPathPrefix is the prefix for volume mount
	// +optional
	MountPathPrefix string `json:"mountPathPrefix,omitempty"`

	// EnvVarPrefix is the prefix for environment variables
	// +optional
	EnvVarPrefix string `json:"envVarPrefix,omitempty"`

	// DataMapping is used to map CR/CSV/CRD attributes to Custom env variables
	// +optional
	DataMapping []corev1.EnvVar `json:"dataMapping,omitempty"`

	// BackingServiceSelectors is used to identify multiple backing services.
	// This would be made a required field after 'services'
	// is removed.
	// +optional
	Services []Service `json:"services"`

	// Application is used to identify the application connecting to the
	// backing service operator.
	// +optional
	Application Application `json:"application,omitempty"`

	// DetectBindingResources is flag used to bind all non-bindable variables from
	// different subresources owned by backing operator CR.
	// +optional
	DetectBindingResources bool `json:"detectBindingResources,omitempty"`
}

// ServiceBindingStatus defines the observed state of ServiceBinding
// +k8s:openapi-gen=true
type ServiceBindingStatus struct {
	// Conditions describes the state of the operator's reconciliation functionality.
	Conditions []conditionsv1.Condition `json:"conditions"`
	// Secret is the name of the intermediate secret
	Secret corev1.LocalObjectReference `json:"secretRef"`
	// ApplicationObjects contains all the application objects filtered by label
	// +optional
	Applications []BoundApplication `json:"applications,omitempty"`
}

// Service defines the selector based on resource name, version, and resource kind
type Service struct {
	metav1.GroupVersionKind     `json:",inline"`
	corev1.LocalObjectReference `json:",inline"`
	EnvVarPrefix                *string `json:"envVarPrefix,omitempty"`

	// +optional
	// This is en EXPERIMENTAL feature till
	// we support Subject Access Reviews.
	Namespace *string `json:"namespace,omitempty"`
}

// BoundApplication defines the application workloads to which the binding secret has
// injected. This is used in the status subresource.
type BoundApplication struct {
	metav1.GroupVersionKind     `json:",inline"`
	corev1.LocalObjectReference `json:",inline"`
}

// Application defines the selector based on labels and GVR
type Application struct {
	// +optional
	LabelSelector               *metav1.LabelSelector `json:"labelSelector,omitempty"`
	metav1.GroupVersionResource `json:",inline"`
	corev1.LocalObjectReference `json:",inline,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBinding expresses intent to bind an operator-backed service with
// an application workload.
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Service Binding"
// +kubebuilder:resource:path=servicebindings,shortName=sb;sbs
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec   `json:"spec,omitempty"`
	Status ServiceBindingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingList contains a list of ServiceBinding
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBinding{}, &ServiceBindingList{})
}