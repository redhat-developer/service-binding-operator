package servicebindingrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestNamespacednameIsSBRNamespacedNameEmpty(t *testing.T) {
	assert.True(t, IsNamespacedNameEmpty(types.NamespacedName{}))
	assert.True(t, IsNamespacedNameEmpty(types.NamespacedName{Namespace: "ns"}))
	assert.True(t, IsNamespacedNameEmpty(types.NamespacedName{Name: "name"}))
	assert.False(t, IsNamespacedNameEmpty(types.NamespacedName{Namespace: "ns", Name: "name"}))
}
