package conditions

import conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"

type ConditionType conditionsv1.ConditionType

type ConditionStatus corev1.ConditionStatus

const (
	// BindingReady indicates that the binding succeeded
	BindingReady ConditionType = "Binding is ready"

	// BindingInProgress indicates that binding is in progress
	BindingInProgress ConditionType = "Binding is in progress"

	// BindingFailed indicates that the binding failed
	BindingFailed ConditionType = "Binding failed"

	ConditionTrue = ConditionStatus(corev1.ConditionTrue)
	ConditionFalse = ConditionStatus(corev1.ConditionFalse)
)