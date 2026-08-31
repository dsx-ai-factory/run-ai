/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package slack

import (
	"context"
	"errors"
	"testing"
	"time"

	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewSlackStateEncryptor(t *testing.T) {
	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("RegisterStore", cache_testing.KeyFromType[SlackAuthSessionState](), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewSlackAuthSessionHandler()
	require.NotNil(t, handler)
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("RegisterStore", cache_testing.KeyFromType[SlackAuthSessionState](), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewSlackAuthSessionHandler()
	require.NotNil(t, handler)

	mockCacher.On("Set", ctx, cache_testing.KeyFromType[SlackAuthSessionState](), mock.Anything, mock.MatchedBy(func(s SlackAuthSessionState) bool {
		return s.Email == "test@run.ai" && s.TenantId == "1" && time.Now().Unix()-s.Timestamp < 10
	})).Return(nil)
	stateId, err := handler.Create(ctx, "1", "test@run.ai")
	require.NoError(t, err)
	require.NotEmpty(t, stateId)
}

func TestCreate_FailedToSetInCache(t *testing.T) {
	ctx := context.Background()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("RegisterStore", cache_testing.KeyFromType[SlackAuthSessionState](), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewSlackAuthSessionHandler()
	require.NotNil(t, handler)

	mockCacher.On("Set", ctx, cache_testing.KeyFromType[SlackAuthSessionState](), mock.Anything, mock.Anything).Return(errors.New("cache error"))
	stateId, err := handler.Create(ctx, "1", "test@run.ai")
	require.EqualError(t, err, "failed to store session state: cache error")
	require.Empty(t, stateId)
}

func TestParse(t *testing.T) {
	ctx := context.Background()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("RegisterStore", cache_testing.KeyFromType[SlackAuthSessionState](), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewSlackAuthSessionHandler()
	require.NotNil(t, handler)

	mockCacher.On("Get", ctx, cache_testing.KeyFromType[SlackAuthSessionState](), "stateId").Return(&SlackAuthSessionState{
		TenantId:  "1",
		Email:     "test@run.ai",
		Timestamp: time.Now().Add(-30 * time.Second).Unix(),
	}, nil)

	state, err := handler.Parse(ctx, "stateId")
	require.NoError(t, err)
	require.Equal(t, "test@run.ai", state.Email)
	require.Equal(t, "1", state.TenantId)
}

func TestParse_FailedToGetFromCache(t *testing.T) {
	ctx := context.Background()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("RegisterStore", cache_testing.KeyFromType[SlackAuthSessionState](), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewSlackAuthSessionHandler()
	require.NotNil(t, handler)

	mockCacher.On("Get", ctx, cache_testing.KeyFromType[SlackAuthSessionState](), "stateId").Return(nil, errors.New("cache error"))

	state, err := handler.Parse(ctx, "stateId")
	require.EqualError(t, err, "failed to get session state: cache error")
	require.Nil(t, state)
}

func TestParse_StateNotValid(t *testing.T) {
	ctx := context.Background()

	mockCacher := cache_testing.SetMock(t)
	mockCacher.On("RegisterStore", cache_testing.KeyFromType[SlackAuthSessionState](), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	handler := NewSlackAuthSessionHandler()
	require.NotNil(t, handler)

	mockCacher.On("Get", ctx, cache_testing.KeyFromType[SlackAuthSessionState](), "stateId").Return(&SlackAuthSessionState{
		TenantId:  "1",
		Email:     "test@run.ai",
		Timestamp: time.Now().Add(-2 * slackAuthExpirationSeconds * time.Second).Unix(),
	}, nil)

	state, err := handler.Parse(ctx, "stateId")
	require.EqualError(t, err, "expired slack state")
	require.Nil(t, state)
}

func TestValidate_NilState(t *testing.T) {
	handler := &SlackAuthSessionHandler{}
	err := handler.validate(nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid slack state: session state is nil")
}

func TestValidate_EmptyTenantId(t *testing.T) {
	slackState := &SlackAuthSessionState{
		Timestamp: time.Now().Unix(),
		Email:     "test@run.ai",
	}

	handler := &SlackAuthSessionHandler{}
	err := handler.validate(slackState)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid slack state: tenantId is empty")
}

func TestValidate_EmptyEmail(t *testing.T) {
	slackState := &SlackAuthSessionState{
		Timestamp: time.Now().Unix(),
		TenantId:  "1",
	}

	handler := &SlackAuthSessionHandler{}
	err := handler.validate(slackState)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid slack state: email is empty")
}

func TestValidate_TimestampExpired(t *testing.T) {
	slackState := &SlackAuthSessionState{
		TenantId:  "1",
		Email:     "test@run.ai",
		Timestamp: time.Now().Add(-3 * time.Minute).Unix(),
	}

	handler := &SlackAuthSessionHandler{}
	err := handler.validate(slackState)
	assert.Error(t, err)
	assert.EqualError(t, err, "expired slack state")
}
