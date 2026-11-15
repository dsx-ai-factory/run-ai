/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleError_IsNotFound(t *testing.T) {
	resp, err := HandleError(context.Background(), "user", "id", "idValue", getError(http.StatusNotFound), "defaultLog")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestHandleError_IsConflict(t *testing.T) {
	resp, err := HandleError(context.Background(), "user", "id", "idValue", getError(http.StatusConflict), "defaultLog")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestHandleError_IsBadRequest(t *testing.T) {
	resp, err := HandleError(context.Background(), "user", "id", "idValue", getError(http.StatusBadRequest), "defaultLog")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestHandleError_IsForbidden(t *testing.T) {
	resp, err := HandleError(context.Background(), "user", "id", "idValue", getError(http.StatusForbidden), "defaultLog")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestHandleError_InternalServerError(t *testing.T) {
	resp, err := HandleError(context.Background(), "user", "id", "idValue", fmt.Errorf("500"), "defaultLog")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestGetHttpStatusCodeAndMessage_Error(t *testing.T) {
	code, message := GetHttpStatusCodeAndMessage(Error{Code: http.StatusBadRequest, Message: "failed to decode base64 string"})
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "bad request", message)
}

// Callers wrap Error with the underlying cause so the log keeps the detail, so
// the code has to stay reachable through the wrapping.
func TestGetHttpStatusCodeAndMessage_WrappedError(t *testing.T) {
	wrapped := fmt.Errorf("%w:%w",
		Error{Code: http.StatusBadRequest, Message: "failed to decode base64 string"},
		errors.New("illegal base64 data at input byte 3"))
	code, message := GetHttpStatusCodeAndMessage(wrapped)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "bad request", message)
}

func TestGetHttpStatusCodeAndMessage_Sentinel(t *testing.T) {
	code, message := GetHttpStatusCodeAndMessage(fmt.Errorf("metadata is unparseable: %w", ErrInvalidInput))
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "bad request", message)
}

func TestGetHttpStatusCodeAndMessage_UnrecognizedError(t *testing.T) {
	code, message := GetHttpStatusCodeAndMessage(errors.New("something went wrong"))
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, "internal server error", message)
}

func getError(code int) error {
	if code == http.StatusNotFound {
		return ErrNotFound
	}
	if code == http.StatusConflict {
		return ErrConflict
	}
	if code == http.StatusBadRequest {
		return ErrInvalidInput
	}
	if code == http.StatusForbidden {
		return ErrForbidden
	}
	return ErrInternal
}
