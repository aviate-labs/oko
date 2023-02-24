package internal

import "fmt"

type InternalError struct {
	Err error
}

func Error(err error) *InternalError {
	return &InternalError{
		Err: err,
	}
}

func (e InternalError) Error() string {
	return fmt.Sprintf(
		"internal error: %s",
		e.Err.Error(),
	)
}
