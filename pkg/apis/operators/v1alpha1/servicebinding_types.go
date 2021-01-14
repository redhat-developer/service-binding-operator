package v1alpha1

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// BindingReady indicates that the overall sbr succeeded
	BindingReady conditionsv1.ConditionType = "Ready"
	// CollectionReady indicates readiness for collection and persistance of intermediate manifests
	CollectionReady conditionsv1.ConditionType = "CollectionReady"
	// InjectionReady indicates readiness to change application manifests to use those intermediate manifests
	// If status is true, it indicates that the binding succeeded
	InjectionReady conditionsv1.ConditionType = "InjectionReady"
	// EmptyServiceSelectorsReason is used when the ServiceBinding has empty
	// services.
	EmptyServiceSelectorsReason = "EmptyServiceSelectors"
	// EmptyApplicationReason is used when the ServiceBinding has empty
	// application.
	EmptyApplicationReason = "EmptyApplication"
	// ApplicationNotFoundReason is used when the application is not found.
	ApplicationNotFoundReason = "ApplicationNotFound"
	// ServiceNotFoundReason is used when the service is not found.
	ServiceNotFoundReason = "ServiceNotFound"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// ServiceBindingSpec defines the desired state of ServiceBinding
type ServiceBindingSpec struct {
	// MountPath is the path inside app container where bindings will be mounted
	// If `SERVICE_BINDING_ROOT` env var is present, mountPath is ignored.
	// If `SERVICE_BINDING_ROOT` is absent and mountPath is present, set `SERVICE_BINDING_ROOT` as mountPath value
	// If `SERVICE_BINDING_ROOT` is absent but mounthPath is absent, set   SERVICE_BINDING_ROOT as `/bindings`
	// When mountPath is used, the file will be mounted directly under that directory
	// Otherwise it will be under `SERVICE_BINDING_ROOT`/<SERVICE-BINDING-NAME>
	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// NamePrefix is the prefix for environment variables or file name
	// +optional
	NamePrefix string `json:"namePrefix,omitempty"`

	// Custom mappings
	// +optional
	Mappings []Mapping `json:"mappings,omitempty"`

	// Services is used to identify multiple backing services.
	// +kubebuilder:validation:MinItems:=1
	Services []Service `json:"services"`

	// Application is used to identify the application connecting to the
	// backing service operator.
	// +optional
	Application *Application `json:"application,omitempty"`

	// DetectBindingResources is flag used to bind all non-bindable variables from
	// different subresources owned by backing operator CR.
	// +optional
	DetectBindingResources *bool `json:"detectBindingResources,omitempty"`

	// BindAsFiles makes available the binding values as files in the application's container
	// See MountPath attribute description for more details.
	// +optional
	BindAsFiles bool `json:"bindAsFiles,omitempty"`
}

// ServiceBindingMapping defines a new binding from set of existing bindings
type Mapping struct {
	// Name is the name of new binding
	Name string `json:"name"`
	// Value is a template which will be rendered and ibjected into the application
	Value string `json:"value"`
}

// ServiceBindingStatus defines the observed state of ServiceBinding
// +k8s:openapi-gen=true
type ServiceBindingStatus struct {
	// Conditions describes the state of the operator's reconciliation functionality.
	// +listType=set
	Conditions []conditionsv1.Condition `json:"conditions"`
	// Secret is the name of the intermediate secret
	Secret string `json:"secret"`
	// Applications contain all the applications filtered by name or label
	// +optional
	// +listType=set
	Applications []BoundApplication `json:"applications,omitempty"`
}

// Service defines the selector based on resource name, version, and resource kind
type Service struct {
	metav1.GroupVersionKind     `json:",inline"`
	corev1.LocalObjectReference `json:",inline"`

	// +optional
	Namespace  *string `json:"namespace,omitempty"`
	NamePrefix *string `json:"namePrefix,omitempty"`
	Id         *string `json:"id,omitempty"`
}

// BoundApplication defines the application workloads to which the binding secret has
// injected.
type BoundApplication struct {
	metav1.GroupVersionKind     `json:",inline"`
	corev1.LocalObjectReference `json:",inline"`
}

// Application defines the selector based on labels and GVR
type Application struct {
	corev1.LocalObjectReference `json:",inline"`
	// +optional
	LabelSelector               *metav1.LabelSelector `json:"labelSelector,omitempty"`
	metav1.GroupVersionResource `json:",inline"`

	// BindingPath refers to the paths in the application workload's schema
	// where the binding workload would be referenced.
	// If BindingPath is not specified the default path locations is going to
	// be used.  The default location for ContainersPath is
	// going to be: "spec.template.spec.containers" and if SecretPath
	// is not specified, the name of the secret object is not going
	// to be specified.
	// +optional
	BindingPath *BindingPath `json:"bindingPath,omitempty"`
}

// BindingPath defines the path to the field where the binding would be
// embedded in the workload
type BindingPath struct {
	// ContainersPath defines the path to the corev1.Containers reference
	// If BindingPath is not specified, the default location is
	// going to be: "spec.template.spec.containers"
	// +optional
	ContainersPath string `json:"containersPath"`

	// SecretPath defines the path to a string field where
	// the name of the secret object is going to be assigned.
	// Note: The name of the secret object is same as that of the name of SBR CR (metadata.name)
	// +optional
	SecretPath string `json:"secretPath"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceBinding expresses intent to bind an operator-backed service with
// an application workload.
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Service Binding"
// +kubebuilder:resource:path=servicebindings,shortName=sbr;sbrs
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ServiceBindingSpec   `json:"spec"`
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

func (sbr ServiceBinding) AsOwnerReference() metav1.OwnerReference {
	var ownerRefController bool = true
	return metav1.OwnerReference{
		Name:       sbr.Name,
		UID:        sbr.UID,
		Kind:       sbr.Kind,
		APIVersion: sbr.APIVersion,
		Controller: &ownerRefController,
	}
}
