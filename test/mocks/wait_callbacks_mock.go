package mocks

import (
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/resourcepoll"
)

type WaitCallbackMock struct {
	resourcepoll.WaitCallbacks
}

func (w WaitCallbackMock) OnFind(sbr *v1alpha1.ServiceBindingRequest) error {
	return nil
}

func (w WaitCallbackMock) OnWait(sbr *v1alpha1.ServiceBindingRequest) error {
	return nil
}


