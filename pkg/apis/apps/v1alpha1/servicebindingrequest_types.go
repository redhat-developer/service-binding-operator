package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceBindingRequestSpec defines the desired state of ServiceBindingRequest
// +k8s:openapi-gen=true
type ServiceBindingRequestSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	CSVName          string `json:"csvName"`
	CSVNamespace     string `json:"csvNamespace"`
	ApplicationLabel string `json:"ApplicationLabel"`
}

// ServiceBindingRequestStatus defines the observed state of ServiceBindingRequest
// +k8s:openapi-gen=true
type ServiceBindingRequestStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
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
	Items           []ServiceBindingRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBindingRequest{}, &ServiceBindingRequestList{})
}
