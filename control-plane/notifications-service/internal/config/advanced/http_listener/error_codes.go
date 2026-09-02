/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http_listener

import (
	"fmt"

	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
)

type ErrorCode int

const (
	ErrSuccess ErrorCode = iota
	ErrUnclassified
	ErrInvalidInput
	ErrFailedCreatingASink
	ErrValidationFailed
	ErrDeletionFailed
)

type ApiError struct {
	originalError error
	api.ApiError
}

func NewApiError(code int32, message string, err error) *ApiError {
	return &ApiError{
		ApiError: api.ApiError{
			Code:    code,
			Message: message,
		},
		originalError: err,
	}
}

func (e *ApiError) Error() string {
	if e.originalError == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Message, e.originalError.Error())
}

func (e *ApiError) Unwrap() error {
	return e.originalError
}

func newErrorResponse(status int, message string, err error) (api.ImplResponse, error) {
	return api.ImplResponse{Code: status}, NewApiError(int32(status), message, err)
}
