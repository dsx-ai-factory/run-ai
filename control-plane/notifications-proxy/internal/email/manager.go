/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package email

import (
	"context"
)

type EmailManager struct {
	sinksClient SinkClienter
}

func NewEmailManager(sinksClient SinkClienter) *EmailManager {
	return &EmailManager{
		sinksClient: sinksClient,
	}
}

func (em *EmailManager) GetEmail(ctx context.Context) (*ChannelConfig, error) {
	return em.sinksClient.GetEmail(ctx)
}

func (em *EmailManager) CreateEmail(ctx context.Context, channelConfig *ChannelConfig) (*ChannelConfig, error) {
	err := em.sinksClient.CreateEmail(ctx, *channelConfig)
	if err != nil {
		return nil, err
	}

	return channelConfig, nil
}

func (em *EmailManager) UpdateEmail(ctx context.Context, channelConfig *ChannelConfig) (*ChannelConfig, error) {
	err := em.sinksClient.UpdateEmail(ctx, *channelConfig)
	if err != nil {
		return nil, err
	}

	return channelConfig, nil
}

func (em *EmailManager) DeleteEmail(ctx context.Context) error {
	return em.sinksClient.DeleteEmail(ctx)
}

func (em *EmailManager) PatchEmail(ctx context.Context, channelConfig *ChannelConfig) (*ChannelConfig, error) {
	return em.sinksClient.PatchEmail(ctx, *channelConfig)
}

func (em *EmailManager) ValidateEmail(ctx context.Context, channelConfig *ChannelConfig) (ValidationResponse, error) {
	return em.sinksClient.ValidateEmail(ctx, *channelConfig)
}
