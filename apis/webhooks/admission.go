package webhooks

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/redhat-developer/service-binding-operator/apis"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func SetupWithManager(mgr ctrl.Manager, serviceAccountName string) {
	mgr.GetWebhookServer().Register("/mutate-servicebinding", &webhook.Admission{
		Handler: &admissionHandler{serviceAccountName: serviceAccountName},
	})
}

// +kubebuilder:webhook:path=/mutate-servicebinding,mutating=true,failurePolicy=fail,sideEffects=None,groups=binding.operators.coreos.com,resources=servicebindings,verbs=create;update,versions=v1alpha1,name=mservicebinding.kb.io,admissionReviewVersions={v1beta1}
// +kubebuilder:webhook:path=/mutate-servicebinding,mutating=true,failurePolicy=fail,sideEffects=None,groups=servicebinding.io,resources=servicebindings,verbs=create;update,versions=v1alpha3,name=mspec-servicebinding.kb.io,admissionReviewVersions={v1beta1}

type admissionHandler struct {
	decoder            *admission.Decoder
	log                logr.Logger
	serviceAccountName string
}

var _ webhook.AdmissionHandler = &admissionHandler{}

func (ah *admissionHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
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

func (ah *admissionHandler) InjectDecoder(decoder *admission.Decoder) error {
	ah.decoder = decoder
	return nil
}

func (ah *admissionHandler) InjectLogger(l logr.Logger) error {
	ah.log = l
	return nil
}
