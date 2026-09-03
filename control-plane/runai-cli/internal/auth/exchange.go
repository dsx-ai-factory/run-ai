/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"fmt"
	"io"
	"net/url"

	api "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"golang.org/x/oauth2"
)

const (
	GrantTypeExchangeToken       = "exchange_token"
	FailedToExchangeErrorWithMsg = "failed to exchange code for token: %s, error %w"
	FailedToExchangeError        = "failed to exchange code for token: %w"
	redirectQueryKey             = "redirect_uri"
)

// TokenApi is an interface that wraps the TokensApi from the identity-manager-client package for testing purposes
type TokenApi interface {
	api.TokensAPI
}

type ControlPlaneTokenExchanger struct {
	client       TokenApi
	codeVerifier string
}

func NewControlPlaneTokenExchanger(client api.TokensAPI, codeVerifier string) *ControlPlaneTokenExchanger {
	return &ControlPlaneTokenExchanger{client: client, codeVerifier: codeVerifier}
}

func (te *ControlPlaneTokenExchanger) Exchange(ctx context.Context, code string, authURL *url.URL) (*oauth2.Token, error) {
	requestBody := te.buildRequestToken(code, authURL)
	token, res, err := te.client.GrantToken(ctx).TokenRequest(requestBody).Execute()
	if err != nil {
		if res != nil {
			defer res.Body.Close()
			data, _ := io.ReadAll(res.Body)
			return nil, fmt.Errorf(FailedToExchangeErrorWithMsg, string(data), err)
		}
		return nil, fmt.Errorf(FailedToExchangeError, err)
	}
	claims, err := GetClaims(token.GetAccessToken())
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return &oauth2.Token{
		AccessToken:  token.GetAccessToken(),
		RefreshToken: token.GetRefreshToken(),
		Expiry:       claims.ExpiresAt.Time,
	}, nil
}

func (te *ControlPlaneTokenExchanger) buildRequestToken(code string, authURL *url.URL) api.TokenRequest {
	requestBody := api.TokenRequest{
		Code:      api.PtrString(code),
		GrantType: api.TokenRequestGrantTypeEnum(GrantTypeExchangeToken).Ptr(),
		ClientID:  api.PtrString(CLIClientID),
	}
	if te.codeVerifier != "" {
		requestBody.CodeVerifier = api.PtrString(te.codeVerifier)
	}
	redirectUrl := getRedirectUrl(authURL)
	if redirectUrl != "" {
		requestBody.RedirectUri = api.PtrString(redirectUrl)
	}
	return requestBody
}

func getRedirectUrl(authURL *url.URL) string {
	if authURL == nil {
		return ""
	}
	return authURL.Query().Get(redirectQueryKey)
}
