/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package channel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/channel"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/channel/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/email"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/slack"
	proxyapi "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

var (
	tenantId = "1"
	ctx      = context.Background()
)

// ------------------------------------------- START Test GetTenantChannels --------------------------
func TestGetNotificationChannels_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := &ChannelClient{sinksClient: sc}

	sinkConfig := &types.SinkConfig{
		Id:   EmailChannelName(),
		Type: "email",
	}
	sink2Config := &types.SinkConfig{
		Id:   SlackChannelName(tenantId),
		Type: "slack",
	}
	sink3Config := &types.SinkConfig{
		Id:   SlackChannelName("2"),
		Type: "slack",
	}
	sink4Config := &types.SinkConfig{
		Id:   "invalid",
		Type: "email",
	}

	sc.On("GetSinks", ctx).Return([]*types.SinkConfig{sinkConfig, sink2Config, sink3Config, sink4Config}, nil)
	res, err := manager.GetTenantChannels(ctx, tenantId)
	expected := []channel.NotificationChannelMetadata{{
		Type: sinkConfig.Type,
		Id:   sinkConfig.Id,
	}, {
		Type: sink2Config.Type,
		Id:   sink2Config.Id,
	}}

	require.NoError(t, err)
	require.Equal(t, expected, res)
}

func TestGetNotificationChannels_Failed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	sc.On("GetSinks", ctx).Return(nil, fmt.Errorf("failed to get sinks"))
	res, err := manager.GetTenantChannels(ctx, tenantId)

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.ErrorContains(t, err, "failed to get sinks")
	require.Nil(t, res)
}

// ------------------------------------------- END Test GetTenantChannels --------------------------

// ------------------------------------------- START Test Delete --------------------------
func TestDelete_NotValidChannelId(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	err := manager.Delete(ctx, tenantId, SlackChannelName("123"))

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrChannelNameNotValid)
	require.ErrorContains(t, err, "channel id not recognized")
}

func TestDelete_Failed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	sc.On("DeleteSink", ctx, SlackChannelName(tenantId)).Return(fmt.Errorf("failed to delete sink"))
	err := manager.Delete(ctx, tenantId, SlackChannelName(tenantId))

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.ErrorContains(t, err, "failed to delete sink")
}

func TestDelete_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	sc.On("DeleteSink", ctx, EmailChannelName()).Return(nil)
	err := manager.Delete(ctx, tenantId, EmailChannelName())

	require.NoError(t, err)
}

// ------------------------------------------- END Test Delete --------------------------

// ------------------------------------------- START Test Create --------------------------
func TestCreateChannel_Slack_NotFound(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	slackChannelId := SlackChannelName(tenantId)
	sc.On("GetSink", ctx, slackChannelId).Return(nil, &proxyapi.Error{Code: http.StatusNotFound})
	sc.On("CreateSink", ctx, slackChannelId, mock.Anything).Return(nil)
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Slack: slack.ChannelConfig{ApiToken: "xoxp-token"},
		Type:  NotificationChannelTypeSlack,
	})

	require.NoError(t, err)
}

func TestCreateChannel_Slack_FailedToGetSinks(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	slackChannelId := SlackChannelName(tenantId)
	sc.On("GetSink", ctx, slackChannelId).Return(nil, fmt.Errorf("connection error"))
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Slack: slack.ChannelConfig{ApiToken: "xoxp-token"},
		Type:  NotificationChannelTypeSlack,
	})

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

func TestCreateChannel_Slack_AlreadyExists(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	slackChannelId := SlackChannelName(tenantId)
	sc.On("GetSink", ctx, slackChannelId).Return(&types.SinkConfig{}, nil)
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Slack: slack.ChannelConfig{ApiToken: "xoxp-token"},
		Type:  NotificationChannelTypeSlack,
	})

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrAlreadyExists)
}

func TestCreateChannel_Slack_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	slackChannelId := SlackChannelName(tenantId)
	sc.On("GetSink", ctx, slackChannelId).Return(nil, nil)
	sc.On("CreateSink", ctx, slackChannelId, mock.Anything).Return(nil)
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Slack: slack.ChannelConfig{ApiToken: "xoxp-token"},
		Type:  NotificationChannelTypeSlack,
	})

	require.NoError(t, err)
}

