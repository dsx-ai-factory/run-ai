package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/stretchr/testify/assert"
)

func TestHandleKeyCloakError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedError error
	}{
		{
			name:          "Not Found Error",
			err:           &gocloak.APIError{Code: http.StatusNotFound, Message: "Resource not found"},
			expectedError: fmt.Errorf("keycloak error: Resource not found: %w", ErrNotFound),
		},
		{
			name:          "Conflict Error",
			err:           &gocloak.APIError{Code: http.StatusConflict, Message: "Conflict occurred"},
			expectedError: fmt.Errorf("keycloak error: Conflict occurred: %w", ErrConflict),
		},
		{
			name:          "Bad Request Error",
			err:           &gocloak.APIError{Code: http.StatusBadRequest, Message: "Invalid input"},
			expectedError: fmt.Errorf("keycloak error: Invalid input: %w", ErrInvalidInput),
		},
		{
			name:          "Bad Request invalid_grant Error",
			err:           &gocloak.APIError{Code: http.StatusBadRequest, Message: "400 Bad Request: invalid_grant: Invalid user credentials"},
			expectedError: fmt.Errorf("keycloak error: 400 Bad Request: invalid_grant: Invalid user credentials: %w", ErrUnauthorized),
		},
		{
			name:          "Unauthorized Error",
			err:           &gocloak.APIError{Code: http.StatusUnauthorized, Message: "Unauthorized request"},
			expectedError: fmt.Errorf("keycloak error: Unauthorized request: %w", ErrUnauthorized),
		},
		{
			name:          "Forbidden Error",
			err:           &gocloak.APIError{Code: http.StatusForbidden, Message: "Unauthorized request"},
			expectedError: fmt.Errorf("keycloak error: Unauthorized request: %w", ErrUnauthorized),
		},
		{
			name:          "Internal Server Error",
			err:           &gocloak.APIError{Code: http.StatusInternalServerError, Message: "Server error"},
			expectedError: fmt.Errorf("keycloak error: Server error: %w", ErrInternal),
		},
		{
			name:          "Non-APIError Error",
			err:           fmt.Errorf("Some non-Keycloak error"),
			expectedError: fmt.Errorf("keycloak error: Some non-Keycloak error: %w", ErrInternal),
		},
		{
			name:          "Nil Error",
			err:           nil,
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotErr := HandleKeyCloakError(tt.err)
			assert.Equal(t, tt.expectedError, gotErr)
		})
	}
}
