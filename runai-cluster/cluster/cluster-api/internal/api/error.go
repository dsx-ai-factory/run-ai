package api

import (
	"errors"
	"net/http"
)

type NotFoundError struct {
	message string
}

func (e NotFoundError) Error() string {
	return "not found"
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		message: message,
	}
}

type NotReadyError struct {
	message string
}

func (e NotReadyError) Error() string {
	return e.message
}

func NewNotReadyError(message string) *NotReadyError {
	return &NotReadyError{
		message: message,
	}
}

type ProtocolError struct {
	message string
}

func NewProtocolError(message string) *ProtocolError {
	return &ProtocolError{
		message: message,
	}
}

func (e ProtocolError) Error() string {
	return e.message
}

func ToHttpError(err error) (string, int) {
	var notFoundError *NotFoundError

	var notReadyError *NotReadyError

	switch {
	case errors.As(err, &notFoundError):
		return "not found", http.StatusNotFound

	case errors.As(err, &notReadyError):
		return err.Error(), http.StatusPreconditionFailed

	default:
		return err.Error(), http.StatusServiceUnavailable
	}
}
