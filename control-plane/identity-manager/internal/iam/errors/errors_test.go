/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestHandleKeyCloakError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedError error
		composedError error
	}{
		{
			name:          "Not Found Error",
			err:           &gocloak.APIError{Code: http.StatusNotFound, Message: "Resource not found"},
			expectedError: fmt.Errorf("keycloak error: Resource not found: %w", ErrNotFound),
			composedError: ErrNotFound,
		},
		{
			name:          "Conflict Error",
			err:           &gocloak.APIError{Code: http.StatusConflict, Message: "Conflict occurred"},
			expectedError: fmt.Errorf("keycloak error: Conflict occurred: %w", ErrConflict),
			composedError: ErrConflict,
		},
		{
			name:          "Bad Request Error",
			err:           &gocloak.APIError{Code: http.StatusBadRequest, Message: "Invalid input"},
			expectedError: fmt.Errorf("keycloak error: Invalid input: %w", ErrInvalidInput),
			composedError: ErrInvalidInput,
		},
		{
			name:          "Bad Request invalid_grant Error",
			err:           &gocloak.APIError{Code: http.StatusBadRequest, Message: "400 Bad Request: invalid_grant: Invalid user credentials"},
			expectedError: fmt.Errorf("keycloak error: 400 Bad Request: invalid_grant: Invalid user credentials: %w", ErrUnauthorized),
			composedError: ErrUnauthorized,
		},
		{
			name:          "Unauthorized Error",
			err:           &gocloak.APIError{Code: http.StatusUnauthorized, Message: "Unauthorized request"},
			expectedError: fmt.Errorf("keycloak error: Unauthorized request: %w", ErrUnauthorized),
			composedError: ErrUnauthorized,
		},
		{
			name:          "Forbidden Error",
			err:           &gocloak.APIError{Code: http.StatusForbidden, Message: "Unauthorized request"},
			expectedError: fmt.Errorf("keycloak error: Unauthorized request: %w", ErrUnauthorized),
			composedError: ErrUnauthorized,
		},
		{
			name:          "Internal Server Error",
			err:           &gocloak.APIError{Code: http.StatusInternalServerError, Message: "Server error"},
			expectedError: fmt.Errorf("keycloak error: Server error: %w", ErrInternal),
			composedError: ErrInternal,
		},
		{
			name:          "Non-APIError Error",
			err:           fmt.Errorf("Some non-Keycloak error"),
			expectedError: fmt.Errorf("keycloak error: Some non-Keycloak error: %w", ErrInternal),
			composedError: ErrInternal,
		},
		{
			name:          "Nil Error",
			err:           nil,
			expectedError: nil,
			composedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotErr := HandleKeyCloakError(tt.err)
			assert.Equal(t, tt.expectedError, gotErr)
			assert.True(t, errors.Is(gotErr, tt.composedError))
		})
	}
}