func TestCreateChannel_Slack_UpdateFailed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	slackChannelId := SlackChannelName(tenantId)
	sc.On("GetSink", ctx, slackChannelId).Return(nil, nil)
	sc.On("CreateSink", ctx, slackChannelId, mock.Anything).Return(errors.New("failed to create"))
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Slack: slack.ChannelConfig{ApiToken: "xoxp-token"},
		Type:  NotificationChannelTypeSlack,
	})

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

func TestCreateChannel_Email_NotFound(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, &proxyapi.Error{Code: http.StatusNotFound})
	sc.On("CreateSink", ctx, EmailChannelName(), mock.Anything).Return(nil)
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Email: email.ChannelConfig{},
		Type:  NotificationChannelTypeEmail,
	})

	require.NoError(t, err)
}

func TestCreateChannel_Email_FailedToGetSinks(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, fmt.Errorf("connection error"))
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Email: email.ChannelConfig{},
		Type:  NotificationChannelTypeEmail,
	})

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

func TestCreateChannel_Email_AlreadyExists(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("GetSink", ctx, EmailChannelName()).Return(&types.SinkConfig{}, nil)
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Email: email.ChannelConfig{},
		Type:  NotificationChannelTypeEmail,
	})

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrAlreadyExists)
}

func TestCreateChannel_Email_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, nil)
	sc.On("CreateSink", ctx, EmailChannelName(), mock.Anything).Return(nil)
	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Email: email.ChannelConfig{},
		Type:  NotificationChannelTypeEmail,
	})

	require.NoError(t, err)
}

func TestCreateChannel_UnsupportedType(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	err := cc.Create(ctx, tenantId, channel.NotificationChannel{
		Type: "unsupported",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrNotSupportedType)
}

// ------------------------------------------- END Test Create --------------------------

// ------------------------------------------- START Test Get --------------------------
func TestGetChannel_Email_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	sinkConfig := &types.SinkConfig{
		Id:   EmailChannelName(),
		Type: "email",
		Config: types.SinkConfigDetails{
			EmailSinkConfig: types.EmailSinkConfig{
				User: "user",
			},
		},
	}

	sc.On("GetSink", ctx, EmailChannelName()).Return(sinkConfig, nil)

	res, err := manager.Get(ctx, tenantId, EmailChannelName())
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, NotificationChannelTypeEmail, res.Type)
	require.Equal(t, "user", res.Email.Username)
}

func TestGetChannel_Slack_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	channelId := SlackChannelName(tenantId)
	sinkConfig := &types.SinkConfig{
		Id:   channelId,
		Type: "slack",
		Config: types.SinkConfigDetails{
			SlackSinkConfig: types.SlackSinkConfig{
				ApiToken: "xoxp-token",
			},
		},
	}

	sc.On("GetSink", ctx, channelId).Return(sinkConfig, nil)

	res, err := manager.Get(ctx, tenantId, channelId)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, NotificationChannelTypeSlack, res.Type)
	require.Equal(t, "xoxp-token", res.Slack.ApiToken)
}

func TestGetChannel_UnsupportedType(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	res, err := manager.Get(ctx, tenantId, "unknown-channel")
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrNotSupportedType)
	require.Nil(t, res)
}

func TestGetChannel_Email_NotFound(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, nil)

	res, err := cc.Get(ctx, tenantId, EmailChannelName())
	require.ErrorIs(t, err, channel.ErrNotFound)
	require.Nil(t, res)
}

func TestGetChannel_Slack_NotFound(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	channelId := SlackChannelName(tenantId)
	sc.On("GetSink", ctx, channelId).Return(nil, nil)

	res, err := cc.Get(ctx, tenantId, channelId)
	require.ErrorIs(t, err, channel.ErrNotFound)
	require.Nil(t, res)
}

func TestGetChannel_Email_Failed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, errors.New("connection error"))

	res, err := cc.Get(ctx, tenantId, EmailChannelName())
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.Nil(t, res)
}

// ------------------------------------------- END Test Get --------------------------

// ------------------------------------------- START Test Validate --------------------------
func TestValidateChannel_Email_Valid(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(&clients.SinkValidityResponse{}, nil)

	resp, err := manager.Validate(ctx, tenantId, channel.NotificationChannel{
		Type: NotificationChannelTypeEmail,
	})
	require.NoError(t, err)
	require.True(t, resp.Status)
}

