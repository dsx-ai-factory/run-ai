/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http_listener

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/mocks"
	"github.com/stretchr/testify/assert"
)

func TestIsHealthy_ShouldReturnTrue(t *testing.T) {
	dao := mocks.NewIsHealthyer(t)
	has := HealthAPIService{
		dao: dao,
	}
	dao.On("IsHealthy").Return(true, nil)

	res, err := has.Health(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "I'm alive!", res.Body)
}

func TestIsHealthy_ShouldReturnFalse(t *testing.T) {
	dao := mocks.NewIsHealthyer(t)
	has := HealthAPIService{
		dao: dao,
	}
	dao.On("IsHealthy").Return(false, errors.New("some error"))

	res, err := has.Health(context.Background())

	apiErr := err.(*ApiError)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "Not healthy", apiErr.Message)
}
