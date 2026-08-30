package utils

import (
	"context"
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