func TestValidateChannel_Email_Invalid(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(&clients.SinkValidityResponse{ErrorCode: ptr.To(int32(4))}, nil)

	resp, err := manager.Validate(ctx, tenantId, channel.NotificationChannel{
		Type: NotificationChannelTypeEmail,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrValidationFailed)
	require.False(t, resp.Status)
}

func TestValidateChannel_Failed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(nil, fmt.Errorf("connection error"))

	resp, err := manager.Validate(ctx, tenantId, channel.NotificationChannel{
		Type: NotificationChannelTypeEmail,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.False(t, resp.Status)
}

func TestValidateChannel_UnsupportedType(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	resp, err := manager.Validate(ctx, tenantId, channel.NotificationChannel{
		Type: "unsupported",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrNotSupportedType)
	require.False(t, resp.Status)
}

func TestValidateChannel_Slack_Valid(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(&clients.SinkValidityResponse{}, nil)

	resp, err := cc.Validate(ctx, tenantId, channel.NotificationChannel{
		Type: NotificationChannelTypeSlack,
	})
	require.NoError(t, err)
	require.True(t, resp.Status)
}

func TestValidateChannel_Slack_Invalid(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(&clients.SinkValidityResponse{ErrorCode: ptr.To(int32(4))}, nil)

	resp, err := cc.Validate(ctx, tenantId, channel.NotificationChannel{
		Type: NotificationChannelTypeSlack,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrValidationFailed)
	require.False(t, resp.Status)
}

func TestValidateChannel_ParseError(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(&clients.SinkValidityResponse{ErrorCode: ptr.To(int32(1))}, nil)

	resp, err := cc.Validate(ctx, tenantId, channel.NotificationChannel{
		Type: NotificationChannelTypeEmail,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrValidationFailed)
	require.False(t, resp.Status)
}

// ------------------------------------------- END Test Validate --------------------------

// ------------------------------------------- START Test Update --------------------------
func TestUpdate_Email_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	sc.On("DeleteSink", ctx, EmailChannelName()).Return(nil)
	sc.On("CreateSink", ctx, EmailChannelName(), mock.Anything).Return(nil)

	err := manager.Update(ctx, tenantId, EmailChannelName(), channel.NotificationChannel{
		Email: email.ChannelConfig{Username: "user"},
	})
	require.NoError(t, err)
}

func TestUpdate_Slack_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	channelId := SlackChannelName(tenantId)
	sc.On("DeleteSink", ctx, channelId).Return(nil)
	sc.On("CreateSink", ctx, channelId, mock.Anything).Return(nil)

	err := manager.Update(ctx, tenantId, channelId, channel.NotificationChannel{
		Slack: slack.ChannelConfig{ApiToken: "xoxp-token"},
	})
	require.NoError(t, err)
}

func TestUpdate_UnsupportedType(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: sc}

	err := manager.Update(ctx, tenantId, "unknown-channel", channel.NotificationChannel{})
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrNotSupportedType)
}

func TestUpdate_ChannelNotExists(t *testing.T) {
	mockSinksClient := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: mockSinksClient}

	channelConfig := channel.NotificationChannel{Email: email.ChannelConfig{
		Username: "user",
	}}

	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(&proxyapi.Error{Code: http.StatusNotFound})

	err := manager.Update(ctx, tenantId, EmailChannelName(), channelConfig)
	require.ErrorIs(t, err, channel.ErrNotFound)
}

