package annotations

import (
	"errors"
	"fmt"
)

type invalidArgumentErr string

func (e invalidArgumentErr) Error() string {
	return fmt.Sprintf("invalid argument value for path %q", string(e))
}

var resourceNameFieldNotFoundErr = errors.New("secret name field not found")

type unknownBindingTypeErr string

func (e unknownBindingTypeErr) Error() string {
	return string(e) + " is not supported"
}

type errInvalidBindingValue string

func (e errInvalidBindingValue) Error() string {
	return fmt.Sprintf("invalid binding value %q", string(e))
}
