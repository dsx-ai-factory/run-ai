/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type passwordGrantSource struct {
	ctx      context.Context
	http     *http.Client
	tokenURL string
	username string
	password string
}

type tokenRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grantType"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func newTokenSource(ctx context.Context, httpClient *http.Client, tokenURL, username, password string) oauth2.TokenSource {
	return oauth2.ReuseTokenSource(nil, &passwordGrantSource{
		ctx:      ctx,
		http:     httpClient,
		tokenURL: tokenURL,
		username: username,
		password: password,
	})
}

func (s *passwordGrantSource) Token() (*oauth2.Token, error) {
	token, err := s.login()
	if err != nil {
		return nil, fmt.Errorf("control-plane login as %q: %w", s.username, err)
	}
	return token, nil
}

func (s *passwordGrantSource) login() (*oauth2.Token, error) {
	body, err := json.Marshal(tokenRequest{Username: s.username, Password: s.password, GrantType: "password"})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.tokenURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := s.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s returned %s", s.tokenURL, response.Status)
	}

	var granted tokenResponse
	if err := json.NewDecoder(response.Body).Decode(&granted); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	expiry, err := tokenExpiry(granted.AccessToken)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  granted.AccessToken,
		RefreshToken: granted.RefreshToken,
		Expiry:       expiry,
	}, nil
}

func tokenIdentity(accessToken string) (string, error) {
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(accessToken, claims); err != nil {
		return "", fmt.Errorf("parsing access token: %w", err)
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("reading token subject: %w", err)
	}
	if subject == "" {
		return "", errors.New("access token has no subject claim")
	}
	return subject, nil
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
