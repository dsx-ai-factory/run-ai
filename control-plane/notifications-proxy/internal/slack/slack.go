/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package slack

import (
	"context"
	"fmt"
)

const slackAppsPageURL = "https://api.slack.com/apps"

var (
	ErrMissingAppCredentials = fmt.Errorf("missing slack app credentials")
	ErrInvalidState          = fmt.Errorf("invalid slack state")
	ErrExpiredState          = fmt.Errorf("expired slack state")
	ErrNotFound              = fmt.Errorf("not found")
	ErrClientFailure         = fmt.Errorf("client failed")
	ErrAlreadyExists         = fmt.Errorf("already exists")
	ErrFailedToParse         = fmt.Errorf("failed to parse server response")
	ErrValidationFailed      = fmt.Errorf("validation failed")
)

type SlackClienter interface {
	CreateApp(accessToken string) (*SlackCredentials, error)
	ExchangeCode(code, clientId, clientSecret string) (*ExchangeCodeResult, error)
	GetAuthUrl(state, clientId string) string
}

type SlackAuthSessionHandlerer interface {
	Create(ctx context.Context, tenantId string, email string) (string, error)
	Parse(ctx context.Context, sessionId string) (*SlackAuthSessionState, error)
}

type Notifier interface {
	Welcome(ctx context.Context, tenantId string, email string, channelId string) error
}

type ChannelClienter interface {
	CreateSlack(ctx context.Context, tenantId string, channel ChannelConfig) (*ChannelConfig, error)
	DeleteSlack(ctx context.Context, tenantId string) (*ChannelConfig, error)
	GetSlack(ctx context.Context, tenantId string) (*ChannelConfig, error)
	PatchSlack(ctx context.Context, tenantId string, channel ChannelConfig) (*ChannelConfig, error)
	ValidateSlack(ctx context.Context, config ChannelConfig) (ValidationResponse, error)
}

type NotificationStateClienter interface {
	UpdateSlackNotificationState(ctx context.Context, slackState NotificationState) error
}
type UserSubscriptionsClienter interface {
	DisableChannel(ctx context.Context, tenantId string, channelId string) error
}
type SlackCredentialsDaoer interface {
	Save(ctx context.Context, slackCredential *SlackCredentials) error
	Get(ctx context.Context) (*SlackCredentials, error)
	Delete(ctx context.Context) error
}

type NotificationState struct {
	TenantId      string
	WorkspaceName *string
	Enabled       *bool
}

type ChannelConfig struct {
	Id                  string
	Type                string
	DirectNotifications *bool
	Recipients          []string
	TemplatesFolder     string
	TemplateFile        string
	ApiToken            string
}

type SlackCredentials struct {
	AppId        string
	ClientId     string
	ClientSecret string
}

type ExchangeCodeResult struct {
	AccessToken string
	TeamName    string
}

type SlackAuthSessionState struct {
	TenantId  string `json:"tenantId"`
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
}
