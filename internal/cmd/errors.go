package cmd

type CommandNotFoundError struct{}

func NewCommandNotFoundError() *CommandNotFoundError {
	return &CommandNotFoundError{}
}

func (e CommandNotFoundError) Error() string {
	return "command not found"
}

type InvalidArgumentsError struct {
	Message string
}

func NewInvalidArgumentsError(msg string) *InvalidArgumentsError {
	return &InvalidArgumentsError{
		Message: msg,
	}
}

func (e InvalidArgumentsError) Error() string {
	return e.Message
}