func TestUpdate_UpdateFailed(t *testing.T) {
	mockSinksClient := mocks.NewSinksClienter(t)
	manager := ChannelClient{sinksClient: mockSinksClient}

	channelConfig := channel.NotificationChannel{Email: email.ChannelConfig{
		Username: "user",
	}}

	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(nil)
	mockSinksClient.On("CreateSink", ctx, EmailChannelName(), mock.Anything).Return(errors.New("error"))

	err := manager.Update(ctx, tenantId, EmailChannelName(), channelConfig)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

// ------------------------------------------- END Test Update --------------------------

// ------------------------------------------- START Test Sink Config To Channel Config --------------------------
func TestSinkConfigToEmailChannelConfig(t *testing.T) {
	sinkConfig := types.SinkConfigDetails{
		EmailSinkConfig: types.EmailSinkConfig{
			User:            "user",
			Password:        "password",
			SmtpHost:        "host",
			SmtpPort:        587,
			From:            "from",
			AuthType:        "auth",
			SmtpTlsEnabled:  ptr.To(true),
			FromDisplayName: "display",
			UseAttachments:  ptr.To(true),
			Subject:         "subject",
		},
		Recipients:          []string{"r1", "r2"},
		DirectNotifications: ptr.To(true),
	}

	res := emailChannelConfig(sinkConfig)

	require.Equal(t, "user", res.Username)
	require.Equal(t, "password", res.Password)
	require.Equal(t, "host", res.SmtpHost)
	require.Equal(t, int32(587), res.SmtpPort)
	require.Equal(t, "from", res.From)
	require.Equal(t, "auth", res.AuthType)
	require.True(t, *res.SmtpTlsEnabled)
	require.Equal(t, []string{"r1", "r2"}, res.Recipients)
	require.Equal(t, "display", res.FromDisplayName)
	require.True(t, *res.DirectNotifications)
	require.True(t, *res.UseAttachments)
	require.Equal(t, "subject", res.Subject)
}

func TestEmailConfigToSinkConfig(t *testing.T) {
	config := email.ChannelConfig{
		Username:            "user",
		Password:            "password",
		SmtpHost:            "host",
		SmtpPort:            587,
		From:                "from",
		AuthType:            "auth",
		SmtpTlsEnabled:      ptr.To(true),
		Recipients:          []string{"r1", "r2"},
		FromDisplayName:     "display",
		DirectNotifications: ptr.To(true),
		UseAttachments:      ptr.To(true),
		Subject:             "subject",
	}

	res := emailConfigToSinkConfig(config)

	require.Equal(t, "user", res.Config.User)
	require.Equal(t, "password", res.Config.Password)
	require.Equal(t, "host", res.Config.SmtpHost)
	require.Equal(t, 587, res.Config.SmtpPort)
	require.Equal(t, "from", res.Config.From)
	require.Equal(t, "auth", res.Config.AuthType)
	require.True(t, *res.Config.SmtpTlsEnabled)
	require.Equal(t, []string{"r1", "r2"}, res.Config.Recipients)
	require.Equal(t, "display", res.Config.FromDisplayName)
	require.True(t, *res.Config.DirectNotifications)
	require.True(t, *res.Config.UseAttachments)
	require.Equal(t, "subject", res.Config.Subject)
}

func TestEmailSinkConfigToChannelConfig(t *testing.T) {
	useAttachments := true
	emailSinkConfig := types.SinkConfigDetails{
		EmailSinkConfig: types.EmailSinkConfig{
			From:           "from",
			Subject:        "subject",
			AuthType:       "authType",
			User:           "user",
			Password:       "password",
			UseAttachments: &useAttachments,
		},
	}
	res := emailChannelConfig(emailSinkConfig)
	expected := email.ChannelConfig{
		Username:       emailSinkConfig.User,
		Password:       emailSinkConfig.Password,
		From:           emailSinkConfig.From,
		Subject:        emailSinkConfig.Subject,
		AuthType:       emailSinkConfig.AuthType,
		UseAttachments: &useAttachments,
	}
	require.Equal(t, expected, *res)
}

func TestChannelConfigToEmailSinkConfig(t *testing.T) {
	useAttachments := true
	channelConfig := email.ChannelConfig{
		Username:       "user",
		Password:       "password",
		From:           "from",
		Subject:        "subject",
		AuthType:       "authType",
		UseAttachments: &useAttachments,
	}
	res := emailConfigToSinkConfig(channelConfig)
	expected := types.SinkConfigDetails{
		EmailSinkConfig: types.EmailSinkConfig{
			User:           channelConfig.Username,
			Password:       channelConfig.Password,
			From:           channelConfig.From,
			Subject:        channelConfig.Subject,
			AuthType:       channelConfig.AuthType,
			UseAttachments: &useAttachments,
		},
	}
	require.Equal(t, expected, res.Config)
}

func TestInferSinkValidity(t *testing.T) {
	tests := []struct {
		name     string
		input    clients.SinkValidityResponse
		expected bool
		hasError bool
	}{
		{"valid response", clients.SinkValidityResponse{ErrorCode: ptr.To(int32(0))}, true, false},
		{"valid nil code", clients.SinkValidityResponse{}, true, false},
		{"invalid response", clients.SinkValidityResponse{ErrorCode: ptr.To(int32(4))}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := inferSinkValidity(tt.input, errToChannelErrMapping)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, resp.Status)
		})
	}
}

