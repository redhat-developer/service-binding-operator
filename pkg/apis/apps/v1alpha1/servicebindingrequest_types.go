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

	// BackingServiceSelector is used to identify the backing service operator.
	//
	// Refer: https://12factor.net/backing-services
	// A backing service is any service the app consumes over the network as
	// part of its normal operation. Examples include datastores (such as
	// MySQL or CouchDB), messaging/queueing systems (such as RabbitMQ or
	// Beanstalkd), SMTP services for outbound email (such as Postfix), and
	// caching systems (such as Memcached).
	//
	// Example 1:
	//	backingServiceSelector:
	//		resourceName: database.example.org
	//      resourceRef: mysql-database
	// Example 2:
	//	backingServiceSelector:
	//		resourceName: database.example.org
	//		resourceVersion: v1alpha1
	//      resourceRef: mysql-database
	BackingServiceSelector BackingServiceSelector `json:"backingServiceSelector"`

	// ApplicationSelector is used to identify the application connecting to the
	// backing service operator.
	// Example 1:
	//	applicationSelector:
	//		matchLabels:
	//			connects-to: postgres
	//			environment: stage
	//		resourceKind: Deployment
	// Example 2:
	//	applicationSelector:
	//		resourceKind: Deployment
	//		resourceName: my-app
	ApplicationSelector ApplicationSelector `json:"applicationSelector"`
}

// BackingServiceSelector defines the selector based on resource name, version, and resource kind
// +k8s:openapi-gen=true
type BackingServiceSelector struct {
	ResourceName    string `json:"resourceName"`
	ResourceVersion string `json:"resourceVersion"`
	ResourceRef     string `json:"resourceRef"`
}

// ApplicationSelector defines the selector based on labels and resource kind
// +k8s:openapi-gen=true
type ApplicationSelector struct {
	MatchLabels  map[string]string `json:"matchLabels"`
	ResourceKind string            `json:"resourceKind"`
}

// ServiceBindingRequestStatus defines the observed state of ServiceBindingRequest
// +k8s:openapi-gen=true
type ServiceBindingRequestStatus struct{}

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
