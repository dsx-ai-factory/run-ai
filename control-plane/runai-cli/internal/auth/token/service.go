/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"golang.org/x/oauth2"
)

const (
	DefaultTokenExpiryDelta = 10 * time.Second
)

type ReadWriter interface {
	ReadToken(path string) (*oauth2.Token, error)
	WriteToken(path string, token *oauth2.Token) error
}

type Refresher interface {
	RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error)
}
type Checker interface {
	Validate(token string) (bool, error)
}

type Service struct {
	token          *oauth2.Token
	authFile       string
	tokenValidator Checker
	tokenRefresher Refresher
	tokenWriter    ReadWriter
}

func NewTokenService(authFile string, tokenValidator Checker,
	tokenRefresher Refresher, tokenSaver ReadWriter) *Service {
	return &Service{
		authFile:       authFile,
		tokenValidator: tokenValidator,
		tokenRefresher: tokenRefresher,
		tokenWriter:    tokenSaver,
	}
}
func NewSimpleTokenServiceFromConfig(getter auth.ConfigGetter) (auth.TokenService, error) {
	authFile := getter.GetAuthenticationPath()
	baseUrl := getter.GetConfigValue(config.AuthenticationUrl)
	if authFile == "" || baseUrl == "" {
		return nil, fmt.Errorf("auth file or base URL is not configured")
	}
	return NewSimpleTokenService(authFile, baseUrl)

}

func NewSimpleTokenService(authFile, baseUrl string) (*Service, error) {
	readWriter := auth.NewFileBased64TokenReadWriter()
	client, err := auth.NewLoginClient(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create login client: %w", err)
	}
	tokenRefresher := NewTokenRefresher(client.TokensAPI)
	tokenValidator := NewTokenValidator(DefaultTokenExpiryDelta)
	return &Service{
		authFile:       authFile,
		tokenWriter:    readWriter,
		tokenRefresher: tokenRefresher,
		tokenValidator: tokenValidator,
	}, nil
}

func (s *Service) SetToken(token *oauth2.Token) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}
	s.token = token
	if err := s.tokenWriter.WriteToken(s.authFile, token); err != nil {
		return fmt.Errorf("failed to write token: %w", err)
	}
	return nil
}

func (s *Service) GetAccessToken(ctx context.Context) (string, error) {
	token, err := s.GetToken(ctx)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", fmt.Errorf("the provided token is nil.\n") //nolint:staticcheck // user facing message
	}
	return token.AccessToken, nil
}

func (s *Service) GetToken(ctx context.Context) (*oauth2.Token, error) {
	var err error
	if s.token == nil {
		s.token, err = s.getStoredToken()
		if err != nil {
			return nil, err
		}
	}

	if s.token.Valid() {
		return s.token, nil
	}
	valid, err := s.tokenValidator.Validate(s.token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("the provided token is invalid. %w \n", err) //nolint:staticcheck // user facing message
	}
	if !valid {
		return nil, fmt.Errorf("the token has expired")
	}
	token, err := s.handleRefreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed handle refresh token flow: %w", err)
	}
	return token, nil
}

func (s *Service) handleRefreshToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := s.tokenRefresher.RefreshToken(ctx, s.token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	s.token = token
	if err = s.tokenWriter.WriteToken(s.authFile, s.token); err != nil {
		return nil, errors.Join(auth.NewStoredTokenError("failed to write token"), err)
	}
	return s.token, nil
}

func (s *Service) getStoredToken() (*oauth2.Token, error) {
	token, err := s.tokenWriter.ReadToken(s.authFile)
	if err != nil {
		return nil, errors.Join(auth.NewStoredTokenError("failed to read stored token"), err)
	}
	return token, nil
}
