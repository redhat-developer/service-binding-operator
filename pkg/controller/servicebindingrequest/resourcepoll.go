package servicebindingrequest

import (
	"context"
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

func (r *Reconciler) OnWait(instance *v1alpha1.ServiceBindingRequest) error {
	r.setWaitingStatus(instance)
	err := r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) OnFind(instance *v1alpha1.ServiceBindingRequest) error {
	r.setBindingInProgressStatus(instance)
	err := r.client.Status().Update(context.TODO(), instance)
	if err != nil {
		return err
	}
	return nil
}
