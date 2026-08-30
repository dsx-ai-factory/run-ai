package client

type ErrorHandlerFunc func(err error)

func DefaultErrorHandler(err error) {
}
