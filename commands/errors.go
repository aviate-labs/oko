package commands

import "fmt"

type OptionsError struct {
	Message string
}

func NewOptionsError(msg string) *OptionsError {
	return &OptionsError{
		Message: msg,
	}
}

func (e OptionsError) Error() string {
	return e.Message
}

type PathNotFoundError struct {
	Path string
}

func NewPathNotFoundError(path string) *PathNotFoundError {
	return &PathNotFoundError{
		Path: path,
	}
}

func (e PathNotFoundError) Error() string {
	return fmt.Sprintf("could not find path: %q", e.Path)
}
