/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package users

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

type UserRepository interface {
	GetUsers(ctx context.Context, realmName string, userFilters models.UserFilters) ([]*models.User, error)
	CreateUser(ctx context.Context, realmName string, user *models.User) (string, error)
	GetUserByID(ctx context.Context, realmName string, userID string) (*models.User, error)
	GetUserByEmail(ctx context.Context, realmName string, email string) (*models.User, error)
	DeleteUser(ctx context.Context, realmName string, userID string) error
	LogoutUser(ctx context.Context, realm string, userID string) error
	ChangeUserPassword(ctx context.Context, realm string, user *models.User, currentPassword string, newPassword string) error
	ResetUserPassword(ctx context.Context, realm string, userID string, password string) error
	SendPasswordResetEmail(ctx context.Context, realm string, userID string, expirationSeconds int, redirectURI string, clientId string) error
	ToggleRequiredAction(ctx context.Context, realm string, actionAlias models.RequiredAction, enabled bool) error
	GetRequiredAction(ctx context.Context, realm string, actionAlias string) (*models.RequiredUserAction, error)
	ToggleUnmanagedUsersAttributes(ctx context.Context, realm string, enabled bool) error
	UpdateUser(ctx context.Context, realmName string, userID string, user *models.User) error
	LogoutUserBySessionID(ctx context.Context, realm string, sessionID string) error
}
