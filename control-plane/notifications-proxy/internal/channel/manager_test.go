/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package channel_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/channel"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/channel/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/email"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/slack"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

const tenantId = "1"

var ctx = context.Background()

func TestGetNotificationChannels_Succeed(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)

	sc.On("GetTenantChannels", ctx, mock.Anything).Return([]channel.NotificationChannelMetadata{{
		Id:   "runai-email",
		Type: "email",
	}, {
		Id:   fmt.Sprintf("slack-%s", tenantId),
		Type: "slack",
	}}, nil)

	res, err := manager.GetNotificationsChannels(ctx, tenantId)
	expected := []channel.NotificationChannelMetadata{{
		Id:   "runai-email",
		Type: "email",
	}, {
		Id:   fmt.Sprintf("slack-%s", tenantId),
		Type: "slack",
	}}
	require.NoError(t, err)
	require.Equal(t, expected, res)
}

func TestGetNotificationChannels_Failed(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)

	sc.On("GetTenantChannels", ctx, mock.Anything).Return(nil, errors.Join(fmt.Errorf("failed to get sinks"), channel.ErrClientFailure))
	res, err := manager.GetNotificationsChannels(ctx, tenantId)

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.ErrorContains(t, err, "failed to get sinks")
	require.Nil(t, res)
}

func TestPatchChannel(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	overrideConfig := channel.NotificationChannel{Email: email.ChannelConfig{
		Username:   "newUsername",
		Password:   "newPassowrd",
		Recipients: []string{"recipients"},
	}}

	sc.On("Patch", ctx, tenantId, "channelId", mock.Anything).Return(nil)
	err := manager.PatchChannel(context.Background(), "1", "channelId", overrideConfig)
	require.NoError(t, err)
}

func TestPatchChannel_Failed(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	overrideConfig := channel.NotificationChannel{Email: email.ChannelConfig{}}

	sc.On("Patch", ctx, tenantId, "channelId", mock.Anything).Return(channel.ErrClientFailure)
	err := manager.PatchChannel(context.Background(), "1", "channelId", overrideConfig)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

func TestDeleteChannel_Succeed(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	channelId := "runai-email"

	sc.On("Delete", ctx, tenantId, channelId).Return(nil)
	err := manager.DeleteChannel(ctx, tenantId, channelId)
	require.NoError(t, err)
}

func TestDeleteChannel_ClientFails(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	channelId := "runai-email"

	sc.On("Delete", ctx, tenantId, channelId).Return(errors.Join(fmt.Errorf("delete failed"), channel.ErrClientFailure))
	err := manager.DeleteChannel(ctx, tenantId, channelId)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.ErrorContains(t, err, "delete failed")
}

func TestValidateChannel_Valid(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	cfg := channel.NotificationChannel{Type: "email", Email: email.ChannelConfig{SmtpHost: "smtp.example.com"}}

	sc.On("Validate", ctx, tenantId, cfg).Return(channel.ValidationResponse{Status: true}, nil)
	resp, err := manager.ValidateChannel(ctx, tenantId, cfg)
	require.NoError(t, err)
	require.True(t, resp.Status)
}

func TestValidateChannel_Invalid(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	cfg := channel.NotificationChannel{Type: "email", Email: email.ChannelConfig{}}

	sc.On("Validate", ctx, tenantId, cfg).Return(channel.ValidationResponse{Status: false}, nil)
	resp, err := manager.ValidateChannel(ctx, tenantId, cfg)
	require.NoError(t, err)
	require.False(t, resp.Status)
}

func TestValidateChannel_ClientFails(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	cfg := channel.NotificationChannel{Type: "slack", Slack: slack.ChannelConfig{}}

	sc.On("Validate", ctx, tenantId, cfg).Return(channel.ValidationResponse{Status: false}, channel.ErrClientFailure)
	resp, err := manager.ValidateChannel(ctx, tenantId, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.False(t, resp.Status)
}

func TestCreateChannel_Succeed(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	cfg := channel.NotificationChannel{Type: "email", Email: email.ChannelConfig{SmtpHost: "smtp.example.com"}}

	sc.On("Create", ctx, tenantId, cfg).Return(nil)
	err := manager.CreateChannel(ctx, tenantId, cfg)
	require.NoError(t, err)
}

func TestCreateChannel_AlreadyExists(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	cfg := channel.NotificationChannel{Type: "slack", Slack: slack.ChannelConfig{}}

	sc.On("Create", ctx, tenantId, cfg).Return(channel.ErrAlreadyExists)
	err := manager.CreateChannel(ctx, tenantId, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrAlreadyExists)
}

func TestCreateChannel_ClientFails(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	cfg := channel.NotificationChannel{Type: "email"}

	sc.On("Create", ctx, tenantId, cfg).Return(errors.Join(fmt.Errorf("create failed"), channel.ErrClientFailure))
	err := manager.CreateChannel(ctx, tenantId, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

func TestGetChannel_Succeed(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	channelId := "runai-email"
	expected := &channel.NotificationChannel{Type: "email", Email: email.ChannelConfig{SmtpHost: "host"}}

	sc.On("Get", ctx, tenantId, channelId).Return(expected, nil)
	res, err := manager.GetChannel(ctx, tenantId, channelId)
	require.NoError(t, err)
	require.Equal(t, expected, res)
}

func TestGetChannel_NotFound(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	channelId := "unknown-channel"

	sc.On("Get", ctx, tenantId, channelId).Return(nil, channel.ErrNotFound)
	res, err := manager.GetChannel(ctx, tenantId, channelId)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrNotFound)
	require.Nil(t, res)
}

func TestGetChannel_ClientFails(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	channelId := "runai-email"

	sc.On("Get", ctx, tenantId, channelId).Return(nil, channel.ErrClientFailure)
	res, err := manager.GetChannel(ctx, tenantId, channelId)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.Nil(t, res)
}

func TestUpdateChannel_Succeed(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	channelId := "runai-email"
	cfg := channel.NotificationChannel{Type: "email", Email: email.ChannelConfig{SmtpHost: "smtp.example.com"}}

	sc.On("Update", ctx, tenantId, channelId, cfg).Return(nil)
	err := manager.UpdateChannel(ctx, tenantId, channelId, cfg)
	require.NoError(t, err)
}

func TestUpdateChannel_ClientFails(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	channelId := fmt.Sprintf("slack-%s", tenantId)
	cfg := channel.NotificationChannel{Type: "slack", Slack: slack.ChannelConfig{}}

	sc.On("Update", ctx, tenantId, channelId, cfg).Return(channel.ErrClientFailure)
	err := manager.UpdateChannel(ctx, tenantId, channelId, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

func TestNewManager(t *testing.T) {
	sc := mocks.NewClienter(t)
	manager := channel.NewManager(sc)
	require.NotNil(t, manager)
}
