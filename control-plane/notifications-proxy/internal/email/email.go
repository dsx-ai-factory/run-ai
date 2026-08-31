/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package email

import (
	"context"
	"fmt"
)

var (
	ErrClient           = fmt.Errorf("client error")
	ErrValidationFailed = fmt.Errorf("validation failed")
	ErrNotFound         = fmt.Errorf("not found")
	ErrAlreadyExists    = fmt.Errorf("already exists")
	ErrFailedToParse    = fmt.Errorf("failed to parse server response")
)

type SinkClienter interface {
	CreateEmail(ctx context.Context, config ChannelConfig) error
	GetEmail(ctx context.Context) (*ChannelConfig, error)
	DeleteEmail(ctx context.Context) error
	UpdateEmail(ctx context.Context, config ChannelConfig) error
	ValidateEmail(ctx context.Context, config ChannelConfig) (ValidationResponse, error)
	PatchEmail(ctx context.Context, config ChannelConfig) (*ChannelConfig, error)
}

type ChannelConfig struct {
	Username            string
	Password            string
	SmtpHost            string
	SmtpPort            int32
	From                string
	AuthType            string
	SmtpTlsEnabled      *bool
	Recipients          []string
	FromDisplayName     string
	DirectNotifications *bool
	UseAttachments      *bool
	Subject             string
}

type ValidationResponse struct {
	Status  bool
	Message string
}
