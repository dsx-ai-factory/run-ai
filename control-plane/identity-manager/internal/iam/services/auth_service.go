/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/auth"
)

type AuthService interface {
	GetUserInfo(ctx context.Context, realm string, accessToken string) (*models.UserInfo, error)
	RefreshToken(ctx context.Context, realm string, refreshToken string, clientId string, clientSecret string) (*models.Tokens, error)
	AuthenticateClient(ctx context.Context, realm string, clientId string, clientSecret string) (*models.Tokens, error)
	ExchangeCodeForToken(ctx context.Context, realm string, redirectUri string, code string, clientId string, clientSecret string, codeVerifier string) (*models.Tokens, error)
	GetClaimsFromToken(ctx context.Context, Token string) (map[string]interface{}, error)
	ExchangeUsernameAndPasswordForToken(ctx context.Context, realm string, username string, password string, clientId string, clientSecret string) (*models.Tokens, error)
	ValidateServiceAccess(context.Context) error
	ExchangeExternalIdpToken(ctx context.Context, realm string, token string, clientId string, clientSecret string, idpAlias string) (*models.Tokens, error)
}

type authService struct {
	authRepository auth.AuthRepository
}

func NewAuthService(authRepository auth.AuthRepository) AuthService {
	return &authService{
		authRepository: authRepository,
	}
}

func (a *authService) GetUserInfo(ctx context.Context, realm string, accessToken string) (*models.UserInfo, error) {
	return a.authRepository.GetUserInfo(ctx, realm, accessToken)
}

func (a *authService) RefreshToken(ctx context.Context, realm string, refreshToken string, clientId string, clientSecret string) (*models.Tokens, error) {
	return a.authRepository.RefreshToken(ctx, realm, refreshToken, clientId, clientSecret)
}

func (a *authService) AuthenticateClient(ctx context.Context, realm string, clientId string, clientSecret string) (*models.Tokens, error) {
	return a.authRepository.AuthenticateClient(ctx, realm, clientId, clientSecret)
}

func (a *authService) ExchangeCodeForToken(ctx context.Context, realm string, redirectUri string, code string, clientId string, clientSecret string, codeVerifier string) (*models.Tokens, error) {
	return a.authRepository.ExchangeCodeForToken(ctx, realm, redirectUri, code, clientId, clientSecret, codeVerifier)
}

func (a *authService) GetClaimsFromToken(ctx context.Context, Token string) (map[string]interface{}, error) {
	return a.authRepository.GetClaimsFromToken(ctx, Token)
}

func (a *authService) ExchangeUsernameAndPasswordForToken(ctx context.Context, realm string, username string, password string, clientId string, clientSecret string) (*models.Tokens, error) {
	return a.authRepository.ExchangeUsernameAndPasswordForToken(ctx, realm, username, password, clientId, clientSecret)
}

func (a *authService) ValidateServiceAccess(ctx context.Context) error {
	return a.authRepository.ValidateServiceAccess(ctx)
}

func (a *authService) ExchangeExternalIdpToken(ctx context.Context, realm string, token string, clientId string, clientSecret string, idpAlias string) (*models.Tokens, error) {
	return a.authRepository.ExchangeExternalIdpToken(ctx, realm, token, clientId, clientSecret, idpAlias)
}
