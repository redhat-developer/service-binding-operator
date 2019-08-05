package servicebindingrequest

import "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"

func (r *Reconciler) SetBindingSuccessStatus(sbr *v1alpha1.ServiceBindingRequest) { 
	sbr.Status.BindingStatus = v1alpha1.BindingSuccess
}

func (r *Reconciler) SetBindingInProgressStatus(sbr *v1alpha1.ServiceBindingRequest) { 
	sbr.Status.BindingStatus = v1alpha1.BindingInProgress
}

func (r *Reconciler) SetBindingFailStatus(sbr *v1alpha1.ServiceBindingRequest) { 
	sbr.Status.BindingStatus = v1alpha1.BindingFail
}


