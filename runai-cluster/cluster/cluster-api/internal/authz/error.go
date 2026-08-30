package authz

import "errors"

type UnauthorizedError struct {
	message string
}

func (e *UnauthorizedError) Error() string {
	return e.message
}

func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{
		message: message,
	}
}

func IsUnauthorizedError(err error) bool {
	var unauthorizedError *UnauthorizedError

	ok := errors.As(err, &unauthorizedError)

	return ok
}
