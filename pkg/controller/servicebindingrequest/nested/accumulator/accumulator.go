package accumulator

import (
	"errors"

	"github.com/imdario/mergo"
)

const valuesKey = "values"

// accumulator is a value accumulator.
type accumulator map[string]interface{}

// unsupportedTypeErr is returned when an unsupported type is encountered.
var unsupportedTypeErr = errors.New("unsupported type")

// Accumulate accumulates the `val` value. An error is returned in the case
// `val` contains an unsupported type.
func (a accumulator) Accumulate(val interface{}) error {
	b := NewAccumulator()
	switch v := val.(type) {
	case map[string]interface{}:
		b[valuesKey] = []map[string]interface{}{v}
	case string:
		b[valuesKey] = []string{v}
	case int:
		b[valuesKey] = []int{v}
	case []map[string]interface{}, []string, []int:
		b[valuesKey] = v
	default:
		return unsupportedTypeErr
	}
	return mergo.Merge(&a, b, mergo.WithAppendSlice, mergo.WithOverride, mergo.WithTypeCheck)
}

// Value returns the accumulated values.
func (a accumulator) Value() interface{} {
	return a[valuesKey]
}

// NewAccumulator returns a new value accumulator
func NewAccumulator() accumulator {
	return accumulator{}
}
