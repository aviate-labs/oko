package schema

import (
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type SchemaError struct {
	Err error
}

func NewSchemaError(err error) *SchemaError {
	return &SchemaError{
		Err: err,
	}
}

func (e SchemaError) Error() string {
	return fmt.Sprintf(
		"schema error: %s",
		e.Err.Error(),
	)
}

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
	return fmt.Sprintf(
		"validation error: %s",
		strings.Join(messages, ", "),
	)
}
