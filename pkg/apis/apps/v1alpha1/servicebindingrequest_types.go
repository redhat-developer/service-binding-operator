package v1alpha1

import (
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
	MountPathPrefix string `json:"mountPathPrefix,omitempty"`

	// EnvVarPrefix is the prefix for environment variables
	// +optional
	EnvVarPrefix string `json:"envVarPrefix,omitempty"`

	// BackingServiceSelector is used to identify the backing service operator.
	BackingServiceSelector BackingServiceSelector `json:"backingServiceSelector"`

	// ApplicationSelector is used to identify the application connecting to the
	// backing service operator.
	ApplicationSelector ApplicationSelector `json:"applicationSelector"`

	// EnvVar defines a list of overrides for the environment variables
	EnvVar []EnvMap `json:"envVar"`
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
	Group       string `json:"group,omitempty"`
	Version     string `json:"version"`
	Kind        string `json:"kind"`
	ResourceRef string `json:"resourceRef"`
}

// EnvMap is a set of Name and Value of an environment variable
// +k8s:openapi-gen=true
type EnvMap struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type BindingStatus string

const (
	BindingSuccess    BindingStatus = "success"
	BindingInProgress BindingStatus = "inProgress"
	BindingFail       BindingStatus = "fail"
)

// ServiceBindingRequestStatus defines the observed state of ServiceBindingRequest
// ApplicationSelector defines the selector based on labels and GVR
// +k8s:openapi-gen=true
type ApplicationSelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
	Group       string            `json:"group,omitempty"`
	Version     string            `json:"version"`
	Resource    string            `json:"resource"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBindingRequest is the Schema for the servicebindings API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
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
