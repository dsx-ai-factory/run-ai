/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package identity_manager_client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	defaultRefreshBuffer = 30 * time.Second
	authHeaderKey        = "Authorization"
	bearerPrefix         = "Bearer "
)

type AuthnManager struct {
	idmClient      IIdmExternalClient
	clientId       string
	clientSecret   string
	accessToken    string
	tokenExpiresAt time.Time
	mu             sync.RWMutex
	refreshBuffer  time.Duration
}

type AuthnManagerConfig struct {
	ClientId      string
	ClientSecret  string
	IdmClient     IIdmExternalClient
	RefreshBuffer time.Duration
}

func NewAuthnManager(ctx context.Context, config AuthnManagerConfig) (*AuthnManager, error) {
	if config.RefreshBuffer == 0 {
		config.RefreshBuffer = defaultRefreshBuffer
	}
	authnManager := &AuthnManager{
		clientId:      config.ClientId,
		clientSecret:  config.ClientSecret,
		idmClient:     config.IdmClient,
		refreshBuffer: config.RefreshBuffer,
	}

	authnManager.mu.Lock()
	_, err := authnManager.setToken(ctx)
	authnManager.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("authn_manager: failed to initialize token: %w", err)
	}
	return authnManager, nil
}

func (mgr *AuthnManager) setToken(ctx context.Context) (string, error) {
	tokenResponse, err := mgr.idmClient.ClientCredentialsLoginV2(ctx, mgr.clientId, mgr.clientSecret)
	if err != nil {
		return "", err
	}
	mgr.accessToken = tokenResponse.GetAccessToken()
	mgr.tokenExpiresAt = time.Now().Add(time.Duration(tokenResponse.GetExpiresIn()) * time.Second)
	return mgr.accessToken, nil
}

func (mgr *AuthnManager) Token(ctx context.Context) (string, error) {
	mgr.mu.RLock()
	token := mgr.accessToken
	expiresAt := mgr.tokenExpiresAt
	buffer := mgr.refreshBuffer
	mgr.mu.RUnlock()

	if token != "" && !time.Now().After(expiresAt.Add(-buffer)) {
		return token, nil
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	if mgr.accessToken != "" && !time.Now().After(mgr.tokenExpiresAt.Add(-mgr.refreshBuffer)) {
		return mgr.accessToken, nil
	}
	return mgr.setToken(ctx)
}

func (mgr *AuthnManager) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := mgr.Token(req.Context())
	if err != nil {
		return nil, fmt.Errorf("authn_manager: failed to retrieve valid token: %w", err)
	}

	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set(authHeaderKey, bearerPrefix+token)
	resp, err := http.DefaultTransport.RoundTrip(clonedReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return mgr.retryWithFreshToken(req, resp)
	}
	return resp, nil
}

func (mgr *AuthnManager) retryWithFreshToken(req *http.Request, staleResp *http.Response) (*http.Response, error) {
	_, _ = io.Copy(io.Discard, staleResp.Body)
	staleResp.Body.Close()

	mgr.mu.Lock()
	token, err := mgr.setToken(req.Context())
	mgr.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("authn_manager: failed to refresh token: %w", err)
	}
	retryReq := req.Clone(req.Context())
	if req.Body != nil {
		if req.GetBody == nil {
			return nil, fmt.Errorf("authn_manager: cannot retry request with non-replayable body")
		}
		body, err := req.GetBody()
		if err != nil {
			return nil, fmt.Errorf("authn_manager: failed to replay request body: %w", err)
		}
		retryReq.Body = body
	}
	retryReq.Header.Set(authHeaderKey, bearerPrefix+token)
	return http.DefaultTransport.RoundTrip(retryReq)
}

func (mgr *AuthnManager) HttpClient() *http.Client {
	return &http.Client{Transport: mgr}
}
