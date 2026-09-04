/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"

	api "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	"golang.org/x/oauth2"
)

const passwordExchangeType = "password"

type LocalUser struct {
	Username string
	Password string
	client   api.TokensAPI
}

func NewLocalUser(username string, password string, client api.TokensAPI) *LocalUser {
	return &LocalUser{
		Username: username,
		Password: password,
		client:   client,
	}
}

func (lu *LocalUser) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	request := api.TokenRequest{
		Username:  api.PtrString(lu.Username),
		Password:  api.PtrString(lu.Password),
		GrantType: api.TokenRequestGrantTypeEnum(passwordExchangeType).Ptr(),
		ClientID:  api.PtrString(CLIClientID),
	}
	token, res, err := lu.client.GrantToken(ctx).TokenRequest(request).Execute()

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return nil, clierrors.HandleResponseError(res, err)
	}

	return &oauth2.Token{
		AccessToken:  token.GetAccessToken(),
		RefreshToken: token.GetRefreshToken(),
	}, nil
}
