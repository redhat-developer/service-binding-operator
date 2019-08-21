package apis

import (
	"github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}
