/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package email_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/email"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/email/mocks"
	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestGetEmail_Succeed(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	expected := &email.ChannelConfig{Username: "user", SmtpHost: "smtp.example.com"}

	sc.On("GetEmail", ctx).Return(expected, nil)

	res, err := manager.GetEmail(ctx)
	require.NoError(t, err)
	require.Equal(t, expected, res)
}

func TestGetEmail_ClientFails(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)

	sc.On("GetEmail", ctx).Return(nil, email.ErrClient)

	res, err := manager.GetEmail(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, email.ErrClient)
	require.Nil(t, res)
}

func TestCreateEmail_Succeed(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	cfg := &email.ChannelConfig{Username: "user"}

	sc.On("CreateEmail", ctx, *cfg).Return(nil)

	res, err := manager.CreateEmail(ctx, cfg)
	require.NoError(t, err)
	require.Equal(t, cfg, res)
}

func TestCreateEmail_ClientFails(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	cfg := &email.ChannelConfig{Username: "user"}

	sc.On("CreateEmail", ctx, *cfg).Return(email.ErrAlreadyExists)

	res, err := manager.CreateEmail(ctx, cfg)
	require.ErrorIs(t, err, email.ErrAlreadyExists)
	require.Nil(t, res)
}

func TestUpdateEmail_Succeed(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	cfg := &email.ChannelConfig{Username: "user", SmtpHost: "smtp.example.com"}

	sc.On("UpdateEmail", ctx, *cfg).Return(nil)

	res, err := manager.UpdateEmail(ctx, cfg)
	require.NoError(t, err)
	require.Equal(t, cfg, res)
}

func TestUpdateEmail_ClientFails(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	cfg := &email.ChannelConfig{Username: "user"}

	sc.On("UpdateEmail", ctx, *cfg).Return(email.ErrClient)

	res, err := manager.UpdateEmail(ctx, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, email.ErrClient)
	require.Nil(t, res)
}

func TestDeleteEmail_Succeed(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)

	sc.On("DeleteEmail", ctx).Return(nil)

	err := manager.DeleteEmail(ctx)
	require.NoError(t, err)
}

func TestDeleteEmail_ClientFails(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)

	sc.On("DeleteEmail", ctx).Return(fmt.Errorf("delete failed"))

	err := manager.DeleteEmail(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "delete failed")
}

func TestPatchEmail_Succeed(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	cfg := &email.ChannelConfig{Username: "new-user"}
	merged := &email.ChannelConfig{Username: "new-user", SmtpHost: "old-host"}

	sc.On("PatchEmail", ctx, *cfg).Return(merged, nil)

	res, err := manager.PatchEmail(ctx, cfg)
	require.NoError(t, err)
	require.Equal(t, merged, res)
}

func TestPatchEmail_ClientFails(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	cfg := &email.ChannelConfig{Username: "user"}

	sc.On("PatchEmail", ctx, *cfg).Return(nil, email.ErrNotFound)

	res, err := manager.PatchEmail(ctx, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, email.ErrNotFound)
	require.Nil(t, res)
}

func TestValidateEmail(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	cfg := &email.ChannelConfig{Username: "user", SmtpHost: "smtp.example.com"}

	sc.On("ValidateEmail", ctx, *cfg).Return(email.ValidationResponse{Status: true}, nil)

	resp, err := manager.ValidateEmail(ctx, cfg)
	require.NoError(t, err)
	require.True(t, resp.Status)
}

func TestValidateEmail_Failed(t *testing.T) {
	sc := mocks.NewSinkClienter(t)
	manager := email.NewEmailManager(sc)
	cfg := &email.ChannelConfig{Username: "user"}

	sc.On("ValidateEmail", ctx, *cfg).Return(email.ValidationResponse{Status: false}, email.ErrClient)

	resp, err := manager.ValidateEmail(ctx, cfg)
	require.Error(t, err)
	require.ErrorIs(t, err, email.ErrClient)
	require.False(t, resp.Status)
}
