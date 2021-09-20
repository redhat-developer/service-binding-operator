/*
Copyright 2021.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"errors"
	"github.com/go-logr/logr"
	"github.com/redhat-developer/service-binding-operator/apis"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var log = logf.Log.WithName("WebHook ServiceBinding")

func (r *ServiceBinding) SetupWebhookWithManager(mgr ctrl.Manager, serviceAccountName string) error {
	mgr.GetWebhookServer().Register("/mutate-servicebinding", &webhook.Admission{
		Handler: &admisionHandler{serviceAccountName: serviceAccountName},
	})
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-servicebinding,mutating=true,failurePolicy=fail,sideEffects=None,groups=binding.operators.coreos.com,resources=servicebindings,verbs=create;update,versions=v1alpha1,name=mservicebinding.kb.io,admissionReviewVersions={v1beta1}
// +kubebuilder:webhook:path=/mutate-servicebinding,mutating=true,failurePolicy=fail,sideEffects=None,groups=servicebinding.io,resources=servicebindings,verbs=create;update,versions=v1alpha3,name=mspec-servicebinding.kb.io,admissionReviewVersions={v1beta1}

type admisionHandler struct {
	decoder            *admission.Decoder
	log                logr.Logger
	serviceAccountName string
}

var _ webhook.AdmissionHandler = &admisionHandler{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (ah *admisionHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.UserInfo.Username == ah.serviceAccountName {
		return admission.Allowed("ok")
	}
	sb := &unstructured.Unstructured{}
	err := ah.decoder.Decode(req, sb)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if req.Operation == v1beta1.Create || req.Operation == v1beta1.Update {
		apis.SetRequester(sb, req.UserInfo)
	} else {
		return admission.Allowed("ok")
	}
	marshaledSB, err := sb.MarshalJSON()
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledSB)
}

func (ah *admisionHandler) InjectDecoder(decoder *admission.Decoder) error {
	ah.decoder = decoder
	return nil
}

func (ah *admisionHandler) InjectLogger(l logr.Logger) error {
	ah.log = l
	return nil
}

// +kubebuilder:webhook:path=/validate-binding-operators-coreos-com-v1alpha1-servicebinding,mutating=false,failurePolicy=fail,sideEffects=None,groups=binding.operators.coreos.com,resources=servicebindings,verbs=update,versions=v1alpha1,name=vservicebinding.kb.io,admissionReviewVersions={v1beta1}

var _ webhook.Validator = &ServiceBinding{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceBinding) ValidateCreate() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceBinding) ValidateUpdate(old runtime.Object) error {
	oldSb, ok := old.(*ServiceBinding)
	if !ok {
		return errors.New("Old object is not service binding")
	}
	err := apis.CanUpdateBinding(r, oldSb)
	if err != nil {
		log.Error(err, "Update failed")
	}
	return err
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceBinding) ValidateDelete() error {
	return nil
}
