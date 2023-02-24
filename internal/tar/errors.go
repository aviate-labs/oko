package tar

import "fmt"

type TarError struct {
	Err error
}

func NewTarError(err error) *TarError {
	return &TarError{
		Err: err,
	}
}

func (e TarError) Error() string {
	return fmt.Sprintf("tar error: %s", e.Err)
}

type UnexpectedStatusCodeError struct {
	StatusCode int
}

func NewUnexpectedStatusCodeError(statusCode int) *UnexpectedStatusCodeError {
	return &UnexpectedStatusCodeError{
		StatusCode: statusCode,
	}
}

func (e UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
}
