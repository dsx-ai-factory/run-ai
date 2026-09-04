/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toggles

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/toggles/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func resetGlobalState() {
	instance = &dummyToggler{}
	once = sync.Once{}
	log = &dummyLogger{}
}

func TestIsEnabled_WithMock(t *testing.T) {
	defer resetGlobalState()

	m := mocks.NewToggler(t)
	m.On("IsEnabled", mock.Anything, "feature-a").Return(true)
	m.On("IsEnabled", mock.Anything, "feature-b").Return(false)
	m.On("IsEnabled", mock.Anything, "feature-unknown").Return(false)
	SetMock(m)

	ctx := context.Background()

	assert.True(t, IsEnabled(ctx, "feature-a"))
	assert.False(t, IsEnabled(ctx, "feature-b"))
	assert.False(t, IsEnabled(ctx, "feature-unknown"))
}

func TestIsEnabled_DummyToggler_AlwaysTrue(t *testing.T) {
	defer resetGlobalState()

	instance = &dummyToggler{}
	ctx := context.Background()

	assert.True(t, IsEnabled(ctx, "any-feature"))
	assert.True(t, IsEnabled(ctx, "another-feature"))
}

func TestClose_ResetsState(t *testing.T) {
	defer resetGlobalState()

	m := mocks.NewToggler(t)
	m.On("Close").Return(nil)
	SetMock(m)

	err := Close()
	assert.NoError(t, err)

	// After close, should be back to dummyToggler
	ctx := context.Background()
	assert.True(t, IsEnabled(ctx, "any-feature"))
}

func TestClose_ReturnsError(t *testing.T) {
	defer resetGlobalState()

	m := mocks.NewToggler(t)
	m.On("Close").Return(fmt.Errorf("close error"))
	SetMock(m)

	err := Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close error")
}

func TestSetLogger_NilIgnored(t *testing.T) {
	original := log
	SetLogger(nil)
	assert.Equal(t, original, log)
}

func TestSetLogger_SetsLogger(t *testing.T) {
	defer resetGlobalState()

	custom := &dummyLogger{}
	SetLogger(custom)
	assert.Equal(t, custom, log)
}

func TestSetMock_SetsInstance(t *testing.T) {
	defer resetGlobalState()

	m := mocks.NewToggler(t)
	SetMock(m)

	assert.Equal(t, m, instance)
}
