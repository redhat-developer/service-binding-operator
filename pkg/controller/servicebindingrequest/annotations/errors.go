package annotations

import (
	"fmt"
)

type invalidArgumentErr string

func (e invalidArgumentErr) Error() string {
	return fmt.Sprintf("invalid argument value for path %q", string(e))
}

type errResourceNameFieldNotFound string

func (e errResourceNameFieldNotFound) Error() string {
	return fmt.Sprintf("secret name field %q not found", string(e))
}

type unknownBindingTypeErr string

func (e unknownBindingTypeErr) Error() string {
	return string(e) + " is not supported"
}

type errInvalidBindingValue string

func (e errInvalidBindingValue) Error() string {
	return fmt.Sprintf("invalid binding value %q", string(e))
}
