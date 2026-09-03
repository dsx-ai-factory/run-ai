/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

type AuthRepository interface {
	GetUserInfo(ctx context.Context, realm string, accessToken string) (*models.UserInfo, error)
	RefreshToken(ctx context.Context, realm string, refreshToken string, clientId string, clientSecret string) (*models.Tokens, error)
	AuthenticateClient(ctx context.Context, realm string, clientId string, clientSecret string) (*models.Tokens, error)
	ExchangeCodeForToken(ctx context.Context, realm string, redirectUri string, code string, clientId string, clientSecret string, codeVerifier string) (*models.Tokens, error)
	ExchangeUsernameAndPasswordForToken(ctx context.Context, realm string, username string, password string, clientId string, clientSecret string) (*models.Tokens, error)
	GetClaimsFromToken(ctx context.Context, Token string) (map[string]interface{}, error)
	ValidateServiceAccess(context.Context) error
	ExchangeExternalIdpToken(ctx context.Context, realm string, token string, clientId string, clientSecret string, idpAlias string) (*models.Tokens, error)
}
