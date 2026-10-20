/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"golang.org/x/oauth2"
)

type Authenticator interface {
	Authenticate(ctx context.Context) (*oauth2.Token, error)
}

type AccessKeyAuthenticator struct {
	clientLogin  *openapi.APIClient
	clientName   string
	clientSecret string
}

func NewAccessKeyAuthenticator(authUrl string, clientName string, clientSecret string) (*AccessKeyAuthenticator, error) {
	clientLogin, err := NewLoginClient(authUrl)
	if err != nil {
		return nil, err
	}
	return &AccessKeyAuthenticator{
		clientLogin:  clientLogin,
		clientName:   clientName,
		clientSecret: clientSecret,
	}, nil
}

func (aa *AccessKeyAuthenticator) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	logger := config.GetInstance().GetLogger()
	logger.Debug(LoginWithAccessKeyCreds)
	applicationId := NewApplicationID(aa.clientName, aa.clientSecret, aa.clientLogin.TokensAPI)
	return applicationId.Authenticate(ctx)
}
