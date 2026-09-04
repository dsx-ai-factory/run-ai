/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package slack

import (
	"context"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

type SlackManager struct {
	channelClient           ChannelClienter
	notificationStateClient NotificationStateClienter
	slackClient             SlackClienter
	slackAuthSession        SlackAuthSessionHandlerer
	notifier                Notifier
	slackCredentialsDao     SlackCredentialsDaoer
	userSubscriptions       UserSubscriptionsClienter
}

func NewSlackManager(
	notificationStateClient NotificationStateClienter,
	channelClient ChannelClienter,
	slackClient SlackClienter,
	slackAuthSession SlackAuthSessionHandlerer,
	notifier Notifier,
	slackCredentialsDao SlackCredentialsDaoer,
	userSubscriptions UserSubscriptionsClienter) *SlackManager {
	return &SlackManager{
		notificationStateClient: notificationStateClient,
		channelClient:           channelClient,
		slackClient:             slackClient,
		slackAuthSession:        slackAuthSession,
		notifier:                notifier,
		slackCredentialsDao:     slackCredentialsDao,
		userSubscriptions:       userSubscriptions,
	}
}

func (sm *SlackManager) AuthorizeApp(ctx context.Context, code string, sessionId string) error {
	creds, err := sm.slackCredentialsDao.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get slack credentials: %w", err)
	}

	parsedState, err := sm.slackAuthSession.Parse(ctx, sessionId)
	if err != nil {
		return fmt.Errorf("failed to parse slack session state: %w", err)
	}

	exchangeRes, err := sm.slackClient.ExchangeCode(code, creds.ClientId, creds.ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to exchange slack code: %w", err)
	}

	channel, err := sm.createOrPatchSlackChannel(ctx, parsedState, exchangeRes)
	if err != nil {
		return err
	}

	err = sm.notificationStateClient.UpdateSlackNotificationState(ctx, NotificationState{
		TenantId:      parsedState.TenantId,
		WorkspaceName: ptr.To(exchangeRes.TeamName),
		Enabled:       ptr.To(true),
	})
	if err != nil {
		return fmt.Errorf("failed to update slack notification state: %w", err)
	}

	if err := sm.notifier.Welcome(ctx, parsedState.TenantId, parsedState.Email, channel.Id); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to notify welcome message: %v", err)
	}

	return nil
}

func (sm *SlackManager) CreateApp(ctx context.Context, accessToken string) (string, error) {
	credentials, err := sm.slackClient.CreateApp(accessToken)
	if err != nil {
		return "", fmt.Errorf("failed to create slack app: %w", err)
	}
	err = sm.slackCredentialsDao.Save(ctx, &SlackCredentials{
		AppId:        credentials.AppId,
		ClientId:     credentials.ClientId,
		ClientSecret: credentials.ClientSecret,
	})
	if err != nil {
		return "", fmt.Errorf("failed to save slack credentials: %w", err)
	}

	return fmt.Sprintf("%s/%s", slackAppsPageURL, credentials.AppId), nil
}

func (sm *SlackManager) DeleteApp(ctx context.Context) error {
	if err := sm.slackCredentialsDao.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete credentials from dao: %w", err)
	}

	return nil
}

func (sm *SlackManager) ConnectWorkspace(ctx context.Context, tenantId string, subjectEmail string) (string, error) {
	creds, err := sm.slackCredentialsDao.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get slack credentials: %w", err)
	}

	sessionId, err := sm.slackAuthSession.Create(ctx, tenantId, subjectEmail)
	if err != nil {
		return "", fmt.Errorf("failed to create slack session: %w", err)
	}

	return sm.slackClient.GetAuthUrl(sessionId, creds.ClientId), nil
}

func (sm *SlackManager) SyncSecrets(ctx context.Context) error {
	slackConfig := config.Get[*SlackConfig]()
	if slackConfig.ClientId == "" || slackConfig.ClientSecret == "" {
		return nil
	}

	err := sm.slackCredentialsDao.Save(ctx, &SlackCredentials{
		ClientId:     slackConfig.ClientId,
		ClientSecret: slackConfig.ClientSecret,
	})
	if err != nil {
		return fmt.Errorf("failed to save credentials from config to dao: %w", err)
	}
	return nil
}

func (sm *SlackManager) DisconnectWorkspace(ctx context.Context, tenantId string) error {
	channelConfig, err := sm.channelClient.DeleteSlack(ctx, tenantId)
	if err != nil {
		return err
	}

	if err := sm.userSubscriptions.DisableChannel(ctx, tenantId, channelConfig.Id); err != nil {
		return errors.Join(fmt.Errorf("failed to disable slack chanel from user subscriptions"), err)
	}

	if err := sm.notificationStateClient.UpdateSlackNotificationState(ctx, NotificationState{
		TenantId:      tenantId,
		Enabled:       ptr.To(false),
		WorkspaceName: ptr.To(""),
	}); err != nil {
		return errors.Join(fmt.Errorf("failed to update tenant subscription: %w", err), ErrClientFailure)
	}

	return nil
}

func (sm *SlackManager) GetSlackConfig(ctx context.Context, tenantId string) (*ChannelConfig, error) {
	return sm.channelClient.GetSlack(ctx, tenantId)
}

func (sm *SlackManager) PatchSlackConfig(ctx context.Context, tenantId string, cfg ChannelConfig) (*ChannelConfig, error) {
	return sm.channelClient.PatchSlack(ctx, tenantId, cfg)
}

func (sm *SlackManager) ValidateSlack(ctx context.Context, cfg ChannelConfig) (ValidationResponse, error) {
	return sm.channelClient.ValidateSlack(ctx, cfg)
}

func (sm *SlackManager) createOrPatchSlackChannel(ctx context.Context, parsedState *SlackAuthSessionState, exchangeRes *ExchangeCodeResult) (*ChannelConfig, error) {
	channel, err := sm.channelClient.CreateSlack(ctx, parsedState.TenantId, ChannelConfig{
		DirectNotifications: ptr.To(true),
		ApiToken:            exchangeRes.AccessToken,
	})
	if err == nil {
		return channel, nil
	}

	if !errors.Is(err, ErrAlreadyExists) {
		return nil, fmt.Errorf("failed to create slack channel: %w", err)
	}

	channel, err = sm.channelClient.PatchSlack(ctx, parsedState.TenantId, ChannelConfig{
		DirectNotifications: ptr.To(true),
		ApiToken:            exchangeRes.AccessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to patch slack channel: %w", err)
	}

	return channel, nil
}
