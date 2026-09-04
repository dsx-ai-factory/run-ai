/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package state

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/channel"
)

// EmailState represents email notification state and implements notification_state.Stater
type EmailState struct {
	Enabled *bool
}

func (ss *EmailState) IsEnabled() bool {
	if ss.Enabled == nil {
		return false
	}
	return *ss.Enabled
}

func (ss *EmailState) GetType() string {
	return channel.NotificationChannelTypeEmail
}

func (ss *EmailState) GetChanName() string {
	return channel.EmailChannelName()
}

func (ss *EmailState) GetMetadata() any {
	return nil
}

// SlackMetadata holds typed metadata for SlackState
type SlackMetadata struct {
	WorkspaceName            *string
	AppCredentialsConfigured *bool
}

// SlackState represents slack notification state and implements notification_state.Stater
type SlackState struct {
	TenantId                 string
	Enabled                  *bool
	WorkspaceName            *string
	AppCredentialsConfigured *bool
}

func (ss *SlackState) IsEnabled() bool {
	if ss.Enabled == nil {
		return false
	}
	return *ss.Enabled
}

func (ss *SlackState) GetType() string {
	return channel.NotificationChannelTypeSlack
}

func (ss *SlackState) GetChanName() string {
	return channel.SlackChannelName(ss.TenantId)
}

func (ss *SlackState) GetMetadata() any {
	return &SlackMetadata{
		WorkspaceName:            ss.WorkspaceName,
		AppCredentialsConfigured: ss.AppCredentialsConfigured,
	}
}
