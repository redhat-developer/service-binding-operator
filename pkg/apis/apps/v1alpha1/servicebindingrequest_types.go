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
	//		resourceKind: databases.example.org
	//      resourceRef: mysql-database
	// Example 2:
	//	backingServiceSelector:
	//		resourceKind: databases.example.org
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
	//		matchLabels:
	//			connects-to: postgres
	//			environment: stage
	ApplicationSelector ApplicationSelector `json:"applicationSelector"`
}

// BackingServiceSelector defines the selector based on resource name, version, and resource kind
// +k8s:openapi-gen=true
type BackingServiceSelector struct {
	Group       *string `json:"group,omitempty"`
	Version     string  `json:"version"`
	Kind        string  `json:"kind"`
	ResourceRef string  `json:"resourceRef"`
}

// ApplicationSelector defines the selector based on labels and resource kind
// +k8s:openapi-gen=true
type ApplicationSelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
	Group       *string           `json:"group,omitempty"`
	Version     string            `json:"version"`
	Kind        string            `json:"kind"`
}

type BindingStatus string

const (
	BindingSuccess    BindingStatus = "success"
	BindingInProgress BindingStatus = "inProgress"
	BindingFail       BindingStatus = "fail"
)

// ServiceBindingRequestStatus defines the observed state of ServiceBindingRequest
// +k8s:openapi-gen=true
type ServiceBindingRequestStatus struct {
	// BindingStatus is the status of the service binding request. Possible values are Success, Failure, InProgress.
	BindingStatus BindingStatus `json:"bindingStatus,omitempty"`
	// Secret is the name of the intermediate secret
	Secret string `json:"secret,omitempty"`
	// ApplicationObjects contains all the application objects filtered by label
	ApplicationObjects []string `json:"applicationObjects,omitempty"`
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
