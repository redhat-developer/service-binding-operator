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

package v1beta1

import (
	"errors"

	"github.com/redhat-developer/service-binding-operator/apis"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var log = logf.Log.WithName("WebHook Spec ServiceBinding")

func (r *ServiceBinding) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-servicebinding-io-v1beta1-servicebinding,mutating=false,failurePolicy=fail,sideEffects=None,groups=servicebinding.io,resources=servicebindings,verbs=create;update,versions=v1beta1,name=vspecservicebinding.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ServiceBinding{}

func checkNameAndSelector(r *ServiceBinding) error {
	if r.Spec.Workload.Name != "" && r.Spec.Workload.Selector != nil && r.Spec.Workload.Selector.MatchLabels != nil {
		err := errors.New("name and selector MUST NOT be defined in the application reference")
		log.Error(err, "name and selector check failed")
		return err
	}
	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceBinding) ValidateCreate() error {
	return checkNameAndSelector(r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (sb *ServiceBinding) ValidateUpdate(old runtime.Object) error {
	oldSb, ok := old.(*ServiceBinding)
	if !ok {
		return errors.New("Old object is not service binding")
	}
	err := apis.CanUpdateBinding(sb, oldSb)
	if err != nil {
		log.Error(err, "Update failed")
		return err
	}
	return checkNameAndSelector(sb)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ServiceBinding) ValidateDelete() error {
	log.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
