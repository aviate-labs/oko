package schema

import (
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type ValidationError struct {
	Errors []gojsonschema.ResultError
}

func NewValidationError(errors []gojsonschema.ResultError) *ValidationError {
	return &ValidationError{
		Errors: errors,
	}
}

func (e ValidationError) Error() string {
	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Description())
	}
	return fmt.Sprintf("validation error:\n%s", strings.Join(messages, ", "))
}
