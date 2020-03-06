package v1alpha1

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// ServiceBindingRequestSpec defines the desired state of ServiceBindingRequest
// +k8s:openapi-gen=true
type ServiceBindingRequestSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags:
	// 	https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// MountPathPrefix is the prefix for volume mount
	// +optional
	MountPathPrefix string `json:"mountPathPrefix,omitempty"`

	// EnvVarPrefix is the prefix for environment variables
	// +optional
	EnvVarPrefix string `json:"envVarPrefix,omitempty"`

	// Custom env variables
	// +optional
	CustomEnvVar []corev1.EnvVar `json:"customEnvVar,omitempty"`

	// BackingServiceSelector is used to identify the backing service operator.
	// +optional
	BackingServiceSelector BackingServiceSelector `json:"backingServiceSelector,omitempty"`

	// BackingServiceSelectors is used to identify multiple backing services.
	BackingServiceSelectors []BackingServiceSelector `json:"backingServiceSelectors,omitempty"`

	// ApplicationSelector is used to identify the application connecting to the
	// backing service operator.
	ApplicationSelector ApplicationSelector `json:"applicationSelector"`

	// DetectBindingResources is flag used to bind all non-bindable variables from
	// different subresources owned by backing operator CR.
	// +optional
	DetectBindingResources bool `json:"detectBindingResources"`
}

// ServiceBindingRequestStatus defines the observed state of ServiceBindingRequest
// +k8s:openapi-gen=true
type ServiceBindingRequestStatus struct {
	// BindingStatus is the status of the service binding request.
	BindingStatus string `json:"bindingStatus,omitempty"`
	// Conditions describes the state of the operator's reconciliation functionality.
	Conditions []conditionsv1.Condition `json:"conditions,omitempty"`
	// Secret is the name of the intermediate secret
	Secret string `json:"secret,omitempty"`
	// ApplicationObjects contains all the application objects filtered by label
	ApplicationObjects []BoundApplication `json:"applications,omitempty"`
}

// BackingServiceSelector defines the selector based on resource name, version, and resource kind
// +k8s:openapi-gen=true
type BackingServiceSelector struct {
	metav1.GroupVersionKind `json:",inline"`
	ResourceRef             string `json:"resourceRef"`
}

// BoundApplication defines the application workloads to which the binding secret has
// injected.
type BoundApplication struct {
	metav1.GroupVersionKind `json:",inline"`
	v1.LocalObjectReference `json:",inline"`
}

// ApplicationSelector defines the selector based on labels and GVR
// +k8s:openapi-gen=true
type ApplicationSelector struct {
	// +optional
	LabelSelector               *metav1.LabelSelector `json:"labelSelector,omitempty"`
	metav1.GroupVersionResource `json:",inline"`
	ResourceRef                 string `json:"resourceRef,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingRequest expresses intent to bind an operator-backed service with
// an application workload.
// +k8s:openapi-gen=true
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Service Binding Request"
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=servicebindingrequests,shortName=sbr;sbrs
type ServiceBindingRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingRequestSpec   `json:"spec,omitempty"`
	Status ServiceBindingRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingRequestList contains a list of ServiceBindingRequest
type ServiceBindingRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ServiceBindingRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBindingRequest{}, &ServiceBindingRequestList{})
}
