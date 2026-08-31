/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toggles

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/toggles/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUnleashToggler_Close_NilClient_ReturnsError(t *testing.T) {
	ut := &unleashToggler{}
	err := ut.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestUnleashToggler_Close_Success(t *testing.T) {
	m := mocks.NewUnleashClienter(t)
	m.On("Close").Return(nil)
	ut := &unleashToggler{client: m}
	err := ut.Close()
	assert.NoError(t, err)
}

func TestUnleashToggler_IsEnabled_NilClient_ReturnsFalse(t *testing.T) {
	ut := &unleashToggler{}
	assert.False(t, ut.IsEnabled(context.Background(), "any-feature"))
}

func TestUnleashToggler_IsEnabled_NilClient_WithTenantId_ReturnsFalse(t *testing.T) {
	ut := &unleashToggler{}
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "test-tenant")
	assert.False(t, ut.IsEnabled(ctx, "some-feature"))
}

func TestUnleashToggler_IsEnabled_DelegatesToClient(t *testing.T) {
	m := mocks.NewUnleashClienter(t)
	m.On("IsEnabled", "my-feature", mock.Anything).Return(true)
	ut := &unleashToggler{client: m}

	result := ut.IsEnabled(context.Background(), "my-feature")
	assert.True(t, result)
}

func TestUnleashToggler_IsEnabled_PassesTenantIdInContext(t *testing.T) {
	m := mocks.NewUnleashClienter(t)
	m.On("IsEnabled", "some-feature", mock.Anything).Return(false)
	ut := &unleashToggler{client: m}

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "tenant-123")
	ut.IsEnabled(ctx, "some-feature")

	m.AssertCalled(t, "IsEnabled", "some-feature", mock.Anything)
}

func TestUnleashToggler_IsEnabled_NoTenantId_StillCallsClient(t *testing.T) {
	m := mocks.NewUnleashClienter(t)
	m.On("IsEnabled", "some-feature", mock.Anything).Return(false)
	ut := &unleashToggler{client: m}

	result := ut.IsEnabled(context.Background(), "some-feature")
	assert.False(t, result)
	m.AssertCalled(t, "IsEnabled", "some-feature", mock.Anything)
}