// ------------------------------------------- END Test Sink Config To Channel Config --------------------------

// ------------------------------------------- START Test Get Slack --------------------------
func TestGetSlack_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	channelId := SlackChannelName(tenantId)
	sinkConfig := &types.SinkConfig{
		Id:   channelId,
		Type: "slack",
		Config: types.SinkConfigDetails{
			SlackSinkConfig: types.SlackSinkConfig{
				ApiToken: "xoxp-token",
			},
			Recipients:          []string{"channel1"},
			DirectNotifications: ptr.To(true),
		},
	}

	sc.On("GetSink", ctx, channelId).Return(sinkConfig, nil)

	res, err := cc.getSlack(ctx, tenantId, errToChannelErrMapping)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "xoxp-token", res.Slack.ApiToken)
	require.Equal(t, []string{"channel1"}, res.Slack.Recipients)
	require.True(t, *res.Slack.DirectNotifications)
}

func TestGetSlack_NotFound(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	channelId := SlackChannelName(tenantId)
	sc.On("GetSink", ctx, channelId).Return(nil, nil)

	res, err := cc.getSlack(ctx, tenantId, errToChannelErrMapping)
	require.ErrorIs(t, err, channel.ErrNotFound)
	require.Nil(t, res)
}

func TestGetSlack_Failed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	channelId := SlackChannelName(tenantId)
	sc.On("GetSink", ctx, channelId).Return(nil, errors.New("connection error"))

	res, err := cc.getSlack(ctx, tenantId, errToChannelErrMapping)
	require.ErrorIs(t, err, channel.ErrClientFailure)
	require.Nil(t, res)
}

// ------------------------------------------- END Test Get Slack --------------------------

// ------------------------------------------- START Test Patch --------------------------

func TestPatchChannel(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: sc}

	currentSink := &types.SinkConfig{
		Id:   EmailChannelName(),
		Type: "email",
		Config: types.SinkConfigDetails{
			EmailSinkConfig: types.EmailSinkConfig{
				User:     "username",
				Password: "passowrd",
				SmtpHost: "host",
			},
		},
	}

	mergedConfig := types.SinkConfigDetails{
		EmailSinkConfig: types.EmailSinkConfig{
			User:     "newUsername",
			Password: "newPassowrd",
			SmtpHost: "host",
		},
		DirectNotifications: ptr.To(true),
		Recipients:          []string{"recipients"},
	}

	overrideConfig := channel.NotificationChannel{Email: email.ChannelConfig{
		Username:   "newUsername",
		Password:   "newPassowrd",
		Recipients: []string{"recipients"},
	}}

	sc.On("GetSink", ctx, EmailChannelName()).Return(currentSink, nil)
	sc.On("UpdateSink", ctx, EmailChannelName(), mock.MatchedBy(func(cfg *types.SinkConfig) bool {
		if cfg.Config.EmailSinkConfig != mergedConfig.EmailSinkConfig {
			return false
		}
		if len(cfg.Config.Recipients) != len(mergedConfig.Recipients) {
			return false
		}
		for i, recipient := range cfg.Config.Recipients {
			if recipient != mergedConfig.Recipients[i] {
				return false
			}
		}
		return true
	})).Return(nil)

	err := cc.Patch(ctx, tenantId, EmailChannelName(), overrideConfig)
	require.NoError(t, err)
}

func TestPatchChannel_GetFailed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: sc}
	overrideConfig := channel.NotificationChannel{Email: email.ChannelConfig{}}

	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, fmt.Errorf("connection error"))

	err := cc.Patch(ctx, tenantId, EmailChannelName(), overrideConfig)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

func TestPatchChannel_NotFound(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: sc}
	overrideConfig := channel.NotificationChannel{Email: email.ChannelConfig{}}

	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, nil)

	err := cc.Patch(ctx, tenantId, EmailChannelName(), overrideConfig)
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrNotFound)
}

