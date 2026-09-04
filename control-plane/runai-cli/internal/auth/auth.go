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

type ApplicationAuthenticator struct {
	clientLogin  *openapi.APIClient
	clientName   string
	clientSecret string
}

func NewApplicationAuthenticator(authUrl string, clientName string, clientSecret string) (*ApplicationAuthenticator, error) {
	clientLogin, err := NewLoginClient(authUrl)
	if err != nil {
		return nil, err
	}
	return &ApplicationAuthenticator{
		clientLogin:  clientLogin,
		clientName:   clientName,
		clientSecret: clientSecret,
	}, nil
}

func (aa *ApplicationAuthenticator) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	logger := config.GetInstance().GetLogger()
	logger.Debug(LoginWithApplicationCreds)
	applicationId := NewApplicationID(aa.clientName, aa.clientSecret, aa.clientLogin.TokensAPI)
	return applicationId.Authenticate(ctx)
}
