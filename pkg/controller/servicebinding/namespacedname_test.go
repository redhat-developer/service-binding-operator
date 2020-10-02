package servicebinding

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

// isNamespacedNameEmpty returns true if any of the fields from the given namespacedName is empty.
func isNamespacedNameEmpty(namespacedName types.NamespacedName) bool {
	return namespacedName.Namespace == "" || namespacedName.Name == ""
}

func TestNamespacednameIsSBRNamespacedNameEmpty(t *testing.T) {
	require.True(t, isNamespacedNameEmpty(types.NamespacedName{}))
	require.True(t, isNamespacedNameEmpty(types.NamespacedName{Namespace: "ns"}))
	require.True(t, isNamespacedNameEmpty(types.NamespacedName{Name: "name"}))
	require.False(t, isNamespacedNameEmpty(types.NamespacedName{Namespace: "ns", Name: "name"}))
}
