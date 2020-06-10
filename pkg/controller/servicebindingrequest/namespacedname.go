package servicebindingrequest

import "k8s.io/apimachinery/pkg/types"

// isNamespacedNameEmpty returns true if any of the fields from the given namespacedName is empty.
func isNamespacedNameEmpty(namespacedName types.NamespacedName) bool {
	return namespacedName.Namespace == "" || namespacedName.Name == ""
}
