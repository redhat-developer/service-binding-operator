package servicebinding

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

func TestNamespacednameIsSBRNamespacedNameEmpty(t *testing.T) {
	require.True(t, isNamespacedNameEmpty(types.NamespacedName{}))
	require.True(t, isNamespacedNameEmpty(types.NamespacedName{Namespace: "ns"}))
	require.True(t, isNamespacedNameEmpty(types.NamespacedName{Name: "name"}))
	require.False(t, isNamespacedNameEmpty(types.NamespacedName{Namespace: "ns", Name: "name"}))
}
