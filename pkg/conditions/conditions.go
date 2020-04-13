package conditions

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
)

const (
	// CollectionReady indicates readiness for collection and persistance of intermediate manifests
	CollectionReady conditionsv1.ConditionType = "CollectionReady"
	// InjectionReady indicates readiness to change application manifests to use those intermediate manifests
	InjectionReady conditionsv1.ConditionType = "InjectionReady"
	// BindingReady indicates readiness to bind
	BindingReady conditionsv1.ConditionType = "BindingReady"
)
