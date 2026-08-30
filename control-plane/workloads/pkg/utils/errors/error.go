package errors

import (
	"strings"
)

func isError(err error, shouldContain string) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())

	return strings.Contains(errMsg, shouldContain)
}

// IsNotFound checks if the error is a 404 Not Found error
func IsNotFound(err error) bool {
	return isError(err, "not found")
}

// IsBadRequest checks if the error is a 400 Bad Request error
func IsBadRequest(err error) bool {
	return isError(err, "bad request")
}

// IsUnauthorized checks if the error is a 401 Unauthorized error
func IsUnauthorized(err error) bool {
	return isError(err, "unauthorized")
}

// IsForbidden checks if the error is a 403 Forbidden error
func IsForbidden(err error) bool {
	return isError(err, "forbidden")
}

// IsInternalServerError checks if the error is a 500 Internal Server Error
func IsInternalServerError(err error) bool {
	return isError(err, "internal server error")
}

// IsServiceUnavailable checks if the error is a 503 Service Unavailable error
func IsServiceUnavailable(err error) bool {
	return isError(err, "service unavailable")
}