func TestPatchChannel_UpdateFailed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: sc}

	currentSink := &types.SinkConfig{
		Id:     EmailChannelName(),
		Type:   NotificationChannelTypeEmail,
		Config: types.SinkConfigDetails{EmailSinkConfig: types.EmailSinkConfig{User: "old"}},
	}
	sc.On("GetSink", ctx, EmailChannelName()).Return(currentSink, nil)
	sc.On("UpdateSink", ctx, EmailChannelName(), mock.Anything).Return(errors.New("update failed"))

	err := cc.Patch(ctx, tenantId, EmailChannelName(), channel.NotificationChannel{Email: email.ChannelConfig{Username: "newUser"}})
	require.Error(t, err)
	require.ErrorIs(t, err, channel.ErrClientFailure)
}

// ------------------------------------------- END Test Patch --------------------------

// ------------------------------------------- START Test IsExists --------------------------
func TestIsExists_Succeed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	state := &mockState{chanName: EmailChannelName(), chanType: "email"}
	sc.On("GetSink", ctx, EmailChannelName()).Return(&types.SinkConfig{Type: "email"}, nil)

	exists, err := cc.IsExists(ctx, tenantId, state)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestIsExists_NotExists(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	state := &mockState{chanName: EmailChannelName(), chanType: "email"}
	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, nil)

	exists, err := cc.IsExists(ctx, tenantId, state)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestIsExists_Failed(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	state := &mockState{chanName: EmailChannelName(), chanType: "email"}
	sc.On("GetSink", ctx, EmailChannelName()).Return(nil, errors.New("connection error"))

	exists, err := cc.IsExists(ctx, tenantId, state)
	require.ErrorIs(t, err, notification_state.ErrClientFailure)
	require.False(t, exists)
}

func TestIsExists_TypeMismatch(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := ChannelClient{sinksClient: sc}

	state := &mockState{chanName: EmailChannelName(), chanType: "email"}
	sc.On("GetSink", ctx, EmailChannelName()).Return(&types.SinkConfig{Type: "slack"}, nil)

	exists, err := cc.IsExists(ctx, tenantId, state)
	require.ErrorIs(t, err, notification_state.ErrIncompatibleSinkType)
	require.False(t, exists)
}

type mockState struct {
	enabled  bool
	chanType string
	chanName string
}

func (m *mockState) IsEnabled() bool     { return m.enabled }
func (m *mockState) GetType() string     { return m.chanType }
func (m *mockState) GetChanName() string { return m.chanName }
func (m *mockState) GetMetadata() any    { return nil }

// ------------------------------------------- END Test IsExists --------------------------

// ------------------------------------------- START Test Email Channels Clienter --------------------------
func TestGetEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	sinkConfig := &types.SinkConfig{
		Id:   EmailChannelName(),
		Type: "email",
		Config: types.SinkConfigDetails{
			EmailSinkConfig: types.EmailSinkConfig{
				User:            "user",
				Password:        "password",
				SmtpHost:        "smtp.host",
				SmtpPort:        587,
				From:            "from",
				AuthType:        "auth_type",
				SmtpTlsEnabled:  ptr.To(true),
				FromDisplayName: "from_display_name",
				UseAttachments:  ptr.To(true),
				Subject:         "subject",
			},
			Recipients:          []string{"recipient1", "recipient2"},
			DirectNotifications: ptr.To(true),
		},
	}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(sinkConfig, nil)

	expectedChannelConfig := &email.ChannelConfig{
		Username:            "user",
		Password:            "password",
		SmtpHost:            "smtp.host",
		SmtpPort:            587,
		From:                "from",
		AuthType:            "auth_type",
		SmtpTlsEnabled:      ptr.To(true),
		FromDisplayName:     "from_display_name",
		DirectNotifications: ptr.To(true),
		UseAttachments:      ptr.To(true),
		Subject:             "subject",
		Recipients:          []string{"recipient1", "recipient2"},
	}

	res, err := cc.GetEmail(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedChannelConfig, res)
}

func TestGetEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(nil, nil)

	res, err := cc.GetEmail(ctx)
	require.ErrorIs(t, err, email.ErrNotFound)
	require.Nil(t, res)
}

func TestGetEmail_Failed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(nil, errors.New("error"))

	res, err := cc.GetEmail(ctx)
	require.ErrorIs(t, err, email.ErrClient)
	require.Nil(t, res)
}

func TestCreateEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	cfg := email.ChannelConfig{Username: "user"}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(nil, nil)
	mockSinksClient.On("CreateSink", ctx, EmailChannelName(), mock.Anything).Return(nil)

	err := cc.CreateEmail(ctx, cfg)
	require.NoError(t, err)
}

func TestCreateEmail_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	cfg := email.ChannelConfig{Username: "user"}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(&types.SinkConfig{}, nil)

	err := cc.CreateEmail(ctx, cfg)
	require.ErrorIs(t, err, email.ErrAlreadyExists)
}

func TestCreateEmail_GetFailed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	cfg := email.ChannelConfig{Username: "user"}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(nil, errors.New("error"))

	err := cc.CreateEmail(ctx, cfg)
	require.ErrorIs(t, err, email.ErrClient)
}

func TestCreateEmail_UpdateFailed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	cfg := email.ChannelConfig{Username: "user"}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(nil, nil)
	mockSinksClient.On("CreateSink", ctx, EmailChannelName(), mock.Anything).Return(errors.New("error"))

	err := cc.CreateEmail(ctx, cfg)
	require.ErrorIs(t, err, email.ErrClient)
}

func TestUpdateEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	cfg := email.ChannelConfig{Username: "user"}

	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(nil)
	mockSinksClient.On("CreateSink", ctx, EmailChannelName(), mock.Anything).Return(nil)

	err := cc.UpdateEmail(ctx, cfg)
	require.NoError(t, err)
}

func TestUpdateEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	cfg := email.ChannelConfig{Username: "user"}

	notFoundErr := &proxyapi.Error{Code: http.StatusNotFound}
	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(notFoundErr)

	err := cc.UpdateEmail(ctx, cfg)
	require.ErrorIs(t, err, email.ErrNotFound)
}

func TestUpdateEmail_DeleteFailed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	cfg := email.ChannelConfig{Username: "user"}

	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(errors.New("error"))

	err := cc.UpdateEmail(ctx, cfg)
	require.ErrorIs(t, err, email.ErrClient)
}

func TestUpdateEmail_UpdateFailed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	cfg := email.ChannelConfig{Username: "user"}

	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(nil)
	mockSinksClient.On("CreateSink", ctx, EmailChannelName(), mock.Anything).Return(errors.New("error"))

	err := cc.UpdateEmail(ctx, cfg)
	require.ErrorIs(t, err, email.ErrClient)
}

func TestDeleteEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(nil)

	err := cc.DeleteEmail(ctx)
	require.NoError(t, err)
}

func TestDeleteEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	notFoundErr := &proxyapi.Error{Code: http.StatusNotFound}
	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(notFoundErr)

	err := cc.DeleteEmail(ctx)
	require.ErrorIs(t, err, email.ErrNotFound)
}

func TestDeleteEmail_Failed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	mockSinksClient.On("DeleteSink", ctx, EmailChannelName()).Return(errors.New("error"))

	err := cc.DeleteEmail(ctx)
	require.ErrorIs(t, err, email.ErrClient)
}

func TestValidateEmail_Valid(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: sc}
	cfg := email.ChannelConfig{SmtpHost: "smtp.example.com"}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(&clients.SinkValidityResponse{}, nil)

	resp, err := cc.ValidateEmail(ctx, cfg)
	require.NoError(t, err)
	require.True(t, resp.Status)
}

func TestValidateEmail_Invalid(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: sc}
	cfg := email.ChannelConfig{}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(&clients.SinkValidityResponse{ErrorCode: ptr.To(int32(4)), ErrorMessage: "bad credentials"}, nil)

	resp, err := cc.ValidateEmail(ctx, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, email.ErrValidationFailed)
	require.Contains(t, err.Error(), "bad credentials")
	require.False(t, resp.Status)
	require.Equal(t, "bad credentials", resp.Message)
}

func TestValidateEmail_ClientFails(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: sc}
	cfg := email.ChannelConfig{Username: "user"}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(nil, fmt.Errorf("connection error"))

	resp, err := cc.ValidateEmail(ctx, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, email.ErrClient)
	require.False(t, resp.Status)
}

