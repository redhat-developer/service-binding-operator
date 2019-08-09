package servicebindingrequest

import (
	"github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

func (r *Reconciler) setBindingSuccessStatus(sbr *v1alpha1.ServiceBindingRequest) {
	sbr.Status.BindingStatus = v1alpha1.BindingSuccess
}

func (r *Reconciler) setBindingInProgressStatus(sbr *v1alpha1.ServiceBindingRequest) {
	sbr.Status.BindingStatus = v1alpha1.BindingInProgress
}

func (r *Reconciler) setBindingFailStatus(sbr *v1alpha1.ServiceBindingRequest) {
	sbr.Status.BindingStatus = v1alpha1.BindingFail
}

func (r *Reconciler) setSecretStatus(sbr *v1alpha1.ServiceBindingRequest) {
	sbr.Status.Secret = sbr.GetName()
}

func (r *Reconciler) setLabelObjectsStatus(sbr *v1alpha1.ServiceBindingRequest, object string) {
	sbr.Status.LabelObjects = append(sbr.Status.LabelObjects, object)
}
