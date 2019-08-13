package resourcepoll

import "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"

type WaitCallbacks interface {
	OnWait(request *v1alpha1.ServiceBindingRequest) error
	OnFind(request *v1alpha1.ServiceBindingRequest) error
}

