package conditions

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
)

const (
	// BindingReady indicates that the binding succeeded
	BindingReady conditionsv1.ConditionType = "Ready"
)