func TestValidateEmail_ParseError(t *testing.T) {
	sc := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: sc}
	cfg := email.ChannelConfig{}

	sc.On("ValidateSink", ctx, "sink-validation", mock.Anything).Return(&clients.SinkValidityResponse{ErrorCode: ptr.To(int32(1))}, nil).Once()

	resp, err := cc.ValidateEmail(ctx, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, email.ErrValidationFailed)
	require.False(t, resp.Status)
}

// ------------------------------------------- END Test Validate Email --------------------------

// ------------------------------------------- START Test Patch Email --------------------------

func TestPatchEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	existingConfig := &types.SinkConfig{
		Id:   EmailChannelName(),
		Type: "email",
		Config: types.SinkConfigDetails{
			EmailSinkConfig: types.EmailSinkConfig{
				User:            "old-user",
				Password:        "old-password",
				SmtpHost:        "old-smtp-host",
				SmtpPort:        587,
				From:            "old-from",
				AuthType:        "old-auth-type",
				SmtpTlsEnabled:  ptr.To(true),
				FromDisplayName: "old-from-display-name",
				UseAttachments:  ptr.To(true),
				Subject:         "old-subject",
			},
			Recipients:          []string{"old-recipient1", "old-recipient2"},
			DirectNotifications: ptr.To(true),
		},
	}

	patchConfig := email.ChannelConfig{
		Username: "new-user",
		Password: "new-password",
		Subject:  "new-subject",
	}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(existingConfig, nil)
	mockSinksClient.On("UpdateSink", ctx, EmailChannelName(), mock.MatchedBy(func(cfg *types.SinkConfig) bool {
		return cfg.Config.EmailSinkConfig.User == "new-user"
	})).Return(nil)

	expectedConfig := &email.ChannelConfig{
		Username:            "new-user",
		Password:            "new-password",
		SmtpHost:            "old-smtp-host",
		SmtpPort:            587,
		From:                "old-from",
		AuthType:            "old-auth-type",
		SmtpTlsEnabled:      ptr.To(true),
		FromDisplayName:     "old-from-display-name",
		UseAttachments:      ptr.To(true),
		Subject:             "new-subject",
		Recipients:          []string{"old-recipient1", "old-recipient2"},
		DirectNotifications: ptr.To(true),
	}

	res, err := cc.PatchEmail(ctx, patchConfig)
	require.NoError(t, err)
	require.Equal(t, expectedConfig, res)
}

func TestPatchEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(nil, nil)

	res, err := cc.PatchEmail(ctx, email.ChannelConfig{})
	require.ErrorIs(t, err, email.ErrNotFound)
	require.Nil(t, res)
}

func TestPatchEmail_GetFailed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(nil, errors.New("error"))

	res, err := cc.PatchEmail(ctx, email.ChannelConfig{})
	require.ErrorIs(t, err, email.ErrClient)
	require.Nil(t, res)
}

func TestPatchEmail_UpdateFailed(t *testing.T) {
	ctx := context.Background()
	mockSinksClient := mocks.NewSinksClienter(t)
	cc := &ChannelClient{sinksClient: mockSinksClient}

	existingConfig := &types.SinkConfig{
		Id:   EmailChannelName(),
		Type: "email",
		Config: types.SinkConfigDetails{
			EmailSinkConfig: types.EmailSinkConfig{
				User: "old-user",
			},
		},
	}

	mockSinksClient.On("GetSink", ctx, EmailChannelName()).Return(existingConfig, nil)
	mockSinksClient.On("UpdateSink", ctx, EmailChannelName(), mock.Anything).Return(errors.New("error"))

	res, err := cc.PatchEmail(ctx, email.ChannelConfig{})
	require.ErrorIs(t, err, email.ErrClient)
	require.Nil(t, res)
}

func TestMergeEmailConfigs(t *testing.T) {
	current := types.SinkConfigDetails{EmailSinkConfig: types.EmailSinkConfig{
		User:     "old-user",
		Password: "old-password",
		SmtpPort: 587,
	}}

	patch := types.SinkConfigDetails{EmailSinkConfig: types.EmailSinkConfig{
		User:     "new-user",
		SmtpPort: 465,
	}}

	res := merge(current, patch)

	require.Equal(t, "new-user", res.User)
	require.Equal(t, "old-password", res.Password)
	require.Equal(t, 465, res.SmtpPort)
}

// ------------------------------------------- END Test Patch Email --------------------------
