/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type passwordGrantSource struct {
	ctx      context.Context
	tokens   idmclient.TokensAPI
	username string
	password string
}

func newTokenSource(ctx context.Context, tokens idmclient.TokensAPI, username, password string) oauth2.TokenSource {
	return oauth2.ReuseTokenSource(nil, &passwordGrantSource{
		ctx:      ctx,
		tokens:   tokens,
		username: username,
		password: password,
	})
}

func (s *passwordGrantSource) Token() (*oauth2.Token, error) {
	request := idmclient.TokenRequest{
		Username:  idmclient.PtrString(s.username),
		Password:  idmclient.PtrString(s.password),
		GrantType: idmclient.TOKENREQUESTGRANTTYPEENUM_PASSWORD.Ptr(),
	}
	response, _, err := s.tokens.GrantToken(s.ctx).TokenRequest(request).Execute()
	if err != nil {
		return nil, fmt.Errorf("control-plane login as %q: %w", s.username, err)
	}

	expiry, err := tokenExpiry(response.GetAccessToken())
	if err != nil {
		return nil, fmt.Errorf("control-plane login as %q: %w", s.username, err)
	}

	return &oauth2.Token{
		AccessToken:  response.GetAccessToken(),
		RefreshToken: response.GetRefreshToken(),
		Expiry:       expiry,
	}, nil
}

func tokenExpiry(accessToken string) (time.Time, error) {
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(accessToken, claims); err != nil {
		return time.Time{}, fmt.Errorf("parsing access token: %w", err)
	}

	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return time.Time{}, fmt.Errorf("reading token expiry: %w", err)
	}
	if expiration == nil {
		return time.Time{}, errors.New("access token has no expiry claim")
	}
	return expiration.Time, nil
}
