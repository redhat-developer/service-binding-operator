package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
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
	BackingServiceSelector BackingServiceSelector `json:"backingServiceSelector"`

	// BackingServiceSelectors is used to identify multiple backing services.
	BackingServiceSelectors []BackingServiceSelector `json:"backingServiceSelectors"`

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
	// Secret is the name of the intermediate secret
	Secret string `json:"secret,omitempty"`
	// ApplicationObjects contains all the application objects filtered by label
	ApplicationObjects []string `json:"applicationObjects,omitempty"`
}

// BackingServiceSelector defines the selector based on resource name, version, and resource kind
// +k8s:openapi-gen=true
type BackingServiceSelector struct {
	metav1.GroupVersionKind `json:",inline"`
	ResourceRef             string `json:"resourceRef"`
}

// ApplicationSelector defines the selector based on labels and GVR
// +k8s:openapi-gen=true
type ApplicationSelector struct {
	// +optional
	LabelSelector               *metav1.LabelSelector `json:"labelSelector,omitempty"`
	metav1.GroupVersionResource `json:",inline"`
	ResourceRef                 string `json:"resourceRef,omitempty"`
	Namespace                   string `json:"namespace,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingRequest is the Schema for the servicebindings API
// +k8s:openapi-gen=true
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
