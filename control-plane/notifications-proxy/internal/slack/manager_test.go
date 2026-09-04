/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package slack_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/slack"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/slack/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestNewSlackManager(t *testing.T) {
	notificationStateClient := mocks.NewNotificationStateClienter(t)
	channelClient := mocks.NewChannelClienter(t)
	slackClient := mocks.NewSlackClienter(t)
	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	notifier := mocks.NewNotifier(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	userSubscription := mocks.NewUserSubscriptionsClienter(t)

	manager := slack.NewSlackManager(notificationStateClient, channelClient, slackClient, authSessionHandler, notifier, slackCredentialsDao, userSubscription)
	require.NotNil(t, manager)
}

func TestAuthorizeApp(t *testing.T) {
	ctx := context.Background()

	stateClient := mocks.NewNotificationStateClienter(t)
	channelClient := mocks.NewChannelClienter(t)
	slackClient := mocks.NewSlackClienter(t)
	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	notifier := mocks.NewNotifier(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	userSubscription := mocks.NewUserSubscriptionsClienter(t)
	manager := slack.NewSlackManager(stateClient, channelClient, slackClient, authSessionHandler, notifier, slackCredentialsDao, userSubscription)

	slackCredentialsDao.On("Get", ctx).Return(&slack.SlackCredentials{
		ClientId:     "client_id",
		ClientSecret: "client_secret",
	}, nil)

	authSessionHandler.On("Parse", ctx, "state_key").Return(&slack.SlackAuthSessionState{
		TenantId: "1",
		Email:    "test@run.ai",
	}, nil)

	slackClient.On("ExchangeCode", "code", "client_id", "client_secret").Return(&slack.ExchangeCodeResult{
		AccessToken: "token",
		TeamName:    "workspace",
	}, nil)

	channelClient.On("CreateSlack", ctx, "1", mock.MatchedBy(func(ch slack.ChannelConfig) bool {
		return *ch.DirectNotifications == true &&
			ch.ApiToken == "token"
	})).Return(&slack.ChannelConfig{Id: "slack-1"}, nil)

	stateClient.On("UpdateSlackNotificationState", ctx, slack.NotificationState{
		TenantId:      "1",
		WorkspaceName: ptr.To("workspace"),
		Enabled:       ptr.To(true),
	}).Return(nil)

	notifier.On("Welcome", ctx, "1", "test@run.ai", "slack-1").Return(nil)

	err := manager.AuthorizeApp(ctx, "code", "state_key")
	require.NoError(t, err)
}

func TestAuthorizeApp_MissingCredentials(t *testing.T) {
	ctx := context.Background()

	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(nil, nil, nil, nil, nil, slackCredentialsDao, nil)

	slackCredentialsDao.On("Get", ctx).Return(nil, slack.ErrMissingAppCredentials)

	err := manager.AuthorizeApp(ctx, "code", "state_key")
	require.ErrorIs(t, err, slack.ErrMissingAppCredentials)
}

func TestAuthorizeApp_ParseStateError(t *testing.T) {
	ctx := context.Background()

	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(nil, nil, nil, authSessionHandler, nil, slackCredentialsDao, nil)

	slackCredentialsDao.On("Get", ctx).Return(&slack.SlackCredentials{
		ClientId:     "client_id",
		ClientSecret: "client_secret",
	}, nil)
	authSessionHandler.On("Parse", ctx, "state_key").Return(nil, errors.New("invalid state"))

	err := manager.AuthorizeApp(ctx, "code", "state_key")
	require.EqualError(t, err, "failed to parse slack session state: invalid state")
}

func TestAuthorizeApp_ExchangeCodeError(t *testing.T) {
	ctx := context.Background()

	slackClient := mocks.NewSlackClienter(t)
	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(nil, nil, slackClient, authSessionHandler, nil, slackCredentialsDao, nil)

	slackCredentialsDao.On("Get", ctx).Return(
		&slack.SlackCredentials{
			ClientId:     "client_id",
			ClientSecret: "client_secret",
		}, nil)
	authSessionHandler.On("Parse", ctx, "state_key").Return(&slack.SlackAuthSessionState{
		TenantId: "1",
		Email:    "test@run.ai",
	}, nil)
	slackClient.On("ExchangeCode", "code", "client_id", "client_secret").Return(nil, errors.New("exchange failed"))

	err := manager.AuthorizeApp(ctx, "code", "state_key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to exchange slack code")
}

func TestAuthorizeApp_CreateChannelError(t *testing.T) {
	ctx := context.Background()

	slackClient := mocks.NewSlackClienter(t)
	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	channelClient := mocks.NewChannelClienter(t)
	manager := slack.NewSlackManager(nil, channelClient, slackClient, authSessionHandler, nil, slackCredentialsDao, nil)

	slackCredentialsDao.On("Get", ctx).Return(
		&slack.SlackCredentials{
			ClientId:     "client_id",
			ClientSecret: "client_secret",
		}, nil)
	authSessionHandler.On("Parse", ctx, "state_key").Return(&slack.SlackAuthSessionState{
		TenantId: "1",
		Email:    "test@run.ai",
	}, nil)
	slackClient.On("ExchangeCode", "code", "client_id", "client_secret").Return(&slack.ExchangeCodeResult{
		AccessToken: "token",
		TeamName:    "workspace",
	}, nil)

	channelClient.On("CreateSlack", ctx, "1", mock.Anything).Return(nil, errors.New("channel creation failed"))

	err := manager.AuthorizeApp(ctx, "code", "state_key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create slack channel: channel creation failed")
}

func TestAuthorizeApp_UpdateStateError(t *testing.T) {
	ctx := context.Background()

	slackClient := mocks.NewSlackClienter(t)
	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	channelClient := mocks.NewChannelClienter(t)
	notificationStateClient := mocks.NewNotificationStateClienter(t)
	manager := slack.NewSlackManager(notificationStateClient, channelClient, slackClient, authSessionHandler, nil, slackCredentialsDao, nil)

	slackCredentialsDao.On("Get", ctx).Return(
		&slack.SlackCredentials{
			ClientId:     "client_id",
			ClientSecret: "client_secret",
		}, nil)
	authSessionHandler.On("Parse", ctx, "state_key").Return(&slack.SlackAuthSessionState{
		TenantId: "1",
		Email:    "test@run.ai",
	}, nil)
	slackClient.On("ExchangeCode", "code", "client_id", "client_secret").Return(&slack.ExchangeCodeResult{
		AccessToken: "token",
		TeamName:    "workspace",
	}, nil)

	channelClient.On("CreateSlack", ctx, "1", mock.Anything).Return(&slack.ChannelConfig{Id: "slack-1"}, nil)

	notificationStateClient.On("UpdateSlackNotificationState", ctx, slack.NotificationState{
		TenantId:      "1",
		WorkspaceName: ptr.To("workspace"),
		Enabled:       ptr.To(true),
	}).Return(errors.New("state update failed"))

	err := manager.AuthorizeApp(ctx, "code", "state_key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to update slack notification state: state update failed")
}

func TestAuthorizeApp_WelcomeNotifierError_DoesNotFail(t *testing.T) {
	ctx := context.Background()

	slackClient := mocks.NewSlackClienter(t)
	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	channelClient := mocks.NewChannelClienter(t)
	notificationStateClient := mocks.NewNotificationStateClienter(t)
	notifier := mocks.NewNotifier(t)
	manager := slack.NewSlackManager(notificationStateClient, channelClient, slackClient, authSessionHandler, notifier, slackCredentialsDao, nil)

	slackCredentialsDao.On("Get", ctx).Return(
		&slack.SlackCredentials{
			ClientId:     "client_id",
			ClientSecret: "client_secret",
		}, nil)
	authSessionHandler.On("Parse", ctx, "state_key").Return(&slack.SlackAuthSessionState{
		TenantId: "1",
		Email:    "test@run.ai",
	}, nil)
	slackClient.On("ExchangeCode", "code", "client_id", "client_secret").Return(&slack.ExchangeCodeResult{
		AccessToken: "token",
		TeamName:    "workspace",
	}, nil)

	channelClient.On("CreateSlack", ctx, "1", mock.Anything).Return(&slack.ChannelConfig{Id: "slack-1"}, nil)

	notificationStateClient.On("UpdateSlackNotificationState", ctx, mock.MatchedBy(func(s slack.NotificationState) bool {
		return s.TenantId == "1" && *s.Enabled == true
	})).Return(nil)

	notifier.On("Welcome", ctx, "1", "test@run.ai", "slack-1").Return(errors.New("notification failed"))

	err := manager.AuthorizeApp(ctx, "code", "state_key")
	require.NoError(t, err)
}

func TestCreateApp(t *testing.T) {
	ctx := context.Background()

	slackClient := mocks.NewSlackClienter(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(nil, nil, slackClient, nil, nil, slackCredentialsDao, nil)

	slackClient.On("CreateApp", "xoxe.xoxp-1-token").Return(&slack.SlackCredentials{
		AppId:        "A123",
		ClientId:     "client_id",
		ClientSecret: "client_secret",
	}, nil)
	slackCredentialsDao.On("Save", ctx, &slack.SlackCredentials{
		AppId:        "A123",
		ClientId:     "client_id",
		ClientSecret: "client_secret",
	}).Return(nil)

	url, err := manager.CreateApp(ctx, "xoxe.xoxp-1-token")
	require.NoError(t, err)
	require.Equal(t, "https://api.slack.com/apps/A123", url)
}

func TestCreateApp_FailedToCreateApp(t *testing.T) {
	ctx := context.Background()

	slackClient := mocks.NewSlackClienter(t)
	manager := slack.NewSlackManager(nil, nil, slackClient, nil, nil, nil, nil)

	slackClient.On("CreateApp", "xoxe.xoxp-1-token").Return(nil, errors.New("create app failed"))

	url, err := manager.CreateApp(ctx, "xoxe.xoxp-1-token")
	require.EqualError(t, err, "failed to create slack app: create app failed")
	require.Empty(t, url)
}

func TestCreateApp_FailedToSaveCredentials(t *testing.T) {
	ctx := context.Background()

	slackClient := mocks.NewSlackClienter(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(nil, nil, slackClient, nil, nil, slackCredentialsDao, nil)

	slackClient.On("CreateApp", "xoxe.xoxp-1-token").Return(&slack.SlackCredentials{
		AppId:        "A123",
		ClientId:     "client_id",
		ClientSecret: "client_secret",
	}, nil)
	slackCredentialsDao.On("Save", ctx, &slack.SlackCredentials{
		AppId:        "A123",
		ClientId:     "client_id",
		ClientSecret: "client_secret",
	}).Return(errors.New("save credentials failed"))

	url, err := manager.CreateApp(ctx, "xoxe.xoxp-1-token")
	require.EqualError(t, err, "failed to save slack credentials: save credentials failed")
	require.Empty(t, url)
}

func TestConnectWorkspace(t *testing.T) {
	ctx := context.Background()

	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	slackClient := mocks.NewSlackClienter(t)
	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	manager := slack.NewSlackManager(nil, nil, slackClient, authSessionHandler, nil, slackCredentialsDao, nil)

	authSessionHandler.On("Create", ctx, "1", "test@run.ai").Return("state_key", nil)
	slackCredentialsDao.On("Get", ctx).Return(&slack.SlackCredentials{
		ClientId: "client_id",
	}, nil)
	slackClient.On("GetAuthUrl", "state_key", "client_id").Return("https://slack.com/oauth/v2/authorize?client_id=123")

	url, err := manager.ConnectWorkspace(ctx, "1", "test@run.ai")
	require.NoError(t, err)
	require.Equal(t, "https://slack.com/oauth/v2/authorize?client_id=123", url)
}

func TestConnectWorkspace_MissingCredentials(t *testing.T) {
	ctx := context.Background()

	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(nil, nil, nil, nil, nil, slackCredentialsDao, nil)

	slackCredentialsDao.On("Get", ctx).Return(nil, slack.ErrMissingAppCredentials)

	url, err := manager.ConnectWorkspace(ctx, "1", "test@run.ai")
	require.ErrorIs(t, err, slack.ErrMissingAppCredentials)
	require.Empty(t, url)
}

func TestConnectWorkspace_CreateStateError(t *testing.T) {
	ctx := context.Background()

	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	authSessionHandler := mocks.NewSlackAuthSessionHandlerer(t)
	manager := slack.NewSlackManager(nil, nil, nil, authSessionHandler, nil, slackCredentialsDao, nil)

	authSessionHandler.On("Create", ctx, "1", "test@run.ai").Return("", errors.New("create state failed"))
	slackCredentialsDao.On("Get", ctx).Return(&slack.SlackCredentials{
		ClientId: "client_id",
	}, nil)

	url, err := manager.ConnectWorkspace(ctx, "1", "test@run.ai")
	require.EqualError(t, err, "failed to create slack session: create state failed")
	require.Empty(t, url)
}

func TestDeleteApp(t *testing.T) {
	ctx := context.Background()

	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(nil, nil, nil, nil, nil, slackCredentialsDao, nil)

	slackCredentialsDao.On("Delete", ctx).Return(nil)
	err := manager.DeleteApp(ctx)
	require.NoError(t, err)
}

func TestDeleteApp_FailedToDeleteCredentials(t *testing.T) {
	ctx := context.Background()

	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(nil, nil, nil, nil, nil, slackCredentialsDao, nil)

	slackCredentialsDao.On("Delete", ctx).Return(errors.New("delete credentials failed"))
	err := manager.DeleteApp(ctx)
	require.EqualError(t, err, "failed to delete credentials from dao: delete credentials failed")
}

func TestDisconnectWorkspace(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	userSubscription := mocks.NewUserSubscriptionsClienter(t)
	notificationStateClient := mocks.NewNotificationStateClienter(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(notificationStateClient, channelClient, nil, nil, nil, slackCredentialsDao, userSubscription)

	channelClient.On("DeleteSlack", ctx, "1").Return(&slack.ChannelConfig{Id: "slack-1"}, nil)
	userSubscription.On("DisableChannel", ctx, "1", "slack-1").Return(nil)
	notificationStateClient.On("UpdateSlackNotificationState", ctx, slack.NotificationState{
		TenantId:      "1",
		Enabled:       ptr.To(false),
		WorkspaceName: ptr.To(""),
	}).Return(nil)
	err := manager.DisconnectWorkspace(ctx, "1")
	require.NoError(t, err)
}

func TestDisconnectWorkspace_FailedToDeleteSlackChannelConfig(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	userSubscription := mocks.NewUserSubscriptionsClienter(t)
	notificationStateClient := mocks.NewNotificationStateClienter(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(notificationStateClient, channelClient, nil, nil, nil, slackCredentialsDao, userSubscription)

	channelClient.On("DeleteSlack", ctx, "1").Return(nil, errors.New("delete slack failed"))
	err := manager.DisconnectWorkspace(ctx, "1")
	require.EqualError(t, err, "delete slack failed")
}

func TestDisconnectWorkspace_FailedToDisableSlackChannel(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	userSubscription := mocks.NewUserSubscriptionsClienter(t)
	notificationStateClient := mocks.NewNotificationStateClienter(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(notificationStateClient, channelClient, nil, nil, nil, slackCredentialsDao, userSubscription)

	channelClient.On("DeleteSlack", ctx, "1").Return(&slack.ChannelConfig{Id: "slack-1"}, nil)
	userSubscription.On("DisableChannel", ctx, "1", "slack-1").Return(errors.New("disable slack channel failed"))
	err := manager.DisconnectWorkspace(ctx, "1")
	require.ErrorContains(t, err, "failed to disable slack chanel from user subscriptions")
	require.ErrorContains(t, err, "disable slack channel failed")
}

func TestDisconnectWorkspace_FailedToUpdateSlackNotificationState(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	userSubscription := mocks.NewUserSubscriptionsClienter(t)
	notificationStateClient := mocks.NewNotificationStateClienter(t)
	slackCredentialsDao := mocks.NewSlackCredentialsDaoer(t)
	manager := slack.NewSlackManager(notificationStateClient, channelClient, nil, nil, nil, slackCredentialsDao, userSubscription)

	channelClient.On("DeleteSlack", ctx, "1").Return(&slack.ChannelConfig{Id: "slack-1"}, nil)
	userSubscription.On("DisableChannel", ctx, "1", "slack-1").Return(nil)
	notificationStateClient.On("UpdateSlackNotificationState", ctx, slack.NotificationState{
		TenantId:      "1",
		WorkspaceName: ptr.To(""),
		Enabled:       ptr.To(false),
	}).Return(errors.New("update slack notification state failed"))
	err := manager.DisconnectWorkspace(ctx, "1")
	require.ErrorContains(t, err, "failed to update tenant subscription")
	require.ErrorContains(t, err, "update slack notification state failed")
}

func TestGetSlackConfig(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	manager := slack.NewSlackManager(nil, channelClient, nil, nil, nil, nil, nil)

	channelClient.On("GetSlack", ctx, "1").Return(&slack.ChannelConfig{Id: "slack-1"}, nil)
	channelConfig, err := manager.GetSlackConfig(ctx, "1")
	require.NoError(t, err)
	require.Equal(t, "slack-1", channelConfig.Id)
}

func TestGetSlackConfig_Failed(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	manager := slack.NewSlackManager(nil, channelClient, nil, nil, nil, nil, nil)

	channelClient.On("GetSlack", ctx, "1").Return(nil, errors.New("get slack config failed"))
	_, err := manager.GetSlackConfig(ctx, "1")
	require.EqualError(t, err, "get slack config failed")
}

func TestPatchSlackConfig(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	manager := slack.NewSlackManager(nil, channelClient, nil, nil, nil, nil, nil)

	channelClient.On("PatchSlack", ctx, "1", mock.MatchedBy(func(ch slack.ChannelConfig) bool {
		return *ch.DirectNotifications == true &&
			ch.ApiToken == "token"
	})).Return(&slack.ChannelConfig{Id: "slack-1"}, nil)

	channelConfig, err := manager.PatchSlackConfig(ctx, "1", slack.ChannelConfig{
		DirectNotifications: ptr.To(true),
		ApiToken:            "token",
	})
	require.NoError(t, err)
	require.Equal(t, "slack-1", channelConfig.Id)
}

func TestPatchSlackConfig_Failed(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	manager := slack.NewSlackManager(nil, channelClient, nil, nil, nil, nil, nil)

	channelClient.On("PatchSlack", ctx, "1", mock.MatchedBy(func(ch slack.ChannelConfig) bool {
		return *ch.DirectNotifications == true &&
			ch.ApiToken == "token"
	})).Return(nil, errors.New("patch slack config failed"))

	_, err := manager.PatchSlackConfig(ctx, "1", slack.ChannelConfig{
		DirectNotifications: ptr.To(true),
		ApiToken:            "token",
	})
	require.EqualError(t, err, "patch slack config failed")
}

func TestValidateSlack(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	manager := slack.NewSlackManager(nil, channelClient, nil, nil, nil, nil, nil)

	cfg := slack.ChannelConfig{
		ApiToken:            "xoxb-token",
		DirectNotifications: ptr.To(true),
	}
	channelClient.On("ValidateSlack", ctx, mock.MatchedBy(func(ch slack.ChannelConfig) bool {
		return ch.ApiToken == "xoxb-token" && *ch.DirectNotifications == true
	})).Return(slack.ValidationResponse{Status: true}, nil)

	resp, err := manager.ValidateSlack(ctx, cfg)
	require.NoError(t, err)
	require.True(t, resp.Status)
}

func TestValidateSlack_Failed(t *testing.T) {
	ctx := context.Background()
	channelClient := mocks.NewChannelClienter(t)
	manager := slack.NewSlackManager(nil, channelClient, nil, nil, nil, nil, nil)

	cfg := slack.ChannelConfig{ApiToken: "xoxb-token"}
	channelClient.On("ValidateSlack", ctx, mock.Anything).Return(slack.ValidationResponse{Status: false}, errors.New("validate failed"))

	resp, err := manager.ValidateSlack(ctx, cfg)
	require.Error(t, err)
	require.EqualError(t, err, "validate failed")
	require.False(t, resp.Status)
}
